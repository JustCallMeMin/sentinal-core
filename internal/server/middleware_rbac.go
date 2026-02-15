package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sentinal/core/pkg/logger"
	"go.uber.org/zap"
)

// RBACMiddleware check if the user has a specific permission slug in their claims
func RBACMiddleware(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, _ := c.Locals("permissions").([]string)
		userID, _ := c.Locals("user_id").(string)

		hasPermission := false
		for _, p := range permissions {
			if p == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			logger.Warn("Access denied: missing required permission",
				zap.String("user_id", userID),
				zap.String("required_permission", requiredPermission),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":    "Access denied: missing required permission",
				"required": requiredPermission,
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission checks if the user has at least one of the required permissions
func RequireAnyPermission(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, _ := c.Locals("permissions").([]string)
		userID, _ := c.Locals("user_id").(string)

		hasAny := false
		for _, required := range requiredPermissions {
			for _, p := range permissions {
				if p == required {
					hasAny = true
					break
				}
			}
			if hasAny {
				break
			}
		}

		if !hasAny {
			logger.Warn("Access denied: missing any required permission",
				zap.String("user_id", userID),
				zap.Strings("required_permissions", requiredPermissions),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":    "Access denied: missing any required permission",
				"required": requiredPermissions,
			})
		}

		return c.Next()
	}
}
