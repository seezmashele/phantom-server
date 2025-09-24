package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Home(t *testing.T) {
	// Create a new handler
	handler := NewHandler()
	
	// Create a request to the home endpoint
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler.Home(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	// Check the content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ct, expected)
	}
	
	// Parse the response body
	var response Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response JSON: %v", err)
	}
	
	// Check the response structure
	if response.Status != "success" {
		t.Errorf("expected status 'success', got %v", response.Status)
	}
	
	if response.Message != "Welcome to the HTTP server!" {
		t.Errorf("expected welcome message, got %v", response.Message)
	}
	
	// Check that data is present
	if response.Data == nil {
		t.Error("expected data field to be present")
	}
}

func TestHandler_Health(t *testing.T) {
	// Create a new handler
	handler := NewHandler()
	
	// Create a request to the health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler.Health(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	// Check the content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ct, expected)
	}
	
	// Parse the response body
	var response Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response JSON: %v", err)
	}
	
	// Check the response structure
	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response.Status)
	}
	
	if response.Message != "Server is running" {
		t.Errorf("expected health message, got %v", response.Message)
	}
	
	// Check that data is present
	if response.Data == nil {
		t.Error("expected data field to be present")
	}
}

func TestHandler_NotFound(t *testing.T) {
	// Create a new handler
	handler := NewHandler()
	
	// Create a request to a non-existent endpoint
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler.NotFound(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
	
	// Check the content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ct, expected)
	}
	
	// Parse the response body
	var response Response
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse response JSON: %v", err)
	}
	
	// Check the response structure
	if response.Status != "error" {
		t.Errorf("expected status 'error', got %v", response.Status)
	}
	
	if response.Message != "The requested resource was not found" {
		t.Errorf("expected not found message, got %v", response.Message)
	}
	
	// Check that data is present and contains path info
	if response.Data == nil {
		t.Error("expected data field to be present")
	}
}