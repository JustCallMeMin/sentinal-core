package transaction

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler manages transaction endpoints
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// Create handles the ingestion of a new transaction
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateTransactionRequest

	// Parse Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate Payload
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Auto-extract IP Address from request (handles X-Forwarded-For, X-Real-IP)
	if req.IPAddress == "" {
		req.IPAddress = c.IP()
	}

	// Extract Tenant ID (Mock Auth for now until SC-029)
	tenantIDStr := c.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing X-Tenant-ID header",
		})
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Tenant ID format",
		})
	}

	// Call Service
	resp, err := h.service.Create(c.UserContext(), tenantID, &req)
	if err != nil {
		// In production, we should map specific errors (e.g. duplicate) to status codes
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
