package auth

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	Sub             int64   `json:"sub"`
	Email           string  `json:"email"`
	FullName        string  `json:"fullName"`
	UserType        string  `json:"userType"`
	IRNumber        string  `json:"ir_number"`
	Designation     string  `json:"designation"`
	Department      any     `json:"department"`
	TaxGroup        string  `json:"taxGroup"`
	Coverage        string  `json:"coverage"`
	Office          Office  `json:"office"`
	RegionalOffices []int64 `json:"regionalOffices"`
	Iat             int64   `json:"iat"`
	Exp             int64   `json:"exp"`

	jwt.RegisteredClaims
}

type Office struct {
	ID            string      `json:"id"`
	DateCreated   string      `json:"dateCreated"`
	CreatedBy     string      `json:"createdBy"`
	Title         string      `json:"title"`
	Code          string      `json:"code"`
	GroupSec      string      `json:"groupSec"`
	Address       *string     `json:"address"`
	Sector        string      `json:"sector"`
	GcRPT         bool        `json:"gcRPT"`
	TaxController *string     `json:"taxController"`
	State         State       `json:"state"`
	LGA           LGA         `json:"lga"`
	StateGroup    StateGroup  `json:"stateGroup"`
	OfficeGroup   OfficeGroup `json:"officeGroup"`
}

type State struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated"`
	CreatedBy   string `json:"createdBy"`
	Title       string `json:"title"`
}

type LGA struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated"`
	CreatedBy   string `json:"createdBy"`
	LGA         string `json:"lga"`
	State       State  `json:"state"`
}

type StateGroup struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated"`
	CreatedBy   string `json:"createdBy"`
	Title       string `json:"title"`
}

type OfficeGroup struct {
	ID          string `json:"id"`
	DateCreated string `json:"dateCreated"`
	CreatedBy   string `json:"createdBy"`
	Title       string `json:"title"`
}
