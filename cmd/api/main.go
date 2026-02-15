package main

import (
	"os"
	"os/signal"
	"syscall"

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
	defer logger.Log.Sync()

	// 3. Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid configuration", zap.Error(err))
	}

	// 4. Create server
	srv := server.New(cfg)

	// 5. Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutting down server...")
		if err := srv.App.Shutdown(); err != nil {
			logger.Error("Server shutdown failed", zap.Error(err))
		}
	}()

	// 6. Start server
	logger.Info("Starting Sentinal Core API",
		zap.String("port", cfg.Port),
		zap.String("env", cfg.AppEnv),
	)

	if err := srv.Listen(":" + cfg.Port); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
