package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sentinal/core/pkg/version"
)

// Server holds the Fiber app and configurations
type Server struct {
	App *fiber.App
}

// New creates a new Server instance
func New() *Server {
	app := fiber.New(fiber.Config{
		AppName:       "Sentinal Core " + version.Version,
		StrictRouting: true,
		ServerHeader:  "Sentinal",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Base Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "sentinal-core",
			"version": version.Version,
			"status":  "uplink_healthy",
		})
	})

	return &Server{
		App: app,
	}
}

// Listen starts the server on the given address
func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}
