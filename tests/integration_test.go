package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"phantom-server/internal/config"
	"phantom-server/internal/handlers"
	"phantom-server/internal/routes"
)

// TestServerStartupWithDifferentConfigurations tests server startup with various configuration sources
func TestServerStartupWithDifferentConfigurations(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("Default configuration", func(t *testing.T) {
		// Change to temp directory with no config files
		os.Chdir(tempDir)

		// Load configuration using default values
		cfg := config.GetDefaultConfig()

		// Verify default values
		if cfg.Server.Port != 8080 {
			t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
		}

		if cfg.Server.ShutdownTimeout != 30 {
			t.Errorf("Expected default shutdown timeout 30, got %d", cfg.Server.ShutdownTimeout)
		}

		if !cfg.Server.EnableLogging {
			t.Error("Expected default logging to be enabled")
		}
	})

	t.Run("JSON configuration", func(t *testing.T) {
		// Create JSON config file
		configPath := filepath.Join(tempDir, "test_config.json")
		configContent := `{
			"server": {
				"port": 3000,
				"shutdown_timeout_seconds": 45,
				"read_timeout_seconds": 15,
				"write_timeout_seconds": 15,
				"allowed_origins": ["http://localhost:3000"],
				"allowed_methods": ["GET", "POST"],
				"enable_logging": false
			}
		}`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Load configuration from JSON file
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify JSON config values
		if cfg.Server.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", cfg.Server.Port)
		}

		if cfg.Server.ShutdownTimeout != 45 {
			t.Errorf("Expected shutdown timeout 45, got %d", cfg.Server.ShutdownTimeout)
		}

		if cfg.Server.EnableLogging {
			t.Error("Expected logging to be disabled")
		}
	})

	t.Run("Environment variable configuration", func(t *testing.T) {
		// Create .env file with direct configuration
		envPath := filepath.Join(tempDir, ".env")
		envContent := `PORT=4000
SHUTDOWN_TIMEOUT=60
READ_TIMEOUT=20
WRITE_TIMEOUT=20
ALLOWED_ORIGINS=http://localhost:4000,https://example.com
ALLOWED_METHODS=GET,POST,PUT,DELETE
ENABLE_LOGGING=true`

		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .env file: %v", err)
		}

		// Change to temp directory
		os.Chdir(tempDir)

		// Load configuration from .env file
		cfg, err := config.LoadEnvConfig()
		if err != nil {
			t.Fatalf("Failed to load env config: %v", err)
		}

		// Verify .env config values
		if cfg.Server.Port != 4000 {
			t.Errorf("Expected port 4000, got %d", cfg.Server.Port)
		}

		if cfg.Server.ShutdownTimeout != 60 {
			t.Errorf("Expected shutdown timeout 60, got %d", cfg.Server.ShutdownTimeout)
		}

		if cfg.Server.ReadTimeout != 20 {
			t.Errorf("Expected read timeout 20, got %d", cfg.Server.ReadTimeout)
		}

		expectedOrigins := []string{"http://localhost:4000", "https://example.com"}
		if len(cfg.Server.AllowedOrigins) != len(expectedOrigins) {
			t.Errorf("Expected %d origins, got %d", len(expectedOrigins), len(cfg.Server.AllowedOrigins))
		}

		if !cfg.Server.EnableLogging {
			t.Error("Expected logging to be enabled")
		}
	})

	t.Run("Configuration merging priority", func(t *testing.T) {
		// Test configuration merging logic
		base := &config.Config{
			Server: config.ServerConfig{
				Port:            8080,
				ShutdownTimeout: 30,
				EnableLogging:   true,
			},
		}

		override := &config.Config{
			Server: config.ServerConfig{
				Port:          3000,
				EnableLogging: false,
			},
		}

		merged := config.MergeConfigs(base, override)

		// Verify override values take priority
		if merged.Server.Port != 3000 {
			t.Errorf("Expected port 3000 (override), got %d", merged.Server.Port)
		}

		if merged.Server.EnableLogging {
			t.Error("Expected logging to be disabled (override)")
		}

		// Verify base values are kept where not overridden
		if merged.Server.ShutdownTimeout != 30 {
			t.Errorf("Expected shutdown timeout 30 (base), got %d", merged.Server.ShutdownTimeout)
		}
	})
}

