package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"student-marketplace/internal/models"
	jwtpkg "student-marketplace/pkg/jwt"
)

func RequireAuth(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return fiber.ErrUnauthorized
	}

	claims, err := jwtpkg.ValidateAccessToken(token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("user_email", claims.Email)
	c.Locals("user_role", claims.Role)

	return c.Next()
}

func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok {
			return fiber.ErrUnauthorized
		}

		for _, r := range roles {
			if string(r) == role {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}
}

func OptionalAuth(c *fiber.Ctx) error {
	token := extractToken(c)
	if token == "" {
		return c.Next()
	}

	claims, err := jwtpkg.ValidateAccessToken(token)
	if err != nil {
		return c.Next()
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("user_email", claims.Email)
	c.Locals("user_role", claims.Role)

	return c.Next()
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals("user_id").(uuid.UUID)
	return id, ok
}

func GetUserRole(c *fiber.Ctx) string {
	role, _ := c.Locals("user_role").(string)
	return role
}

func extractToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// Also check cookie for web clients
	if cookie := c.Cookies("access_token"); cookie != "" {
		return cookie
	}
	return ""
}
