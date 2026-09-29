package reconciliation

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PrepareAssessmentInput auto-extracts transaction and breakdown info from payment DB,
// generating synthetic bank references and payment dates if not provided.
func PrepareAssessmentInput(
	paymentDB *gorm.DB,
	assessmentNumber string,
	bankPool []string,
	bankRefMode string,
	providedPaymentDate string,
	providedBankRef string,
	providedAmount float64,
	providedPaymentChannel string,
	officeID, stateID, userID, userName string,
	writePaymentRecord bool,
) (*ReconciliationInput, error) {
	if strings.TrimSpace(assessmentNumber) == "" {
		return nil, errors.New("assessment number is required")
	}

	// 1. Fetch latest transaction
	var tx Transaction
	err := paymentDB.Where("assessment_number = ?", assessmentNumber).
		Order("id DESC").
		First(&tx).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch transaction for %s: %w", assessmentNumber, err)
	}

	paymentReference := tx.PaymentReference

	// 2. Fetch latest breakdown
	var breakdown PaymentBreakdown
	err = paymentDB.Where("assessment_number = ?", assessmentNumber).
		Order("id DESC").
		First(&breakdown).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch breakdown for %s: %w", assessmentNumber, err)
	}

	amount := providedAmount
	if amount <= 0 {
		amount = breakdown.Amount
	}

	// 3. Format or calculate payment date
	paymentDateStr := strings.TrimSpace(providedPaymentDate)
	if paymentDateStr == "" {
		// Default to ~60-180 minutes ago with jitter
		jitterSec, _ := rand.Int(rand.Reader, big.NewInt(7200))
		calcDate := time.Now().UTC().Add(-time.Duration(3600+jitterSec.Int64()) * time.Second)
		paymentDateStr = calcDate.Format("2006-01-02 15:04:05")
	}

	// 4. Generate bank reference if not provided
	bankRef := strings.TrimSpace(providedBankRef)
	if bankRef == "" {
		if strings.EqualFold(bankRefMode, "same") || len(bankPool) == 0 {
			bankRef = paymentReference
		} else {
			bankIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(bankPool))))
			selectedBank := bankPool[bankIdx.Int64()]

			var dateFormatted string
			parsedDt, err := time.Parse("2006-01-02 15:04:05", paymentDateStr)
			if err == nil {
				dateFormatted = parsedDt.Format("02-01-2006")
			} else {
				dateFormatted = time.Now().Format("02-01-2006")
			}

			truncatedRef := paymentReference
			if len(truncatedRef) > 2 {
				truncatedRef = truncatedRef[2:]
			}
			bankRef = fmt.Sprintf("%s/SPEC/7/%s/%s", selectedBank, dateFormatted, truncatedRef)
		}
	}

	return &ReconciliationInput{
		AssessmentNumber:   assessmentNumber,
		PaymentReference:   paymentReference,
		BankReference:      bankRef,
		PaymentChannel:     providedPaymentChannel,
		PaymentDate:        paymentDateStr,
		Amount:             amount,
		OfficeID:           officeID,
		StateID:            stateID,
		UserID:             userID,
		UserName:           userName,
		WritePaymentRecord: writePaymentRecord,
	}, nil
}