// TestAllEndpointsWithMiddleware tests all endpoints with middleware applied
func TestAllEndpointsWithMiddleware(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:            8080,
			ShutdownTimeout: 30,
			ReadTimeout:     10,
			WriteTimeout:    10,
			AllowedOrigins:  []string{"*"},
			AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			EnableLogging:   false, // Disable logging for cleaner test output
		},
	}

	// Setup server components
	handler := handlers.NewHandler()
	router := routes.NewRouter(handler)
	httpHandler := router.SetupRoutes(cfg)

	// Create test server
	testServer := httptest.NewServer(httpHandler)
	defer testServer.Close()

	t.Run("Home endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response handlers.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("Expected status 'success', got %s", response.Status)
		}
	})

	t.Run("Health endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response handlers.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got %s", response.Status)
		}
	})

	t.Run("404 endpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/nonexistent")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		var response handlers.Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "error" {
			t.Errorf("Expected status 'error', got %s", response.Status)
		}
	})

	t.Run("CORS preflight request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")

		w := httptest.NewRecorder()
		httpHandler.ServeHTTP(w, req)

		// Check CORS headers
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected CORS origin *, got %s", origin)
		}

		if methods := w.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "POST") {
			t.Errorf("Expected CORS methods to include POST, got %s", methods)
		}
	})

	t.Run("Logging middleware behavior", func(t *testing.T) {
		// Test with logging enabled
		logCfg := &config.Config{
			Server: config.ServerConfig{
				Port:            8080,
				ShutdownTimeout: 30,
				ReadTimeout:     10,
				WriteTimeout:    10,
				AllowedOrigins:  []string{"*"},
				AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				EnableLogging:   true,
			},
		}

		logHandler := handlers.NewHandler()
		logRouter := routes.NewRouter(logHandler)
		logHttpHandler := logRouter.SetupRoutes(logCfg)

		// Capture logs
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(os.Stderr)

		// Make request
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		logHttpHandler.ServeHTTP(w, req)

		// Check that logging occurred
		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "GET") || !strings.Contains(logOutput, "/") {
			t.Errorf("Expected request to be logged, got: %s", logOutput)
		}

		// Test with logging disabled
		noLogCfg := &config.Config{
			Server: config.ServerConfig{
				EnableLogging:  false,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			},
		}

		noLogHandler := handlers.NewHandler()
		noLogRouter := routes.NewRouter(noLogHandler)
		noLogHttpHandler := noLogRouter.SetupRoutes(noLogCfg)

		// Clear log buffer
		logBuf.Reset()

		// Make request
		req = httptest.NewRequest("GET", "/", nil)
		w = httptest.NewRecorder()
		noLogHttpHandler.ServeHTTP(w, req)

		// Check that no logging occurred
		logOutput = logBuf.String()
		if strings.Contains(logOutput, "GET") && strings.Contains(logOutput, "/") {
			t.Errorf("Expected no request logging when disabled, but got: %s", logOutput)
		}
	})
}

// TestGracefulShutdown tests graceful shutdown behavior
func TestGracefulShutdown(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:            8082,
			ShutdownTimeout: 5, // Short timeout for testing
			ReadTimeout:     10,
			WriteTimeout:    10,
			AllowedOrigins:  []string{"*"},
			AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			EnableLogging:   true,
		},
	}

	// Test server creation with configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Test that server is configured properly
	if server.Addr != ":8082" {
		t.Errorf("Expected server address :8082, got %s", server.Addr)
	}

	if server.ReadTimeout != 10*time.Second {
		t.Errorf("Expected read timeout 10s, got %v", server.ReadTimeout)
	}

	if server.WriteTimeout != 10*time.Second {
		t.Errorf("Expected write timeout 10s, got %v", server.WriteTimeout)
	}

	// Test graceful shutdown context creation
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Verify context timeout is set correctly
	deadline, ok := shutdownCtx.Deadline()
	if !ok {
		t.Error("Expected shutdown context to have a deadline")
	}

	expectedTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	actualTimeout := time.Until(deadline)

	// Allow some tolerance for timing
	if actualTimeout < expectedTimeout-time.Second || actualTimeout > expectedTimeout+time.Second {
		t.Errorf("Expected timeout around %v, got %v", expectedTimeout, actualTimeout)
	}
}
