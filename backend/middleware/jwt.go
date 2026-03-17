package middleware

import (
	"strings"

	"outcraftly/accounts/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected is a Fiber middleware that validates the Bearer JWT (or cookie).
// On success it sets "userID", "email", and "role" in c.Locals().
// Tokens are validated against JWT_ISSUER and JWT_AUDIENCE.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract token — prefer Authorization header, fall back to cookie.
		tokenStr := ""
		if authHeader := c.Get("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				tokenStr = parts[1]
			}
		}
		if tokenStr == "" {
			tokenStr = c.Cookies("accounts_token")
		}
		if tokenStr == "" {
			return unauthorized(c, "Missing token (header or cookie)")
		}

		// 2. Parse & validate — enforce issuer + audience.
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(config.Cfg.JWTSecret), nil
		},
			jwt.WithIssuer(config.Cfg.JWTIssuer),
			jwt.WithAudience(config.Cfg.JWTAudience),
		)

		if err != nil || !token.Valid {
			return unauthorized(c, "Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return unauthorized(c, "Invalid token claims")
		}

		// 3. Populate locals.
		c.Locals("userID", claims["sub"].(string))
		c.Locals("email", claims["email"].(string))
		role, _ := claims["role"].(string)
		if role == "" {
			role = "user"
		}
		c.Locals("role", role)

		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": msg})
}
