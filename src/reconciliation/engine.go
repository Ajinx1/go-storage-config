package reconciliation

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Engine orchestrates batch and single assessment reconciliation workflows.
type Engine struct {
	cfg Config
}

// NewEngine initializes and validates the reconciliation engine.
func NewEngine(cfg Config) (*Engine, error) {
	if cfg.PaymentDB == nil {
		return nil, errors.New("paymentDB is required")
	}
	if cfg.AssessmentDB == nil {
		return nil, errors.New("assessmentDB is required")
	}
	if len(cfg.BankPool) == 0 {
		cfg.BankPool = []string{"PRB", "GTB", "ZIB"}
	}

	if cfg.AuditDB == nil {
		return nil, errors.New("auditDB is required")
	}

	if cfg.LedgerEncryptionKey == "" {
		cfg.LedgerEncryptionKey = "6f6c8df539d6b8aa243c7c9f14e1b3451e9c3a02b613cb1784b64db4e4f6e70a"
	}

	if cfg.PaymentSigningKey == "" {
		cfg.PaymentSigningKey = "4c1238166435c4813f66cc45af6b74b8f2e7178edfa3bc9f702173b50cce4b7d4252a2a11943f4ce55a4fcf55f36270a86f11ad4310a5b62f01549bac5715c1a"
	}

	cfg.WritePaymentRecord = true

	return &Engine{cfg: cfg}, nil
}

func getConfigVal(arr []string, idx int, defaultVal string) string {
	if len(arr) == 0 {
		return defaultVal
	}
	if idx < len(arr) {
		return arr[idx]
	}
	return arr[len(arr)-1]
}

func getFloatVal(arr []float64, idx int, defaultVal float64) float64 {
	if len(arr) == 0 {
		return defaultVal
	}
	if idx < len(arr) {
		return arr[idx]
	}
	return arr[len(arr)-1]
}

func (e *Engine) ReconcileBatch(ctx context.Context, req BatchReconcileRequest) (*BatchReconcileResponse, error) {
	if len(req.AssessmentNumbers) == 0 {
		return nil, errors.New("at least one assessment number is required")
	}

	reqType := req.GetType()
	if reqType != "" && reqType != "staff" && reqType != "office" {
		return nil, fmt.Errorf("invalid the_type '%s': must be 'staff' or 'office'", reqType)
	}

	response := &BatchReconcileResponse{
		Results: make([]ReconciliationItemResult, 0, len(req.AssessmentNumbers)),
	}

	writeRecord := true
	if req.WritePaymentRecord != nil {
		writeRecord = *req.WritePaymentRecord
	}

	for i, asm := range req.AssessmentNumbers {
		asmClean := strings.TrimSpace(asm)
		if asmClean == "" {
			continue
		}

		response.TotalAttempted++

		userID := getConfigVal(req.UserIDs, i, "system")
		userName := getConfigVal(req.UserNames, i, "system")

		if reqType == "staff" {
			isAdd := req.IsAdd()
			res, err := ProcessStaffPenaltyInterest(ctx, &e.cfg, asmClean, isAdd, req.DryRun, userID, userName)
			if err != nil {
				response.TotalFailed++
				response.Results = append(response.Results, ReconciliationItemResult{
					AssessmentNumber: asmClean,
					Status:           "FAILED",
					Message:          err.Error(),
				})
			} else {
				response.TotalSuccessful++
				response.TotalAmountReconciled += res.ReconciledAmount
				response.Results = append(response.Results, *res)
			}
			continue
		}

		if req.UpdatePenaltyInterest {
			res, err := UpdatePaymentPenaltyInterest(ctx, &e.cfg, asmClean, userID, userName)
			if err != nil {
				response.TotalFailed++
				response.Results = append(response.Results, ReconciliationItemResult{
					AssessmentNumber: asmClean,
					Status:           "FAILED",
					Message:          err.Error(),
				})
			} else {
				response.TotalSuccessful++
				response.TotalAmountReconciled += res.ReconciledAmount
				response.Results = append(response.Results, *res)
			}
			continue
		}

		paymentRef := getConfigVal(req.PaymentReferences, i, "")
		bankRef := getConfigVal(req.BankReferences, i, "")
		channel := getConfigVal(req.PaymentChannels, i, "")
		paymentDate := getConfigVal(req.PaymentDates, i, "")
		amount := getFloatVal(req.Amounts, i, 0.0)
		officeID := getConfigVal(req.OfficeIDs, i, "")
		stateID := getConfigVal(req.StateIDs, i, "")

		// Prepare / auto-discover missing inputs
		input, err := PrepareAssessmentInput(
			e.cfg.PaymentDB,
			asmClean,
			e.cfg.BankPool,
			req.BankRefMode,
			paymentDate,
			bankRef,
			amount,
			channel,
			officeID,
			stateID,
			userID,
			userName,
			writeRecord,
		)
		if err != nil {
			response.TotalFailed++
			response.Results = append(response.Results, ReconciliationItemResult{
				AssessmentNumber: asmClean,
				Status:           "FAILED",
				Message:          err.Error(),
			})
			continue
		}

		// If user explicitly provided paymentRef, override
		if paymentRef != "" {
			input.PaymentReference = paymentRef
		}

		// Execute flow
		res, err := ReconcileSingleAssessment(ctx, &e.cfg, *input)
		if err != nil {
			response.TotalFailed++
			response.Results = append(response.Results, ReconciliationItemResult{
				AssessmentNumber: asmClean,
				Status:           "FAILED",
				Message:          err.Error(),
			})
		} else {
			response.TotalSuccessful++
			response.TotalAmountReconciled += res.ReconciledAmount
			response.Results = append(response.Results, *res)
		}
	}

	return response, nil
}
