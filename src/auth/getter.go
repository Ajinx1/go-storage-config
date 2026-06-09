package auth

import (
	"strconv"
	"strings"
	"time"
)

func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.Exp
}

func (c *Claims) GetUserID() int64 {
	return c.Sub
}

func (c *Claims) GetTaxpayerID() string {
	return c.ID
}

func (c *Claims) GetTIN() string {
	return c.TIN
}

func (c *Claims) GetTaxID() string {
	return c.TaxID
}

func (c *Claims) GetDisplayName() string {
	return strings.TrimSpace(c.FullName)
}

func (c *Claims) GetOfficeID() int64 {

	if c.Office == nil {
		return 0
	}

	id, err := strconv.ParseInt(c.Office.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetOfficeCode() string {

	if c.Office == nil {
		return ""
	}

	return strings.TrimSpace(c.Office.Code)
}

func (c *Claims) GetOfficeName() string {

	if c.Office == nil {
		return ""
	}

	return strings.TrimSpace(c.Office.Title)
}

func (c *Claims) GetStateID() int64 {

	if c.Office == nil || c.Office.State == nil {
		return 0
	}

	id, err := strconv.ParseInt(c.Office.State.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetOfficeGroupID() int64 {

	if c.Office == nil || c.Office.OfficeGroup == nil {
		return 0
	}

	id, err := strconv.ParseInt(c.Office.OfficeGroup.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetStateGroupID() int64 {

	if c.Office == nil || c.Office.StateGroup == nil {
		return 0
	}

	id, err := strconv.ParseInt(c.Office.StateGroup.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetStateName() string {

	if c.Office == nil || c.Office.State == nil {
		return ""
	}

	return strings.TrimSpace(c.Office.State.Title)
}

func (c *Claims) GetLGAName() string {

	if c.Office == nil || c.Office.LGA == nil {
		return ""
	}

	return strings.TrimSpace(c.Office.LGA.LGA)
}

func (c *Claims) GetOfficeGroupName() string {

	if c.Office == nil || c.Office.OfficeGroup == nil {
		return ""
	}

	return strings.TrimSpace(c.Office.OfficeGroup.Title)
}
