package auth

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (c *Claims) ScopeTaxpayerQuery(
	db *gorm.DB,
	column string,
) (*gorm.DB, error) {

	if !c.IsTaxpayer() {
		return db, nil
	}

	return db.Where(fmt.Sprintf(`%s = ?`, column), c.GetTaxID()), nil
}

func (c *Claims) ScopeQuery(
	db *gorm.DB,
	column string,
) (*gorm.DB, error) {

	switch {

	case c.IsNational():
		return db, nil

	case c.IsRegional():

		officeIDs := c.GetOfficeIDs()

		if len(officeIDs) == 0 {
			return nil, errors.New("regional user has no assigned offices")
		}

		return db.Where(fmt.Sprintf(`%s = ANY(?)`, column), pq.Array(officeIDs)), nil

	case c.IsOffice():
		officeID := c.GetOfficeIDString()

		if officeID == "" {
			return nil, errors.New("office user has no assigned office")
		}

		return db.Where(fmt.Sprintf(`%s = ?`, column), officeID), nil

	default:
		return nil, errors.New("invalid user coverage level")
	}
}
