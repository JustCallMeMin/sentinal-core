package server

import (
	"github.com/gofiber/fiber/v2"
)

// RBACMiddleware check if the user has a specific permission slug in their claims
func RBACMiddleware(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Permissions are set in c.Locals by AuthMiddleware
		permissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions (no roles assigned)",
			})
		}

		// Check for the required permission slug
		hasPermission := false
		for _, p := range permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":    "Access denied: missing required permission",
				"required": permission,
			})
		}

		return c.Next()
	}
}
