package server

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/pkg/logger"
	"go.uber.org/zap"
)

const UoWKey = "uow"

// TransactionMiddleware wraps each request in a database transaction
func TransactionMiddleware(uow repositories.UnitOfWork) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// We only want transactions for state-changing methods
		// or if specifically requested. For this project, we'll be safe
		// and wrap POST, PUT, PATCH, DELETE.
		method := c.Method()
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			// No transaction for read-only (standard path)
			// But we still attach the UoW (pool-based) for consistency
			c.Locals(UoWKey, uow)
			return c.Next()
		}

		err := uow.Do(c.Context(), func(txUow repositories.UnitOfWork) error {
			// Attach the transaction-based UoW to the request context
			c.Locals(UoWKey, txUow)

			// Execute the rest of the stack
			return c.Next()
		})

		if err != nil {
			logger.Error("Transaction middleware error", zap.Error(err))
			// Error is already handled by handler return or uow.Do
			// Fiber will pick up the error returned from c.Next() inside uow.Do
		}

		return err
	}
}

// GetUoW retrieves the UnitOfWork from fiber context
func GetUoW(c *fiber.Ctx) repositories.UnitOfWork {
	uow, ok := c.Locals(UoWKey).(repositories.UnitOfWork)
	if !ok {
		return nil
	}
	return uow
}
