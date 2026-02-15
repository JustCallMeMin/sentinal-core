package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
	goredis "github.com/redis/go-redis/v9"
	"github.com/sentinal/core/pkg/config"
)

// RateLimitMiddleware returns a limiter middleware for specific routes
// Default is 5 requests per minute per IP
func RateLimitMiddleware(cfg *config.Config, redisClient *goredis.Client) fiber.Handler {
	var store fiber.Storage

	if redisClient != nil {
		store = redis.New(redis.Config{
			URL: "redis://" + cfg.RedisAddr,
		})
	}

	return limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: time.Duration(cfg.RateLimitWindowMinutes) * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests from this IP. Please try again later.",
			})
		},
		Storage: store,
	})
}
