package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinal/core/internal/api/auth"
	"github.com/sentinal/core/internal/api/transaction"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/version"
)

// Server holds the Fiber app and configurations
type Server struct {
	App                *fiber.App
	Config             *config.Config
	DB                 *pgxpool.Pool
	UoW                repositories.UnitOfWork
	TransactionHandler *transaction.Handler
	AuthHandler        *auth.Handler
	TokenService       auth.TokenService
}

// New creates a new Server instance
func New(cfg *config.Config, db *pgxpool.Pool, uow repositories.UnitOfWork, txHandler *transaction.Handler, authHandler *auth.Handler, tokenService auth.TokenService) *Server {
	app := fiber.New(fiber.Config{
		AppName:       "Sentinal Core " + version.Version,
		StrictRouting: true,
		ServerHeader:  "Sentinal",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(TraceMiddleware())          // Our custom structured logger + trace ID
	app.Use(TransactionMiddleware(uow)) // Atomic transaction per request

	s := &Server{
		App:                app,
		Config:             cfg,
		DB:                 db,
		UoW:                uow,
		TransactionHandler: txHandler,
		AuthHandler:        authHandler,
		TokenService:       tokenService,
	}

	// Register Routes
	s.RegisterRoutes()

	return s
}

// Listen starts the server on the given address
func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}
