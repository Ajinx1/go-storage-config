package reconciliation

import (
	"context"
	"fmt"
	"math"
	"strings"
)

func ProcessStateTaxLedger(
	ctx context.Context,
	cfg *Config,
	assessmentNumber string,
	taxTypes []string,
	add bool,
	dryRun bool,
) (*ReconciliationItemResult, error) {
	asmClean := strings.TrimSpace(assessmentNumber)
	if asmClean == "" {
		return nil, fmt.Errorf("assessment number is required")
	}

	asmHash := HashAssessmentNo(asmClean)

	var cleanTaxTypes []string
	for _, tt := range taxTypes {
		cleaned := strings.ToUpper(strings.TrimSpace(tt))
		if cleaned != "" {
			cleanTaxTypes = append(cleanTaxTypes, cleaned)
		}
	}

	query := cfg.AssessmentDB.WithContext(ctx).
		Model(&TaxLedger{}).
		Where("assessment_number_hash = ? AND type = 'DR'", asmHash)
	if len(cleanTaxTypes) > 0 {
		query = query.Where("UPPER(tax_type) IN (?)", cleanTaxTypes)
	}

	var ledgers []TaxLedger
	if err := query.Find(&ledgers).Error; err != nil {
		return nil, fmt.Errorf("failed to query tax_ledgers for %s: %w", asmClean, err)
	}

	if len(ledgers) == 0 {
		if len(cleanTaxTypes) > 0 {
			return nil, fmt.Errorf("no DR tax ledger entries found for assessment %s with tax types [%s]", asmClean, strings.Join(cleanTaxTypes, ", "))
		}
		return nil, fmt.Errorf("no DR tax ledger entries found for assessment %s", asmClean)
	}

	var totalAmount float64
	for _, l := range ledgers {
		totalAmount += l.Amount
	}
	totalAmount = math.Round(totalAmount*100) / 100

	targetActive := !add

	if !dryRun {
		updateQuery := cfg.AssessmentDB.WithContext(ctx).
			Model(&TaxLedger{}).
			Where("assessment_number_hash = ? AND type = 'DR'", asmHash)
		if len(cleanTaxTypes) > 0 {
			updateQuery = updateQuery.Where("UPPER(tax_type) IN (?)", cleanTaxTypes)
		}

		if err := updateQuery.Update(`"isActive"`, targetActive).Error; err != nil {
			return nil, fmt.Errorf("failed to update tax_ledgers for %s: %w", asmClean, err)
		}
	}

	actionText := "deactivated"
	if targetActive {
		actionText = "reactivated"
	}

	var msg string
	if len(cleanTaxTypes) > 0 {
		msg = fmt.Sprintf("Tax ledgers %s successfully (isActive=%t, tax_types: [%s], count: %d, total: %.2f)", actionText, targetActive, strings.Join(cleanTaxTypes, ", "), len(ledgers), totalAmount)
		if dryRun {
			msg = fmt.Sprintf("[DRY RUN] Tax ledgers would be %s (isActive=%t, tax_types: [%s], count: %d, total: %.2f)", actionText, targetActive, strings.Join(cleanTaxTypes, ", "), len(ledgers), totalAmount)
		}
	} else {
		msg = fmt.Sprintf("Tax ledgers %s successfully (isActive=%t, count: %d, total: %.2f)", actionText, targetActive, len(ledgers), totalAmount)
		if dryRun {
			msg = fmt.Sprintf("[DRY RUN] Tax ledgers would be %s (isActive=%t, count: %d, total: %.2f)", actionText, targetActive, len(ledgers), totalAmount)
		}
	}

	return &ReconciliationItemResult{
		AssessmentNumber: asmClean,
		Status:           "SUCCESS",
		Message:          msg,
		ReconciledAmount: totalAmount,
	}, nil
}
