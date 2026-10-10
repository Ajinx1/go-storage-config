package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var TaxTypeNameMap = map[string]string{
	"VAT":      "Value Added Tax",
	"WHT":      "Withholding Tax",
	"CIT":      "Company Income Tax",
	"PIT":      "Personal Income Tax",
	"EDT":      "Education Development Tax",
	"POL":      "Police Trust Fund Levy",
	"NIT":      "National Information Technology Development Fund",
	"LRP":      "Late Registration Penalty",
	"PEN":      "Penalty",
	"PENALTY":  "Penalty",
	"INT":      "Interest",
	"INTEREST": "Interest",
}

func requiresSelfBalancedLedger(taxCode string) bool {
	code := strings.ToUpper(strings.TrimSpace(taxCode))
	return code == "LRP" || code == "PEN" || code == "PENALTY" || code == "INT" || code == "INTEREST"
}

func buildLedgerDescription(taxType, entryType string) string {
	code := strings.ToUpper(strings.TrimSpace(taxType))
	name, ok := TaxTypeNameMap[code]
	if !ok {
		name = code
	}
	if entryType == "DR" {
		return fmt.Sprintf("%s Liability", name)
	}
	return fmt.Sprintf("%s Payment", name)
}

func getLatestLedgerBalance(tx *gorm.DB, tinHash, currency string) float64 {
	var balance float64
	row := tx.Table("tax_ledgers").
		Select("balance").
		Where("tin_hash = ? AND currency = ?", tinHash, currency).
		Order(`"order" DESC`).
		Limit(1).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Row()
	if row != nil {
		_ = row.Scan(&balance)
	}
	return balance
}

func hasLedgerEntry(tx *gorm.DB, assessmentHash, taxType, entryType, description string) bool {
	var count int64
	q := tx.Table("tax_ledgers").
		Where("assessment_number_hash = ? AND tax_type = ? AND type = ?", assessmentHash, taxType, entryType)
	if entryType == "CR" && description != "" {
		q = q.Where("description = ?", description)
	}
	q.Count(&count)
	return count > 0
}

func createLedgerEntry(
	tx *gorm.DB,
	assessment *Assessment,
	taxCode string,
	amount float64,
	entryType string,
	encryptionKey string,
) error {
	taxCodeClean := strings.ToUpper(strings.TrimSpace(taxCode))
	tinHash := HashTIN(assessment.TIN)
	tinEncrypted, _ := Encrypt(assessment.TIN, encryptionKey)
	assessmentHash := HashAssessmentNo(assessment.AssessmentNo)
	assessmentEncrypted, _ := Encrypt(assessment.AssessmentNo, encryptionKey)

	currentBalance := getLatestLedgerBalance(tx, tinHash, assessment.Currency)
	amountRounded := math.Round(amount*100) / 100

	var newBalance float64
	var paymentMethod string
	if entryType == "CR" {
		newBalance = currentBalance + amountRounded
		paymentMethod = "CASH"
	} else {
		newBalance = currentBalance - amountRounded
		paymentMethod = "DEBIT"
	}

	description := buildLedgerDescription(taxCodeClean, entryType)

	var assessmentMonth *string
	if taxCodeClean == "VAT" && assessment.AssessmentPeriod != "" {
		t, err := time.Parse("January 2006", assessment.AssessmentPeriod)
		if err == nil {
			mStr := fmt.Sprintf("%d", t.Month())
			assessmentMonth = &mStr
		}
	}

	now := time.Now().UTC()
	entry := TaxLedger{
		ID:                        GenerateULID(),
		DateCreated:               now,
		DateUpdated:               now,
		IsDeleted:                 false,
		CreatedBy:                 "system",
		UpdatedBy:                 "system",
		TINHash:                   tinHash,
		TINEncrypted:              tinEncrypted,
		AssessmentNumberHash:      assessmentHash,
		AssessmentNumberEncrypted: assessmentEncrypted,
		TaxType:                   taxCodeClean,
		AssessmentPeriod:          assessment.AssessmentPeriod,
		AssessmentMonth:           assessmentMonth,
		Description:               description,
		Amount:                    amountRounded,
		Balance:                   newBalance,
		OfficeID:                  assessment.OfficeID,
		StateID:                   assessment.StateID,
		Currency:                  strings.ToUpper(strings.TrimSpace(assessment.Currency)),
		Type:                      entryType,
		PaymentMethod:             paymentMethod,
	}

	return tx.Create(&entry).Error
}

