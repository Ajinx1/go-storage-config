package auth

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

type StringOrNumber string

func (s *StringOrNumber) UnmarshalJSON(data []byte) error {

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringOrNumber(str)
		return nil
	}

	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = StringOrNumber(strconv.FormatInt(num, 10))
		return nil
	}

	var floatNum float64
	if err := json.Unmarshal(data, &floatNum); err == nil {
		*s = StringOrNumber(strconv.FormatFloat(floatNum, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("failed to parse sub as string or number: %s", string(data))
}

func (s StringOrNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

type RegionalOffices []int64

func (r *RegionalOffices) UnmarshalJSON(data []byte) error {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return err
	}
	res := make([]int64, len(rawItems))
	for i, item := range rawItems {
		var s string
		if err := json.Unmarshal(item, &s); err == nil {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %q as int64: %w", s, err)
			}
			res[i] = val
			continue
		}
		var val int64
		if err := json.Unmarshal(item, &val); err != nil {
			return fmt.Errorf("failed to parse array element as string or int64: %s", string(item))
		}
		res[i] = val
	}
	*r = res
	return nil
}

func (r RegionalOffices) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	strings := make([]string, len(r))
	for i, val := range r {
		strings[i] = strconv.FormatInt(val, 10)
	}
	return json.Marshal(strings)
}

type Claims struct {
	Sub             StringOrNumber  `json:"sub,omitempty"`
	IRNumber        string          `json:"ir_number,omitempty"`
	Designation     string          `json:"designation,omitempty"`
	Department      string          `json:"department,omitempty"`
	TaxGroup        string          `json:"taxGroup,omitempty"`
	Coverage        string          `json:"coverage,omitempty"`
	RegionalOffices RegionalOffices `json:"regionalOffices,omitempty"`

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

type RequestScope struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	OfficeID string `json:"office_id"`
	StateID  string `json:"state_id"`
}

type ScopedRequest interface {
	GetRequestScope() *RequestScope
}

func (c *Claims) String() string {
	return fmt.Sprintf("%s (%s) [%s]",
		c.GetDisplayName(),
		c.Email,
		c.Coverage,
	)
}
