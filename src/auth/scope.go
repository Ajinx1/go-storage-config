package auth

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (c *Claims) ScopeQuery(
	db *gorm.DB,
	staffColumn string,
	taxpayerColumn ...string,
) (*gorm.DB, error) {

	if c.IsTaxpayer() {

		if len(taxpayerColumn) == 0 {
			return db, nil
		}

		return db.Where(
			fmt.Sprintf(`%s = ?`, taxpayerColumn[0]),
			c.GetTaxID(),
		), nil
	}

	switch {

	case c.IsNational():
		return db, nil

	case c.IsRegional():

		officeIDs := c.GetOfficeIDs()

		if len(officeIDs) == 0 {
			return nil, errors.New("regional user has no assigned offices")
		}

		return db.Where(
			fmt.Sprintf(`%s = ANY(?)`, staffColumn),
			pq.Array(officeIDs),
		), nil

	case c.IsOffice():

		officeID := c.GetOfficeIDString()

		if officeID == "" {
			return nil, errors.New("office user has no assigned office")
		}

		return db.Where(
			fmt.Sprintf(`%s = ?`, staffColumn),
			officeID,
		), nil

	default:
		return nil, errors.New("invalid user coverage level")
	}
}
