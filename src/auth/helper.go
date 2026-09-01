package auth

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func Middleware(secret string) fiber.Handler {

	return func(c *fiber.Ctx) error {

		claims, err := GetClaims(c.Get("Authorization"), secret)

		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(ContextClaimsKey, claims)

		return c.Next()
	}
}

func ClaimsFromContext(c *fiber.Ctx) (*Claims, error) {

	claims, ok := c.Locals(ContextClaimsKey).(*Claims)
	if !ok {
		return nil, errors.New("user claims not found")
	}

	return claims, nil
}

func (c *Claims) GetActorID() string {

	if c.IsTaxpayer() {
		return c.GetTaxID()
	}

	return string(c.GetUserID())
}

func (c *Claims) PopulateRequestScope(req ScopedRequest) {

	scope := req.GetRequestScope()

	scope.UserID = c.GetActorID()
	scope.UserName = c.GetDisplayName()

	if c.Office != nil {
		scope.OfficeID = c.Office.ID
	}

	if scope.OfficeID == "" && c.IsRegional() {
		officeIDs := c.GetOfficeIDs()
		if len(officeIDs) > 0 {
			scope.OfficeID = strconv.FormatInt(officeIDs[0], 10)
		}
	}

	if stateID := c.GetStateID(); stateID > 0 {
		scope.StateID = strconv.FormatInt(stateID, 10)
	}

}
