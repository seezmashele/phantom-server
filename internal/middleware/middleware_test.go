package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestChain(t *testing.T) {
	t.Run("no middleware", func(t *testing.T) {
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("final"))
		})

		chained := Chain()(finalHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		if w.Body.String() != "final" {
			t.Errorf("Expected 'final', got %s", w.Body.String())
		}
	})

	t.Run("single middleware", func(t *testing.T) {
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("final"))
		})

		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("m1-"))
				next.ServeHTTP(w, r)
			})
		}

		chained := Chain(middleware1)(finalHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		expected := "m1-final"
		if w.Body.String() != expected {
			t.Errorf("Expected %s, got %s", expected, w.Body.String())
		}
	})

	t.Run("multiple middleware", func(t *testing.T) {
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("final"))
		})

		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("m1-"))
				next.ServeHTTP(w, r)
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("m2-"))
				next.ServeHTTP(w, r)
			})
		}

		chained := Chain(middleware1, middleware2)(finalHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		expected := "m1-m2-final"
		if w.Body.String() != expected {
			t.Errorf("Expected %s, got %s", expected, w.Body.String())
		}
	})
}

func TestChainOrder(t *testing.T) {
	var executionOrder []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m2-after")
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "final")
	})

	chained := Chain(middleware1, middleware2)(finalHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	chained.ServeHTTP(w, req)

	expected := []string{"m1-before", "m2-before", "final", "m2-after", "m1-after"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("Expected %d steps, got %d", len(expected), len(executionOrder))
	}

	for i, step := range expected {
		if executionOrder[i] != step {
			t.Errorf("Step %d: expected %s, got %s", i, step, executionOrder[i])
		}
	}
}

func TestLogger(t *testing.T) {
	t.Run("logging enabled", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		handler := Logger(true)(testHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Body.String() != "OK" {
			t.Errorf("Expected 'OK', got %s", w.Body.String())
		}

		logOutput := buf.String()
		if logOutput == "" {
			t.Error("Expected log output but got none")
		}
		if !strings.Contains(logOutput, "GET") {
			t.Errorf("Expected log to contain 'GET', got: %s", logOutput)
		}
		if !strings.Contains(logOutput, "/test") {
			t.Errorf("Expected log to contain '/test', got: %s", logOutput)
		}
	})

	t.Run("logging disabled", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		handler := Logger(false)(testHandler)
		req := httptest.NewRequest("POST", "/api/users", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Body.String() != "OK" {
			t.Errorf("Expected 'OK', got %s", w.Body.String())
		}

		logOutput := buf.String()
		if logOutput != "" {
			t.Errorf("Expected no log output but got: %s", logOutput)
		}
	})
}

func TestLoggerWithChain(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	headerMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chained"))
	})

	chained := Chain(Logger(true), headerMiddleware)(testHandler)
	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()
	chained.ServeHTTP(w, req)

	if w.Body.String() != "chained" {
		t.Errorf("Expected 'chained', got %s", w.Body.String())
	}

	if w.Header().Get("X-Test") != "middleware" {
		t.Errorf("Expected X-Test header 'middleware', got %s", w.Header().Get("X-Test"))
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "POST") {
		t.Errorf("Expected log to contain 'POST', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/api/test") {
		t.Errorf("Expected log to contain '/api/test', got: %s", logOutput)
	}
}
