package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"phantom-server/internal/config"
	"phantom-server/internal/handlers"
	"phantom-server/internal/routes"
)

func main() {
	// Load configuration using priority system (env > .env > json > defaults)
	cfg, err := loadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize handlers, router, and middleware
	handler := handlers.NewHandler()
	router := routes.NewRouter(handler)
	httpHandler := router.SetupRoutes(cfg)

	// Create HTTP server with configuration timeouts
	server := createServer(cfg, httpHandler)

	// Start HTTP server with graceful shutdown handling
	if err := startServerWithGracefulShutdown(server, cfg); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// loadConfiguration loads configuration with priority: env > .env > json > defaults
func loadConfiguration() (*config.Config, error) {
	// Start with default configuration
	cfg := config.GetDefaultConfig()

	// Try to load from JSON file if CONFIG_PATH is specified
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		if jsonCfg, err := config.LoadConfig(configPath); err == nil {
			cfg = config.MergeConfigs(cfg, jsonCfg)
		} else {
			log.Printf("Warning: Failed to load JSON config from %s: %v", configPath, err)
		}
	}

	// Load environment variables (including .env file)
	envCfg, err := config.LoadEnvConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	// Merge with environment configuration (highest priority)
	cfg = config.MergeConfigs(cfg, envCfg)

	return cfg, nil
}

// createServer creates an HTTP server with configuration timeouts
func createServer(cfg *config.Config, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second, // Standard idle timeout
	}

	return server
}

// startServerWithGracefulShutdown starts the server and handles graceful shutdown
func startServerWithGracefulShutdown(server *http.Server, cfg *config.Config) error {
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Starting HTTP server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErr:
		return err
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		
		// Create shutdown context with timeout
		shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		log.Println("Server shutdown completed successfully")
		return nil
	}
}
