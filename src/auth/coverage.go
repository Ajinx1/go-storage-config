package auth

import (
	"strconv"
	"strings"
)

const (
	CoverageOffice   = "OFFICE"
	CoverageRegional = "REGIONAL"
	CoverageNational = "NATIONAL"
	ContextClaimsKey = "user_claims"
)

func (c *Claims) IsOffice() bool {
	return strings.EqualFold(c.Coverage, CoverageOffice)
}

func (c *Claims) IsRegional() bool {
	return strings.EqualFold(c.Coverage, CoverageRegional)
}

func (c *Claims) IsNational() bool {
	return strings.EqualFold(c.Coverage, CoverageNational)
}

func (c *Claims) GetOfficeID() string {
	return c.Office.ID
}

func (c *Claims) GetOfficeIDs() []int64 {

	switch strings.ToUpper(c.Coverage) {

	case CoverageNational:
		return nil

	case CoverageRegional:
		return c.RegionalOffices

	case CoverageOffice:
		id, err := strconv.ParseInt(c.Office.ID, 10, 64)

		if err != nil {
			return []int64{}
		}

		return []int64{id}
	}

	return []int64{}
}

func (c *Claims) HasAccessToOffice(officeID int64) bool {

	if c.IsNational() {
		return true
	}

	for _, id := range c.GetOfficeIDs() {
		if id == officeID {
			return true
		}
	}

	return false
}
