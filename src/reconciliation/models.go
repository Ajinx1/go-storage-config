package reconciliation

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Assessment struct {
	ID               string     `gorm:"column:id;primaryKey"`
	AssessmentNo     string     `gorm:"column:assessmentNo;uniqueIndex"`
	TIN              string     `gorm:"column:tin"`
	BusinessName     string     `gorm:"column:businessName"`
	TaxType          string     `gorm:"column:taxType"`
	AssessmentType   string     `gorm:"column:assessmentType"`
	AssessmentYear   int        `gorm:"column:assessmentYear"`
	AssessmentPeriod string     `gorm:"column:assessmentPeriod"`
	PeriodStart      string     `gorm:"column:periodStart"`
	PeriodEnd        string     `gorm:"column:periodEnd"`
	ExpirationDate   string     `gorm:"column:Expiration_date"`
	Amount           float64    `gorm:"column:amount"`
	TotalAmount      float64    `gorm:"column:total_amount"`
	Currency         string     `gorm:"column:currency"`
	PaymentStatus    string     `gorm:"column:paymentStatus"`
	FilingStatus     string     `gorm:"column:filingStatus"`
	PaymentRefStatus string     `gorm:"column:paymentRefStatus"`
	PaymentReference string     `gorm:"column:paymentReference"`
	PaymentDate      *time.Time `gorm:"column:payment date"`
	OfficeID         string     `gorm:"column:officeId"`
	StateID          string     `gorm:"column:stateId"`
	TaxLiability     string     `gorm:"column:taxLiability;type:jsonb"`
	SIP              bool       `gorm:"column:sip"`
}

func (Assessment) TableName() string {
	return "assessments"
}

// TaxLedger maps to the 'tax_ledgers' table in the Assessment DB.
type TaxLedger struct {
	ID                        string    `gorm:"column:id;primaryKey"`
	DateCreated               time.Time `gorm:"column:date_created"`
	DateUpdated               time.Time `gorm:"column:date_updated"`
	IsDeleted                 bool      `gorm:"column:isDeleted"`
	CreatedBy                 string    `gorm:"column:created_by"`
	UpdatedBy                 string    `gorm:"column:updated_by"`
	TINHash                   string    `gorm:"column:tin_hash"`
	TINEncrypted              string    `gorm:"column:tin_encrypted"`
	AssessmentNumberHash      string    `gorm:"column:assessment_number_hash"`
	AssessmentNumberEncrypted string    `gorm:"column:assessment_number_encrypted"`
	TaxType                   string    `gorm:"column:tax_type"`
	AssessmentPeriod          string    `gorm:"column:assessment_period"`
	AssessmentMonth           *string   `gorm:"column:assessment_month"`
	Description               string    `gorm:"column:description"`
	Amount                    float64   `gorm:"column:amount"`
	Balance                   float64   `gorm:"column:balance"`
	OfficeID                  string    `gorm:"column:officeId"`
	StateID                   string    `gorm:"column:stateId"`
	Currency                  string    `gorm:"column:currency"`
	Type                      string    `gorm:"column:type"` // "DR" or "CR"
	PaymentMethod             string    `gorm:"column:payment_method"`
}

func (TaxLedger) TableName() string {
	return "tax_ledgers"
}

