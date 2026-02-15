package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sentinal/core/internal/api/auth"
	"github.com/sentinal/core/internal/api/transaction"
	"github.com/sentinal/core/internal/database"
	"github.com/sentinal/core/internal/repository"
	"github.com/sentinal/core/internal/server"
	"github.com/sentinal/core/pkg/config"
	"github.com/sentinal/core/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// 2. Initialize Logger
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	defer func() {
		_ = logger.Log.Sync()
	}()

	// 3. Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid configuration", zap.Error(err))
	}

	// 4. Initialize Database
	dbPool, err := database.Init(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// 5. Initialize Redis
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr: cfg.RedisAddr,
		})
		// Check connection (optional, we could also do this lazily)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis connection failed, rate limiting will be in-memory", zap.Error(err))
		} else {
			logger.Info("Connected to Redis", zap.String("addr", cfg.RedisAddr))
		}
	}

	// 6. Initialize Unit of Work
	uow := repository.NewUnitOfWork(dbPool)

	// 6. Initialize Services & Handlers
	txService := transaction.NewService(uow)
	txHandler := transaction.NewHandler(txService)

	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiry)
	authService := auth.NewService(uow, tokenService, cfg)
	authHandler := auth.NewHandler(authService)

	// 8. Create server
	srv := server.New(cfg, dbPool, redisClient, uow, txHandler, authHandler, tokenService)

	// 9. Graceful shutdown coordination
	shutdownComplete := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutting down server gracefully...")

		// TODO(SC-020): Cleanup other resources here

		// Set a timeout for shutdown process to prevent hanging
		if err := srv.App.ShutdownWithTimeout(10 * time.Second); err != nil {
			logger.Error("Server shutdown failed or timed out", zap.Error(err))
		}

		if redisClient != nil {
			_ = redisClient.Close()
		}

		close(shutdownComplete)
	}()

	// 7. Start server
	logger.Info("Starting Sentinal Core API",
		zap.String("port", cfg.Port),
		zap.String("env", cfg.AppEnv),
	)

	if err := srv.Listen(":" + cfg.Port); err != nil {
		// Listen returns error when server is closed, this is normal
		logger.Info("Server listener closed", zap.Error(err))
	}

	// Wait for shutdown goroutine to finish (cleanup, etc.)
	<-shutdownComplete
	logger.Info("Server exit complete. Goodbye!")
}
