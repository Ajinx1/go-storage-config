package auth

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

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
		return db.Where(fmt.Sprintf(`%s = ?`, column), c.GetOfficeIDString()), nil

	default:
		return nil, errors.New("invalid user coverage level")
	}
}
