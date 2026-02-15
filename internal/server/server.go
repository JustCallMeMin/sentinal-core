package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/version"
)

// Server holds the Fiber app and configurations
type Server struct {
	App    *fiber.App
	Config *config.Config
	// TODO(SC-011): Add DB *pgxpool.Pool here
}

// New creates a new Server instance
func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		AppName:       "Sentinal Core " + version.Version,
		StrictRouting: true,
		ServerHeader:  "Sentinal",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(TraceMiddleware()) // Our custom structured logger + trace ID

	// TODO(SC-031): Move routes to internal/server/routes.go when scaling

	// Base Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "sentinal-core",
			"version": version.Version,
			"status":  "uplink_healthy",
		})
	})

	// Health Check (Liveness)
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "sentinal-core",
			"version":   version.Version,
			"timestamp": time.Now().Unix(),
		})
	})

	// Readiness Check
	app.Get("/readyz", func(c *fiber.Ctx) error {
		// TODO(SC-011): Perform actual DB ping here
		// if err := s.DB.Ping(); err != nil { return c.Status(503).JSON(...) }

		return c.JSON(fiber.Map{
			"status":    "ready",
			"service":   "sentinal-core",
			"version":   version.Version,
			"timestamp": time.Now().Unix(),
		})
	})

	return &Server{
		App:    app,
		Config: cfg,
	}
}

// Listen starts the server on the given address
func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}
