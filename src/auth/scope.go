package auth

import (
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (c *Claims) ScopeQuery(
	db *gorm.DB,
	column string,
) *gorm.DB {

	switch {

	case c.IsNational():
		return db

	case c.IsRegional():

		officeIDs := c.GetOfficeIDs()

		if len(officeIDs) == 0 {
			return db.Where("1 = 0")
		}

		return db.Where(fmt.Sprintf(`%s = ANY(?)`, column), pq.Array(officeIDs))

	case c.IsOffice():
		return db.Where(fmt.Sprintf(`%s = ?`, column), c.GetOfficeIDString())

	default:
		return db.Where("1 = 0")
	}
}
