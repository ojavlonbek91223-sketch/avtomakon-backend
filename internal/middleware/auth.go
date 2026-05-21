package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/avtomakon/backend/internal/pkg/jwt"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

func RequireAuth(mgr *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Token kerak")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "Noto'g'ri Authorization format")
		}

		claims, err := mgr.ParseToken(parts[1])
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Noto'g'ri yoki muddati o'tgan token")
		}

		c.Locals(ContextUserID, claims.UserID)
		c.Locals(ContextRole, claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(ContextRole).(string)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "Auth kerak")
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusForbidden, "Ruxsat yo'q")
	}
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	v, ok := c.Locals(ContextUserID).(uuid.UUID)
	return v, ok
}

// OptionalAuth — agar token bo'lsa, foydalanuvchi ID'sini contextga qo'yadi,
// bo'lmasa ham handler ishlayveradi. Public endpointlar uchun
// (masalan, feed — login bo'lmasa ham public postlar ko'rinadi).
func OptionalAuth(mgr *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return c.Next()
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}
		claims, err := mgr.ParseToken(parts[1])
		if err != nil {
			return c.Next() // yaroqsiz token — anonim sifatida davom
		}
		c.Locals(ContextUserID, claims.UserID)
		c.Locals(ContextRole, claims.Role)
		return c.Next()
	}
}
