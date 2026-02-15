package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sentinal/core/internal/server"
)

func main() {
	// Create server
	srv := server.New()

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		if err := srv.App.Shutdown(); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Sentinal Core API on :%s", port)
	if err := srv.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
