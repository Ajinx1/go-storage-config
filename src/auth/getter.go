package auth

import (
	"strconv"
	"strings"
)

func (c *Claims) GetUserID() int64 {
	return c.Sub
}

func (c *Claims) GetOfficeID() int64 {
	id, err := strconv.ParseInt(c.Office.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetOfficeCode() string {
	return strings.TrimSpace(c.Office.Code)
}

func (c *Claims) GetOfficeName() string {
	return strings.TrimSpace(c.Office.Title)
}

func (c *Claims) GetDisplayName() string {
	return strings.TrimSpace(c.FullName)
}

func (c *Claims) GetStateID() int64 {
	id, err := strconv.ParseInt(c.Office.State.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetOfficeGroupID() int64 {
	id, err := strconv.ParseInt(c.Office.OfficeGroup.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (c *Claims) GetStateGroupID() int64 {
	id, err := strconv.ParseInt(c.Office.StateGroup.ID, 10, 64)
	if err != nil {
		return 0
	}

	return id
}
