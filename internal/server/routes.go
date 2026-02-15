package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sentinal/core/pkg/version"
)

// RegisterRoutes sets up all API routes
func (s *Server) RegisterRoutes() {
	// Base Routes
	s.App.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "sentinal-core",
			"version": version.Version,
			"status":  "uplink_healthy",
		})
	})

	// Health Check (Liveness)
	s.App.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "sentinal-core",
			"version":   version.Version,
			"timestamp": time.Now().Unix(),
		})
	})

	// Readiness Check
	s.App.Get("/readyz", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.DB.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "down",
				"service": "sentinal-core",
				"error":   "database connection lost",
			})
		}

		return c.JSON(fiber.Map{
			"status":    "ready",
			"service":   "sentinal-core",
			"version":   version.Version,
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 Group
	v1 := s.App.Group("/api/v1")

	// Auth Routes
	if s.AuthHandler != nil {
		auth := v1.Group("/auth")
		auth.Post("/login", s.AuthHandler.Login)
	}

	// Transaction Routes
	if s.TransactionHandler != nil {
		v1.Post("/transactions", s.TransactionHandler.Create)
	}
}
