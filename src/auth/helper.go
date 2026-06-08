package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func Middleware(secret string) fiber.Handler {

	return func(c *fiber.Ctx) error {

		claims, err := GetClaims(c.Get("Authorization"), secret)

		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals("user_claims", claims)

		return c.Next()
	}
}

func ClaimsFromContext(c *fiber.Ctx) (*Claims, error) {

	claims, ok := c.Locals("user_claims").(*Claims)
	if !ok {
		return nil, errors.New("user claims not found")
	}

	return claims, nil
}