// Transaction maps to 'tbl_transactions' in the Payment DB.
type Transaction struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement"`
	AssessmentNumber string    `gorm:"column:assessment_number"`
	PaymentReference string    `gorm:"column:payment_reference"`
	TransactionID    string    `gorm:"column:transaction_id"`
	Amount           float64   `gorm:"column:amount"`
	Status           string    `gorm:"column:status"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (Transaction) TableName() string {
	return "tbl_transactions"
}

// PaymentBreakdown maps to 'tbl_payment_breakdowns' in the Payment DB.
type PaymentBreakdown struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement"`
	AssessmentNumber string    `gorm:"column:assessment_number"`
	TaxTypes         string    `gorm:"column:tax_types;type:jsonb"`
	Amount           float64   `gorm:"column:amount"`
	TotalAmount      float64   `gorm:"column:total_amount"`
	RemainingAmount  float64   `gorm:"column:remaining_amount"`
	Status           string    `gorm:"column:status"`
	OfficeID         string    `gorm:"column:office_id"`
	StateID          string    `gorm:"column:state_id"`
	Currency         string    `gorm:"column:currency"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (PaymentBreakdown) TableName() string {
	return "tbl_payment_breakdowns"
}

// Collection maps to 'collections' table in the Assessment DB.
type Collection struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement"`
	Tax              string    `gorm:"column:tax"`
	OfficeID         string    `gorm:"column:officeId"`
	StateID          string    `gorm:"column:stateId"`
	Description      string    `gorm:"column:description"`
	Currency         string    `gorm:"column:currency"`
	Amount           float64   `gorm:"column:amount"`
	AssessmentPeriod string    `gorm:"column:assessmentPeriod"`
	AssessmentNo     string    `gorm:"column:assessmentNo"`
	TaxID            string    `gorm:"column:taxId"`
	PaymentReference string    `gorm:"column:paymentReference"`
	TransactionID    string    `gorm:"column:transactionId"`
	PaymentBankRef   string    `gorm:"column:paymentBankRef"`
	PaymentGate      string    `gorm:"column:paymentGate"`
	PaymentDate      time.Time `gorm:"column:paymentDate"`
	CreatedAt        time.Time `gorm:"column:createdAt"`
	UpdatedAt        time.Time `gorm:"column:updatedAt"`
}

func (Collection) TableName() string {
	return "collections"
}

// PaymentData maps to 'tbl_payment_data_db' in the Payment DB.
type PaymentData struct {
	ID                  string     `gorm:"column:id;primaryKey"`
	ReceiptNumber       string     `gorm:"column:receipt_number;uniqueIndex"`
	TaxType             string     `gorm:"column:tax_type"`
	Currency            string     `gorm:"column:currency"`
	Amount              float64    `gorm:"column:amount"`
	AssessmentNumber    string     `gorm:"column:assessment_number"`
	PaymentReference    string     `gorm:"column:payment_reference"`
	PayerName           string     `gorm:"column:payer_name"`
	PayerTIN            string     `gorm:"column:payer_tin"`
	AssessmentPeriod    string     `gorm:"column:assessment_period"`
	OfficeID            string     `gorm:"column:office_id"`
	ContractDate        *time.Time `gorm:"column:contract_date"`
	PeriodEnd           *time.Time `gorm:"column:period_end"`
	TaxYear             int        `gorm:"column:tax_year"`
	ExpiryDate          *time.Time `gorm:"column:expiry_date"`
	PaymentDate         time.Time  `gorm:"column:payment_date"`
	PaymentStatus       string     `gorm:"column:payment_status"`
	FilingStatus        string     `gorm:"column:filing_status"`
	TaxPayable          string     `gorm:"column:tax_payable;type:jsonb"`
	MetaData            string     `gorm:"column:meta_data;type:jsonb"`
	Vendor              string     `gorm:"column:vendor"`
	PaymentCode         string     `gorm:"column:payment_code"`
	PsspReferenceNumber string     `gorm:"column:pssp_reference_number"`
	SettlementStatus    string     `gorm:"column:settlement_status"`
	Signature           string     `gorm:"column:signature"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (PaymentData) TableName() string {
	return "tbl_payment_data_db"
}

// LegacyReconciliationLog maps to 'tbl_legacy_payment_reconciliation_logs' in Payment DB.

// AuditLog maps to 'audit_logs' in the Central Audit DB.
type AuditLog struct {
	ID          string    `gorm:"column:id;primaryKey"`
	Service     string    `gorm:"column:service"`
	Action      string    `gorm:"column:action"`
	Description string    `gorm:"column:description"`
	EntityType  string    `gorm:"column:entity_type"`
	EntityID    string    `gorm:"column:entity_id"`
	UserID      string    `gorm:"column:user_id"`
	OfficeID    string    `gorm:"column:office_id"`
	StateID     string    `gorm:"column:state_id"`
	Payload     string    `gorm:"column:payload;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (AuditLog) TableName() string {
	return "tbl_audit_logs"
}

