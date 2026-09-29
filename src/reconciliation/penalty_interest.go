package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpdatePaymentPenaltyInterest updates an existing payment record and ledger with new penalty/interest charges.
func UpdatePaymentPenaltyInterest(
	ctx context.Context,
	cfg *Config,
	assessmentNumber string,
	userID, userName string,
) (*ReconciliationItemResult, error) {
	if strings.TrimSpace(assessmentNumber) == "" {
		return nil, fmt.Errorf("assessment number is required")
	}

	var receiptNumber string
	var newPaymentAmount float64
	var bankRef string
	var paymentRef string

	err := cfg.AssessmentDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Fetch & Lock Assessment
		var assessment Assessment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`"assessmentNo" = ?`, assessmentNumber).
			First(&assessment).Error; err != nil {
			return fmt.Errorf("assessment %s not found: %w", assessmentNumber, err)
		}

		// 2. Extract tax liabilities from assessment
		var liabilityWrapper TaxLiabilityWrapper
		if assessment.TaxLiability != "" {
			_ = json.Unmarshal([]byte(assessment.TaxLiability), &liabilityWrapper)
		}
		if len(liabilityWrapper.TaxLiabilities) == 0 {
			return fmt.Errorf("assessment %s has no tax liabilities defined", assessmentNumber)
		}

		// 3. Retrieve payment record from Payment DB
		var paymentRecord PaymentData
		if err := cfg.PaymentDB.WithContext(ctx).
			Where("assessment_number = ?", assessmentNumber).
			First(&paymentRecord).Error; err != nil {
			return fmt.Errorf("payment record not found in tbl_payment_data_db: %w", err)
		}

		receiptNumber = paymentRecord.ReceiptNumber
		bankRef = paymentRecord.PsspReferenceNumber
		paymentRef = paymentRecord.PaymentCode
		if paymentRef == "" {
			paymentRef = paymentRecord.PaymentReference
		}

		// 4. Merge liabilities
		originalPaidMap := make(map[string]float64)
		if paymentRecord.TaxPayable != "" {
			var origItems []map[string]interface{}
			_ = json.Unmarshal([]byte(paymentRecord.TaxPayable), &origItems)
			for _, it := range origItems {
				code, _ := it["tax_code"].(string)
				if code == "" {
					code, _ = it["taxCode"].(string)
				}
				code = strings.ToUpper(strings.TrimSpace(code))
				amt, _ := it["amount"].(float64)
				originalPaidMap[code] = amt
			}
		}

		mergedPaidMap := make(map[string]float64)
		for _, item := range liabilityWrapper.TaxLiabilities {
			code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
			origAmt := originalPaidMap[code]
			mergedPaidMap[code] = math.Max(origAmt, item.Amount)
		}
		for code, amt := range originalPaidMap {
			if _, exists := mergedPaidMap[code]; !exists {
				mergedPaidMap[code] = amt
			}
		}

		// Recompute total amount
		var totalAmount float64
		var taxPayableList []map[string]interface{}
		for code, amt := range mergedPaidMap {
			amtRounded := math.Round(amt*100) / 100
			totalAmount += amtRounded
			taxPayableList = append(taxPayableList, map[string]interface{}{
				"tax_code": code,
				"amount":   amtRounded,
			})
		}
		totalAmount = math.Round(totalAmount*100) / 100
		newPaymentAmount = totalAmount

		taxPayableJSON, _ := json.Marshal(taxPayableList)

		// Recalculate signature
		newSig, _ := GeneratePaymentSignature(
			paymentRecord.ReceiptNumber,
			paymentRecord.AssessmentNumber,
			newPaymentAmount,
			paymentRecord.PaymentReference,
			paymentRecord.PayerTIN,
			paymentRecord.TaxYear,
			paymentRecord.Currency,
			paymentRecord.OfficeID,
			cfg.PaymentSigningKey,
		)

		// Update PaymentData record in Payment DB
		cfg.PaymentDB.Model(&paymentRecord).Updates(map[string]interface{}{
			"amount":      newPaymentAmount,
			"tax_payable": string(taxPayableJSON),
			"signature":   newSig,
		})

		// 5. Update PaymentBreakdown in Payment DB
		var breakdown PaymentBreakdown
		if err := cfg.PaymentDB.Where("assessment_number = ?", assessmentNumber).
			Order("id DESC").First(&breakdown).Error; err == nil {
			cfg.PaymentDB.Model(&breakdown).Updates(map[string]interface{}{
				"tax_types":        string(taxPayableJSON),
				"amount":           newPaymentAmount,
				"total_amount":     newPaymentAmount,
				"remaining_amount": 0.00,
				"status":           "completed",
			})
		}

		// 6. Insert into Tax Ledger and Collections
		assessmentHash := HashAssessmentNo(assessment.AssessmentNo)
		collPaymentDate := paymentRecord.PaymentDate

		for code, amt := range mergedPaidMap {
			if amt <= 0 {
				continue
			}
			amtRounded := math.Round(amt*100) / 100

			// Idempotent TaxLedger DR
			if !hasLedgerEntry(tx, assessmentHash, code, "DR", "") {
				_ = createLedgerEntry(tx, &assessment, code, amtRounded, "DR", cfg.LedgerEncryptionKey)
			}
			// Idempotent TaxLedger CR
			if !hasLedgerEntry(tx, assessmentHash, code, "CR", "") {
				_ = createLedgerEntry(tx, &assessment, code, amtRounded, "CR", cfg.LedgerEncryptionKey)
			}

			// Idempotent Collection
			var collCount int64
			tx.Table("collections").
				Where("assessmentNo = ? AND paymentBankRef = ? AND tax = ?",
					assessment.AssessmentNo, paymentRecord.PsspReferenceNumber, code).
				Count(&collCount)

			if collCount == 0 {
				coll := Collection{
					Tax:              code,
					OfficeID:         assessment.OfficeID,
					StateID:          assessment.StateID,
					Description:      fmt.Sprintf("%s Payment", code),
					Currency:         strings.ToUpper(strings.TrimSpace(assessment.Currency)),
					Amount:           amtRounded,
					AssessmentPeriod: assessment.AssessmentPeriod,
					AssessmentNo:     assessment.AssessmentNo,
					TaxID:            assessment.TIN,
					PaymentReference: paymentRef,
					TransactionID:    paymentRecord.ReceiptNumber,
					PaymentBankRef:   paymentRecord.PsspReferenceNumber,
					PaymentGate:      paymentRecord.Vendor,
					PaymentDate:      collPaymentDate,
					CreatedAt:        time.Now().UTC(),
					UpdatedAt:        time.Now().UTC(),
				}
				tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&coll)
			}
		}

		// 7. Clear remaining liability and mark as fully PAID
		var updatedLiabilities []TaxLiabilityItem
		for code := range mergedPaidMap {
			name := TaxTypeNameMap[code]
			if name == "" {
				name = code
			}
			updatedLiabilities = append(updatedLiabilities, TaxLiabilityItem{
				TaxID:       "",
				TaxCode:     code,
				TaxCodeName: name,
				Amount:      0.0,
			})
		}
		newTaxLiabJSON, _ := json.Marshal(TaxLiabilityWrapper{
			TaxLiabilities:      updatedLiabilities,
			TotalTaxLiabilities: 0.0,
		})

		tx.Model(&assessment).Updates(map[string]interface{}{
			"taxLiability":     string(newTaxLiabJSON),
			"paymentStatus":    "PAID",
			"filingStatus":     "FILED",
			"paymentRefStatus": "USED",
		})

		return nil
	})

	if err != nil {
		cfg.PaymentDB.Create(&LegacyReconciliationLog{
			AssessmentNumber: assessmentNumber,
			PaymentReference: paymentRef,
			BankReference:    bankRef,
			UserID:           userID,
			UserName:         userName,
			Status:           "FAILED",
			Message:          err.Error(),
			Amount:           0.00,
			CreatedAt:        time.Now().UTC(),
		})
		return &ReconciliationItemResult{
			AssessmentNumber: assessmentNumber,
			Status:           "FAILED",
			Message:          err.Error(),
		}, err
	}

	// Log SUCCESS
	cfg.PaymentDB.Create(&LegacyReconciliationLog{
		AssessmentNumber: assessmentNumber,
		PaymentReference: paymentRef,
		BankReference:    bankRef,
		UserID:           userID,
		UserName:         userName,
		Status:           "SUCCESS",
		Message:          "payment penalty and interest update completed successfully",
		Amount:           newPaymentAmount,
		CreatedAt:        time.Now().UTC(),
	})

	return &ReconciliationItemResult{
		AssessmentNumber: assessmentNumber,
		PaymentReference: paymentRef,
		BankReference:    bankRef,
		ReceiptNumber:    receiptNumber,
		ReconciledAmount: newPaymentAmount,
		Status:           "SUCCESS",
		Message:          "Payment penalty and interest updated successfully",
	}, nil
}
