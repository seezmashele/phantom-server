package handlers

import (
	"encoding/json"
	"net/http"

	gojson "github.com/goccy/go-json"
)

// Handler contains HTTP request handlers for different endpoints
type Handler struct {
	// Can include dependencies like database connections, services, etc.
	// For now, this is a simple struct that can be extended later
}

// NewHandler creates a new Handler instance
func NewHandler() *Handler {
	return &Handler{}
}

// Response represents a standard HTTP response structure
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Home handles the "/" endpoint and returns a welcome message
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "success",
		Message: "Welcome to the HTTP server!",
		Data: map[string]string{
			"version": "1.0.0",
			"service": "http-server",
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Health handles the "/health" endpoint and returns health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "healthy",
		Message: "Server is running",
		Data: map[string]interface{}{
			"uptime": "running",
			"status": "ok",
			"checks": map[string]string{
				"server": "healthy",
			},
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// NotFound handles undefined routes and returns a 404 error response
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "error",
		Message: "The requested resource was not found",
		Data: map[string]interface{}{
			"path":   r.URL.Path,
			"method": r.Method,
		},
	}

	h.writeJSONResponse(w, http.StatusNotFound, response)
}

// writeJSONResponse writes a JSON response using goccy/go-json
func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := gojson.NewEncoder(w).Encode(data); err != nil {
		// Fallback to standard library if goccy/go-json fails
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to encode response",
		})
	}
}
