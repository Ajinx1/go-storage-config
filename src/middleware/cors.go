package middleware

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CORS(cfg CORSConfig) fiber.Handler {

	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))

	for _, origin := range cfg.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	localhostRe := regexp.MustCompile(
		`^https?://(localhost|127\.0\.0\.1)(:\d+)?$`,
	)

	isOriginAllowed := func(origin string) bool {

		if _, ok := allowedOrigins[origin]; ok {
			return true
		}

		if cfg.AllowLocalhost && localhostRe.MatchString(origin) {
			return true
		}

		return false
	}

	return func(c *fiber.Ctx) error {

		origin := c.Get("Origin")

		if origin != "" && isOriginAllowed(origin) {

			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-API-KEY",
			)
			c.Set(
				"Access-Control-Allow-Methods",
				"POST, OPTIONS, GET, PUT, DELETE, PATCH",
			)
			c.Set("Access-Control-Max-Age", "86400")
			c.Set("Vary", "Origin")
		}

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