// Tax Models & JSON helpers

type TaxLiabilityWrapper struct {
	TaxLiabilities      []TaxLiabilityItem `json:"taxLiabilities"`
	TotalTaxLiabilities float64            `json:"totalTaxLiabilities"`
}

type TaxLiabilityItem struct {
	TaxID       string  `json:"taxId"`
	TaxCode     string  `json:"taxCode"`
	TaxCodeName string  `json:"taxCodeName"`
	Amount      float64 `json:"amount"`
}

type BreakdownTaxItem struct {
	TaxCode string  `json:"tax_code"`
	Amount  float64 `json:"amount"`
}

type TaxPaymentItem struct {
	TaxCode string  `json:"taxCode"`
	Amount  float64 `json:"amount"`
}

// Config holds DB sessions, secrets, and operational params.
type Config struct {
	PaymentDB           *gorm.DB
	AssessmentDB        *gorm.DB
	AuditDB             *gorm.DB
	LedgerEncryptionKey string
	PaymentSigningKey   string
	BankPool            []string
	WritePaymentRecord  bool
}

// Request & Response DTOs

type BatchReconcileRequest struct {
	TheType               string    `json:"the_type,omitempty"`
	Type                  string    `json:"type,omitempty"`
	Add                   *bool     `json:"add,omitempty"`
	DryRun                bool      `json:"dry_run,omitempty"`
	AssessmentNumbers     []string  `json:"assessment_numbers"`
	PaymentReferences     []string  `json:"payment_references,omitempty"`
	BankReferences        []string  `json:"bank_references,omitempty"`
	PaymentChannels       []string  `json:"payment_channels,omitempty"`
	PaymentDates          []string  `json:"payment_dates,omitempty"`
	Amounts               []float64 `json:"amounts,omitempty"`
	OfficeIDs             []string  `json:"office_ids,omitempty"`
	StateIDs              []string  `json:"state_ids,omitempty"`
	UserIDs               []string  `json:"user_ids,omitempty"`
	UserNames             []string  `json:"user_names,omitempty"`
	WritePaymentRecord    *bool     `json:"write_payment_record,omitempty"`
	UpdatePenaltyInterest bool      `json:"update_penalty_interest,omitempty"`
	BankRefMode           string    `json:"bank_ref_mode,omitempty"`
}

func (r *BatchReconcileRequest) GetType() string {
	t := strings.ToLower(strings.TrimSpace(r.TheType))
	if t == "" {
		t = strings.ToLower(strings.TrimSpace(r.Type))
	}
	return t
}

func (r *BatchReconcileRequest) IsAdd() bool {
	if r.Add == nil {
		return true
	}
	return *r.Add
}

type ReconciliationItemResult struct {
	AssessmentNumber string  `json:"assessment_number"`
	PaymentReference string  `json:"payment_reference"`
	BankReference    string  `json:"bank_reference"`
	ReceiptNumber    string  `json:"receipt_number"`
	ReconciledAmount float64 `json:"reconciled_amount"`
	Status           string  `json:"status"` // SUCCESS | FAILED
	Message          string  `json:"message"`
}

type BatchReconcileResponse struct {
	TotalAttempted        int                        `json:"total_attempted"`
	TotalSuccessful       int                        `json:"total_successful"`
	TotalFailed           int                        `json:"total_failed"`
	TotalAmountReconciled float64                    `json:"total_amount_reconciled"`
	Results               []ReconciliationItemResult `json:"results"`
}

type ReconciliationInput struct {
	AssessmentNumber   string
	PaymentReference   string
	BankReference      string
	PaymentChannel     string
	PaymentDate        string
	Amount             float64
	OfficeID           string
	StateID            string
	UserID             string
	UserName           string
	WritePaymentRecord bool
}
