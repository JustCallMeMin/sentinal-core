package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sentinal/core/pkg/logger"
	"go.uber.org/zap"
)

// TraceMiddleware injects a request ID and logs the request
func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get or generate trace ID
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set it in response header and locals
		c.Set("X-Trace-ID", traceID)
		c.Locals("trace_id", traceID)

		// Continue stack
		err := c.Next()

		// Log after request
		latency := time.Since(start)

		logFields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
		}

		// Determine log level based on status code, not just error presence
		statusCode := c.Response().StatusCode()

		if statusCode >= 500 {
			// Server errors (5xx)
			if err != nil {
				logFields = append(logFields, zap.Error(err))
			}
			logger.Error("Request failed - Server error", logFields...)
		} else if statusCode >= 400 {
			// Client errors (4xx)
			if err != nil {
				logFields = append(logFields, zap.Error(err))
			}
			logger.Warn("Request failed - Client error", logFields...)
		} else {
			// Success (2xx, 3xx)
			logger.Info("Request processed", logFields...)
		}

		return err
	}
}