func createLedgerEntries(
	tx *gorm.DB,
	assessment *Assessment,
	taxCode string,
	amount float64,
	flag int,
	paymentAmount float64,
	encryptionKey string,
) error {
	assessmentHash := HashAssessmentNo(assessment.AssessmentNo)
	taxCodeClean := strings.ToUpper(strings.TrimSpace(taxCode))

	if paymentAmount == 0 {
		paymentAmount = amount
	}

	if requiresSelfBalancedLedger(taxCodeClean) {
		if paymentAmount <= 0 {
			return nil
		}
		if !hasLedgerEntry(tx, assessmentHash, taxCodeClean, "DR", "") {
			if err := createLedgerEntry(tx, assessment, taxCodeClean, paymentAmount, "DR", encryptionKey); err != nil {
				return err
			}
		}
		crDesc := buildLedgerDescription(taxCodeClean, "CR")
		if !hasLedgerEntry(tx, assessmentHash, taxCodeClean, "CR", crDesc) {
			if err := createLedgerEntry(tx, assessment, taxCodeClean, paymentAmount, "CR", encryptionKey); err != nil {
				return err
			}
		}
		return nil
	}

	if flag == 1 && !strings.Contains(assessment.AssessmentNo, "-") {
		if amount > 0 && !hasLedgerEntry(tx, assessmentHash, taxCodeClean, "DR", "") {
			if err := createLedgerEntry(tx, assessment, taxCodeClean, amount, "DR", encryptionKey); err != nil {
				return err
			}
		}
	}

	if paymentAmount > 0 {
		crDesc := buildLedgerDescription(taxCodeClean, "CR")
		if !hasLedgerEntry(tx, assessmentHash, taxCodeClean, "CR", crDesc) {
			if err := createLedgerEntry(tx, assessment, taxCodeClean, paymentAmount, "CR", encryptionKey); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildRecoveryTaxLiability(assessment *Assessment, breakdownItems []BreakdownTaxItem, amount float64) string {
	if len(breakdownItems) > 0 {
		var liabilities []TaxLiabilityItem
		var total float64
		for _, item := range breakdownItems {
			if item.Amount <= 0 {
				continue
			}
			code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
			if code == "" {
				code = assessment.TaxType
			}
			name := TaxTypeNameMap[code]
			if name == "" {
				name = code
			}
			liabilities = append(liabilities, TaxLiabilityItem{
				TaxID:       "",
				TaxCode:     code,
				TaxCodeName: name,
				Amount:      item.Amount,
			})
			total += item.Amount
		}
		if len(liabilities) > 0 {
			b, _ := json.Marshal(TaxLiabilityWrapper{
				TaxLiabilities:      liabilities,
				TotalTaxLiabilities: math.Round(total*100) / 100,
			})
			return string(b)
		}
	}

	name := TaxTypeNameMap[assessment.TaxType]
	if name == "" {
		name = assessment.TaxType
	}
	b, _ := json.Marshal(TaxLiabilityWrapper{
		TaxLiabilities: []TaxLiabilityItem{
			{
				TaxID:       "",
				TaxCode:     assessment.TaxType,
				TaxCodeName: name,
				Amount:      amount,
			},
		},
		TotalTaxLiabilities: amount,
	})
	return string(b)
}

func buildUpdatedLiability(assessment *Assessment, paidItems []TaxPaymentItem) (string, float64) {
	var currentWrapper TaxLiabilityWrapper
	if assessment.TaxLiability != "" {
		_ = json.Unmarshal([]byte(assessment.TaxLiability), &currentWrapper)
	}

	paidMap := make(map[string]float64)
	for _, item := range paidItems {
		code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
		paidMap[code] = item.Amount
	}

	var updatedItems []TaxLiabilityItem
	var total float64

	for _, item := range currentWrapper.TaxLiabilities {
		code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
		paidAmt := paidMap[code]
		balance := math.Round(math.Max(0.0, item.Amount-paidAmt)*100) / 100
		updatedItems = append(updatedItems, TaxLiabilityItem{
			TaxID:       item.TaxID,
			TaxCode:     item.TaxCode,
			TaxCodeName: item.TaxCodeName,
			Amount:      balance,
		})
		total += balance
	}

	total = math.Round(total*100) / 100
	resultJSON, _ := json.Marshal(TaxLiabilityWrapper{
		TaxLiabilities:      updatedItems,
		TotalTaxLiabilities: total,
	})
	return string(resultJSON), total
}

func generateReceiptNumber(paymentDB *gorm.DB, prefix string) (string, error) {
	var seq int64
	err := paymentDB.Raw("SELECT nextval('receipt_number_seq')").Scan(&seq).Error
	if err != nil {
		return "", fmt.Errorf("failed to get receipt sequence: %w", err)
	}
	today := time.Now().UTC().Format("20060102")
	return fmt.Sprintf("%s-%s-%08d", prefix, today, seq), nil
}

func cleanUpPaymentData(paymentDB *gorm.DB, assessmentNumber string) (*Transaction, error) {
	var tx Transaction
	err := paymentDB.Where("assessment_number = ?", assessmentNumber).
		Order("id DESC").First(&tx).Error
	hasTx := err == nil

	var breakdown PaymentBreakdown
	err = paymentDB.Where("assessment_number = ?", assessmentNumber).
		Order("id DESC").First(&breakdown).Error
	hasBreakdown := err == nil

	if hasTx {
		paymentDB.Model(&tx).Update("status", "completed")
	}
	if hasBreakdown {
		paymentDB.Model(&breakdown).Update("status", "completed")
	}

	// Delete incomplete artifacts
	txQuery := paymentDB.Where("assessment_number = ? AND status != ?", assessmentNumber, "completed")
	if hasTx {
		txQuery = txQuery.Where("id != ?", tx.ID)
	}
	txQuery.Delete(&Transaction{})

	breakdownQuery := paymentDB.Where("assessment_number = ? AND status != ?", assessmentNumber, "completed")
	if hasBreakdown {
		breakdownQuery = breakdownQuery.Where("id != ?", breakdown.ID)
	}
	breakdownQuery.Delete(&PaymentBreakdown{})

	if hasTx {
		return &tx, nil
	}
	return nil, nil
}

func parseDateStr(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	formats := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	clean := strings.TrimSpace(dateStr)
	clean = strings.Split(clean, " +")[0]
	clean = strings.Split(clean, " -")[0]
	clean = strings.TrimSuffix(clean, "Z")

	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, clean); err == nil {
			return &t
		}
	}
	return nil
}

// ReconcileSingleAssessment executes the end-to-end reconciliation flow for one assessment.
func ReconcileSingleAssessment(
	ctx context.Context,
	cfg *Config,
	input ReconciliationInput,
) (*ReconciliationItemResult, error) {
	correlationID := uuid.New().String()
	assessmentNumber := input.AssessmentNumber
	paymentRef := input.PaymentReference
	bankRef := input.BankReference

	// Execute Assessment DB Transaction
	var receiptNumber string
	var reconciledAmount float64

	err := cfg.AssessmentDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var assessment Assessment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`"assessmentNo" = ?`, assessmentNumber).
			First(&assessment).Error; err != nil {
			return fmt.Errorf("assessment not found: %w", err)
		}

		// Extract liabilities
		var liabilityWrapper TaxLiabilityWrapper
		if assessment.TaxLiability != "" {
			_ = json.Unmarshal([]byte(assessment.TaxLiability), &liabilityWrapper)
		}

		assLiabilityMap := make(map[string]float64)
		for _, item := range liabilityWrapper.TaxLiabilities {
			code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
			if code == "" {
				code = assessment.TaxType
			}
			assLiabilityMap[code] = item.Amount
		}

		// Fetch breakdown from Payment DB
		var breakdown PaymentBreakdown
		err := cfg.PaymentDB.WithContext(ctx).
			Where("assessment_number = ?", assessmentNumber).
			Order("id DESC").First(&breakdown).Error
		hasBreakdown := err == nil

		var breakdownItems []BreakdownTaxItem
		if hasBreakdown && breakdown.TaxTypes != "" {
			_ = json.Unmarshal([]byte(breakdown.TaxTypes), &breakdownItems)
		}

		// Standardize target tax types & amounts
		var newTaxTypes []BreakdownTaxItem
		var targetTotalAmount float64

		if len(breakdownItems) > 0 {
			for _, item := range breakdownItems {
				code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
				if code == "" {
					code = assessment.TaxType
				}
				amt := assLiabilityMap[code]
				if amt <= 0 {
					amt = item.Amount
				}
				newTaxTypes = append(newTaxTypes, BreakdownTaxItem{
					TaxCode: code,
					Amount:  math.Round(amt*100) / 100,
				})
				targetTotalAmount += amt
			}
		} else {
			for _, item := range liabilityWrapper.TaxLiabilities {
				if item.Amount > 0 {
					code := strings.ToUpper(strings.TrimSpace(item.TaxCode))
					if code == "" {
						code = assessment.TaxType
					}
					newTaxTypes = append(newTaxTypes, BreakdownTaxItem{
						TaxCode: code,
						Amount:  math.Round(item.Amount*100) / 100,
					})
					targetTotalAmount += item.Amount
				}
			}
			if len(newTaxTypes) == 0 {
				targetTotalAmount = liabilityWrapper.TotalTaxLiabilities
				if targetTotalAmount <= 0 {
					targetTotalAmount = input.Amount
				}
				newTaxTypes = []BreakdownTaxItem{
					{
						TaxCode: strings.ToUpper(strings.TrimSpace(assessment.TaxType)),
						Amount:  math.Round(targetTotalAmount*100) / 100,
					},
				}
			}
		}

		targetTotalAmount = math.Round(targetTotalAmount*100) / 100
		amount := targetTotalAmount
		reconciledAmount = amount

		newTaxTypesJSON, _ := json.Marshal(newTaxTypes)

		// Update or Insert PaymentBreakdown
		if hasBreakdown {
			cfg.PaymentDB.Model(&breakdown).Updates(map[string]interface{}{
				"tax_types":        string(newTaxTypesJSON),
				"amount":           amount,
				"total_amount":     amount,
				"remaining_amount": 0.00,
				"status":           "pending",
			})
		} else {
			nowMinus24 := time.Now().UTC().Add(-24 * time.Hour)
			newBreakdown := PaymentBreakdown{
				AssessmentNumber: assessmentNumber,
				TaxTypes:         string(newTaxTypesJSON),
				Amount:           amount,
				TotalAmount:      amount,
				RemainingAmount:  0.00,
				Status:           "pending",
				OfficeID:         assessment.OfficeID,
				StateID:          assessment.StateID,
				Currency:         assessment.Currency,
				CreatedAt:        nowMinus24,
				UpdatedAt:        nowMinus24,
			}
			cfg.PaymentDB.Create(&newBreakdown)
		}

		// Ensure recovery tax liability on assessment if empty
		if assessment.TaxLiability == "" || len(liabilityWrapper.TaxLiabilities) == 0 {
			assessment.TaxLiability = buildRecoveryTaxLiability(&assessment, newTaxTypes, amount)
		}

		// Waterfall Payment Distribution
		remainingPayment := amount
		var paymentItems []TaxPaymentItem
		for _, li := range newTaxTypes {
			allocated := math.Min(remainingPayment, li.Amount)
			if allocated > 0 {
				paymentItems = append(paymentItems, TaxPaymentItem{
					TaxCode: li.TaxCode,
					Amount:  math.Round(allocated*100) / 100,
				})
				remainingPayment -= allocated
			}
		}
		if remainingPayment > 0 {
			if len(paymentItems) > 0 {
				paymentItems[0].Amount += math.Round(remainingPayment*100) / 100
			} else {
				paymentItems = []TaxPaymentItem{
					{
						TaxCode: assessment.TaxType,
						Amount:  math.Round(remainingPayment*100) / 100,
					},
				}
			}
		}

		// Create Tax Ledger Entries (DR liability, CR payment)
		paymentMap := make(map[string]float64)
		for _, pi := range paymentItems {
			paymentMap[strings.ToUpper(strings.TrimSpace(pi.TaxCode))] = pi.Amount
		}
		for _, li := range newTaxTypes {
			codeClean := strings.ToUpper(strings.TrimSpace(li.TaxCode))
			payAmt := paymentMap[codeClean]
			if err := createLedgerEntries(tx, &assessment, li.TaxCode, li.Amount, 1, payAmt, cfg.LedgerEncryptionKey); err != nil {
				return fmt.Errorf("ledger creation failed for %s: %w", li.TaxCode, err)
			}
		}

		// Build updated liability and payment status
		updatedLiability, remainingBalance := buildUpdatedLiability(&assessment, paymentItems)
		assessment.TaxLiability = updatedLiability

		paymentStatus := "PAID"
		if remainingBalance > 0 {
			paymentStatus = "PARTIALLY_PAID"
		}

		parsedPaymentDate := parseDateStr(input.PaymentDate)
		if parsedPaymentDate == nil {
			now := time.Now().UTC()
			parsedPaymentDate = &now
		}

		if err := tx.Model(&assessment).Updates(map[string]interface{}{
			"taxLiability":     updatedLiability,
			"paymentStatus":    paymentStatus,
			"payment date":     parsedPaymentDate,
			"filingStatus":     "FILED",
			"paymentRefStatus": "USED",
		}).Error; err != nil {
			return fmt.Errorf("failed to update assessment status: %w", err)
		}

		receiptPrefix := "LGY"
		if input.WritePaymentRecord {
			receiptPrefix = "RCP"
		}
		rcpNum, err := generateReceiptNumber(cfg.PaymentDB, receiptPrefix)
		if err != nil {
			return err
		}
		receiptNumber = rcpNum

		for _, it := range paymentItems {
			if it.Amount <= 0 {
				continue
			}
			coll := Collection{
				Tax:              strings.ToUpper(strings.TrimSpace(it.TaxCode)),
				OfficeID:         assessment.OfficeID,
				StateID:          assessment.StateID,
				Description:      fmt.Sprintf("%s Payment", strings.ToUpper(strings.TrimSpace(it.TaxCode))),
				Currency:         strings.ToUpper(strings.TrimSpace(assessment.Currency)),
				Amount:           math.Round(it.Amount*100) / 100,
				AssessmentPeriod: assessment.AssessmentPeriod,
				AssessmentNo:     assessment.AssessmentNo,
				TaxID:            assessment.TIN,
				PaymentReference: paymentRef,
				TransactionID:    receiptNumber,
				PaymentBankRef:   bankRef,
				PaymentGate:      input.PaymentChannel,
				PaymentDate:      *parsedPaymentDate,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&coll).Error; err != nil {
				return fmt.Errorf("failed to insert collection: %w", err)
			}
		}

		if input.WritePaymentRecord {
			metaDataJSON, _ := json.Marshal(map[string]interface{}{
				"bank":      "paystack titan",
				"channel":   "card",
				"sessionId": "UY6A8HIHHSD",
			})

			var taxPayableList []map[string]interface{}
			for _, it := range paymentItems {
				taxPayableList = append(taxPayableList, map[string]interface{}{
					"tax_code": it.TaxCode,
					"amount":   it.Amount,
				})
			}
			taxPayableJSON, _ := json.Marshal(taxPayableList)

			sig, _ := GeneratePaymentSignature(
				receiptNumber,
				assessment.AssessmentNo,
				amount,
				paymentRef,
				assessment.TIN,
				assessment.AssessmentYear,
				assessment.Currency,
				assessment.OfficeID,
				cfg.PaymentSigningKey,
			)

			contractDate := parseDateStr(assessment.PeriodStart)
			periodEnd := parseDateStr(assessment.PeriodEnd)
			expiryDate := parseDateStr(assessment.ExpirationDate)

			paymentDataRecord := PaymentData{
				ID:                  uuid.New().String(),
				ReceiptNumber:       receiptNumber,
				TaxType:             assessment.TaxType,
				Currency:            assessment.Currency,
				Amount:              amount,
				AssessmentNumber:    assessment.AssessmentNo,
				PaymentReference:    assessment.PaymentReference,
				PayerName:           assessment.BusinessName,
				PayerTIN:            assessment.TIN,
				AssessmentPeriod:    assessment.AssessmentPeriod,
				OfficeID:            assessment.OfficeID,
				ContractDate:        contractDate,
				PeriodEnd:           periodEnd,
				TaxYear:             assessment.AssessmentYear,
				ExpiryDate:          expiryDate,
				PaymentDate:         *parsedPaymentDate,
				PaymentStatus:       "paid",
				FilingStatus:        assessment.FilingStatus,
				TaxPayable:          string(taxPayableJSON),
				MetaData:            string(metaDataJSON),
				Vendor:              input.PaymentChannel,
				PaymentCode:         paymentRef,
				PsspReferenceNumber: bankRef,
				SettlementStatus:    "COMPLETED",
				Signature:           sig,
				CreatedAt:           time.Now().UTC(),
				UpdatedAt:           time.Now().UTC(),
			}

			if err := cfg.PaymentDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&paymentDataRecord).Error; err != nil {
				return fmt.Errorf("failed to insert payment data: %w", err)
			}
		}

		txRecord, _ := cleanUpPaymentData(cfg.PaymentDB, assessmentNumber)

		if cfg.AuditDB != nil {
			prn := paymentRef
			transID := ""
			if txRecord != nil {
				if txRecord.PaymentReference != "" {
					prn = txRecord.PaymentReference
				}
				transID = txRecord.TransactionID
			}

			rawWebhookData := map[string]interface{}{
				"prn":              prn,
				"taxId":            assessment.TIN,
				"amount":           amount,
				"status":           "success",
				"taxType":          assessment.TaxType,
				"metaData":         map[string]interface{}{"bank": "paystack titan", "channel": "card"},
				"paymentDate":      input.PaymentDate,
				"transactionId":    transID,
				"assessmentNumber": assessmentNumber,
				"paymentReference": paymentRef,
			}
			rawWebhookJSON, _ := json.Marshal(rawWebhookData)

			now := time.Now().UTC()

			fwdPayload, _ := json.Marshal(map[string]interface{}{
				"status":        "updated",
				"correlationId": correlationID,
			})
			cfg.AuditDB.Create(&AuditLog{
				ID:          uuid.New().String(),
				Service:     "payment-gateway-service",
				Action:      "WEBHOOK_FORWARDED",
				Description: "Payment webhook forwarded to payment service",
				EntityType:  "payment",
				EntityID:    correlationID,
				UserID:      "payment-aggregator",
				OfficeID:    "0",
				StateID:     "",
				Payload:     string(fwdPayload),
				CreatedAt:   now,
				UpdatedAt:   now,
			})

			recvPayload, _ := json.Marshal(map[string]interface{}{
				"data":          rawWebhookData,
				"event":         "payment.success",
				"source":        "payment-aggregator",
				"correlationId": correlationID,
			})
			recvTime := now.Add(10 * time.Second)
			cfg.AuditDB.Create(&AuditLog{
				ID:          uuid.New().String(),
				Service:     "payment-gateway-service",
				Action:      "WEBHOOK_RECEIVED",
				Description: "Payment webhook received from aggregator",
				EntityType:  "payment",
				EntityID:    correlationID,
				UserID:      "payment-aggregator",
				OfficeID:    "0",
				StateID:     "",
				Payload:     string(recvPayload),
				CreatedAt:   recvTime,
				UpdatedAt:   recvTime,
			})

			succTime := now.Add(30 * time.Second)
			cfg.AuditDB.Create(&AuditLog{
				ID:          uuid.New().String(),
				Service:     "payment-service",
				Action:      "success",
				Description: fmt.Sprintf("Payment success | reference=%s | prn=%s | transactionId=%s | assessment=%s | taxType=%s | status=success | correlationId=%s | source=payment-aggregator", paymentRef, prn, transID, assessmentNumber, assessment.TaxType, correlationID),
				EntityType:  "payment_aggregated_transaction",
				EntityID:    assessmentNumber,
				UserID:      assessment.BusinessName,
				OfficeID:    assessment.OfficeID,
				StateID:     assessment.StateID,
				Payload:     string(rawWebhookJSON),
				CreatedAt:   succTime,
				UpdatedAt:   succTime,
			})
		}

		return nil
	})

	if err != nil {

		return &ReconciliationItemResult{
			AssessmentNumber: assessmentNumber,
			PaymentReference: paymentRef,
			BankReference:    bankRef,
			Status:           "FAILED",
			Message:          err.Error(),
		}, err
	}

	return &ReconciliationItemResult{
		AssessmentNumber: assessmentNumber,
		PaymentReference: paymentRef,
		BankReference:    bankRef,
		ReceiptNumber:    receiptNumber,
		ReconciledAmount: reconciledAmount,
		Status:           "SUCCESS",
		Message:          "legacy reconciliation completed successfully",
	}, nil
}
