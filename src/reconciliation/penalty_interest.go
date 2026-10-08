package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

		var assessment Assessment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`"assessmentNo" = ?`, assessmentNumber).
			First(&assessment).Error; err != nil {
			return fmt.Errorf("assessment %s not found: %w", assessmentNumber, err)
		}

		var liabilityWrapper TaxLiabilityWrapper
		if assessment.TaxLiability != "" {
			_ = json.Unmarshal([]byte(assessment.TaxLiability), &liabilityWrapper)
		}
		if len(liabilityWrapper.TaxLiabilities) == 0 {
			return fmt.Errorf("assessment %s has no tax liabilities defined", assessmentNumber)
		}

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

		return &ReconciliationItemResult{
			AssessmentNumber: assessmentNumber,
			Status:           "FAILED",
			Message:          err.Error(),
		}, err
	}

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

func parseAmount(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

func RemovePenaltyAndInterest(taxLiabilityRaw string) (string, float64, float64, float64, error) {
	if strings.TrimSpace(taxLiabilityRaw) == "" {
		taxLiabilityRaw = "{}"
	}

	var root map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(taxLiabilityRaw))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to parse taxLiability json: %w", err)
	}

	var rawLiabilities []interface{}
	if list, ok := root["taxLiabilities"].([]interface{}); ok {
		rawLiabilities = list
	} else if list, ok := root["tax_liabilities"].([]interface{}); ok {
		rawLiabilities = list
	}

	filteredLiabilities := make([]interface{}, 0, len(rawLiabilities))
	var penaltyRemoved float64
	var interestRemoved float64
	var newTotal float64

	for _, it := range rawLiabilities {
		itemMap, ok := it.(map[string]interface{})
		if !ok {
			filteredLiabilities = append(filteredLiabilities, it)
			continue
		}

		taxCode := ""
		if tc, ok := itemMap["taxCode"].(string); ok {
			taxCode = tc
		} else if tc, ok := itemMap["tax_code"].(string); ok {
			taxCode = tc
		}
		code := strings.ToUpper(strings.TrimSpace(taxCode))
		amt := parseAmount(itemMap["amount"])

		if code == "PENALTY" {
			penaltyRemoved += amt
			continue
		} else if code == "INTEREST" {
			interestRemoved += amt
			continue
		}

		newTotal += amt
		filteredLiabilities = append(filteredLiabilities, itemMap)
	}

	penaltyRemoved = math.Round(penaltyRemoved*100) / 100
	interestRemoved = math.Round(interestRemoved*100) / 100
	totalRemoved := math.Round((penaltyRemoved+interestRemoved)*100) / 100
	newTotal = math.Round(newTotal*100) / 100

	root["taxLiabilities"] = filteredLiabilities
	root["totalTaxLiabilities"] = newTotal

	newBytes, err := json.Marshal(root)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to serialize updated taxLiability: %w", err)
	}

	return string(newBytes), penaltyRemoved, interestRemoved, totalRemoved, nil
}

func RemoveAssessmentPenaltyInterest(
	ctx context.Context,
	cfg *Config,
	assessmentNumber string,
	dryRun bool,
	userID, userName string,
) (*ReconciliationItemResult, error) {
	if strings.TrimSpace(assessmentNumber) == "" {
		return nil, fmt.Errorf("assessment number is required")
	}

	var assessment Assessment
	if err := cfg.AssessmentDB.WithContext(ctx).
		Where(`"assessmentNo" = ?`, assessmentNumber).
		First(&assessment).Error; err != nil {
		return nil, fmt.Errorf("assessment %s not found: %w", assessmentNumber, err)
	}

	newLiabilityJSON, pRemoved, iRemoved, totalRemoved, err := RemovePenaltyAndInterest(assessment.TaxLiability)
	if err != nil {
		return nil, fmt.Errorf("failed to process tax liabilities for %s: %w", assessmentNumber, err)
	}

	if !dryRun {
		updates := map[string]interface{}{
			"taxLiability": newLiabilityJSON,
			"sip":          true,
		}
		if err := cfg.AssessmentDB.WithContext(ctx).Model(&Assessment{}).
			Where(`"assessmentNo" = ?`, assessmentNumber).
			Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update assessment %s: %w", assessmentNumber, err)
		}
	}

	msg := fmt.Sprintf("Penalty (%.2f) and Interest (%.2f) removed successfully; total removed: %.2f (SIP=true)", pRemoved, iRemoved, totalRemoved)
	if dryRun {
		msg = fmt.Sprintf("[DRY RUN] Penalty (%.2f) and Interest (%.2f) would be removed; total: %.2f (SIP=true)", pRemoved, iRemoved, totalRemoved)
	}

	return &ReconciliationItemResult{
		AssessmentNumber: assessmentNumber,
		Status:           "SUCCESS",
		Message:          msg,
		ReconciledAmount: totalRemoved,
	}, nil
}

func RevertAssessmentPenaltyInterest(
	ctx context.Context,
	cfg *Config,
	assessmentNumber string,
	dryRun bool,
	userID, userName string,
) (*ReconciliationItemResult, error) {
	if strings.TrimSpace(assessmentNumber) == "" {
		return nil, fmt.Errorf("assessment number is required")
	}

	var assessment Assessment
	if err := cfg.AssessmentDB.WithContext(ctx).
		Where(`"assessmentNo" = ?`, assessmentNumber).
		First(&assessment).Error; err != nil {
		return nil, fmt.Errorf("assessment %s not found: %w", assessmentNumber, err)
	}

	if !dryRun {
		updates := map[string]interface{}{
			"sip": false,
		}
		if err := cfg.AssessmentDB.WithContext(ctx).Model(&Assessment{}).
			Where(`"assessmentNo" = ?`, assessmentNumber).
			Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to revert assessment %s: %w", assessmentNumber, err)
		}
	}

	msg := "Penalty and interest reverted successfully (SIP=false)"
	if dryRun {
		msg = "[DRY RUN] Penalty and interest would be reverted (SIP=false)"
	}

	return &ReconciliationItemResult{
		AssessmentNumber: assessmentNumber,
		Status:           "SUCCESS",
		Message:          msg,
		ReconciledAmount: 0.00,
	}, nil
}

func ProcessStaffPenaltyInterest(
	ctx context.Context,
	cfg *Config,
	assessmentNumber string,
	add bool,
	dryRun bool,
	userID, userName string,
) (*ReconciliationItemResult, error) {
	if add {
		return RemoveAssessmentPenaltyInterest(ctx, cfg, assessmentNumber, dryRun, userID, userName)
	}
	return RevertAssessmentPenaltyInterest(ctx, cfg, assessmentNumber, dryRun, userID, userName)
}
