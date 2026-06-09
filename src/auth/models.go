package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	// Staff fields
	Sub             int64   `json:"sub,omitempty"`
	IRNumber        string  `json:"ir_number,omitempty"`
	Designation     string  `json:"designation,omitempty"`
	Department      string  `json:"department,omitempty"`
	TaxGroup        string  `json:"taxGroup,omitempty"`
	Coverage        string  `json:"coverage,omitempty"`
	RegionalOffices []int64 `json:"regionalOffices,omitempty"`

	// Taxpayer fields
	ID    string `json:"id,omitempty"`
	TaxID string `json:"taxId,omitempty"`
	TIN   string `json:"tin,omitempty"`

	// Common fields
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	UserType string `json:"userType"`

	// Office
	Office *Office `json:"office,omitempty"`
	Iat    int64   `json:"iat"`
	Exp    int64   `json:"exp"`

	jwt.RegisteredClaims
}

type Office struct {
	ID            string       `json:"id"`
	DateCreated   string       `json:"dateCreated,omitempty"`
	CreatedBy     string       `json:"createdBy,omitempty"`
	Title         string       `json:"title"`
	Code          string       `json:"code"`
	GroupSec      string       `json:"groupSec,omitempty"`
	Address       *string      `json:"address,omitempty"`
	Sector        string       `json:"sector,omitempty"`
	GcRPT         bool         `json:"gcRPT,omitempty"`
	TaxController *string      `json:"taxController,omitempty"`
	State         *State       `json:"state,omitempty"`
	LGA           *LGA         `json:"lga,omitempty"`
	StateGroup    *StateGroup  `json:"stateGroup,omitempty"`
	OfficeGroup   *OfficeGroup `json:"officeGroup,omitempty"`

	// Taxpayer token fields
	Region    *Region   `json:"region,omitempty"`
	TaxGroup2 *TaxGroup `json:"taxGroup,omitempty"`
}

type State struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	Title       string `json:"title"`
}

type LGA struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	LGA         string `json:"lga"`
	State       *State `json:"state,omitempty"`
}

type StateGroup struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	Title       string `json:"title"`
}

type OfficeGroup struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	Title       string `json:"title"`
}

type Region struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	Title       string `json:"title"`
}

type TaxGroup struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	Title       string `json:"title"`
}

func (c *Claims) String() string {
	return fmt.Sprintf("%s (%s) [%s]",
		c.GetDisplayName(),
		c.Email,
		c.Coverage,
	)
}
