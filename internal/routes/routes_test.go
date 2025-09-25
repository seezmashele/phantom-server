package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"phantom-server/internal/config"
	"phantom-server/internal/handlers"
)

func TestNewRouter(t *testing.T) {
	handler := handlers.NewHandler()
	router := NewRouter(handler)

	if router == nil {
		t.Fatal("NewRouter returned nil")
	}

	if router.mux == nil {
		t.Error("Router mux is nil")
	}

	if router.handler != handler {
		t.Error("Router handler not set correctly")
	}
}

func TestSetupRoutes(t *testing.T) {
	handler := handlers.NewHandler()
	router := NewRouter(handler)
	cfg := config.GetDefaultConfig()

	// Setup routes
	finalHandler := router.SetupRoutes(cfg)

	if finalHandler == nil {
		t.Fatal("SetupRoutes returned nil handler")
	}

	// Test home route
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	finalHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test health route
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	finalHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test 404 route
	req = httptest.NewRequest("GET", "/nonexistent", nil)
	w = httptest.NewRecorder()
	finalHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSetupCORS(t *testing.T) {
	handler := handlers.NewHandler()
	router := NewRouter(handler)
	cfg := config.GetDefaultConfig()

	corsHandler := router.setupCORS(cfg)

	if corsHandler == nil {
		t.Fatal("setupCORS returned nil")
	}
}
