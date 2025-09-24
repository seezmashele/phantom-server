package middleware

import (
	"log"
	"net/http"
	"time"
)

// Middleware represents a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middleware functions into a single middleware
// Middleware are applied in the order they are provided (left to right)
// The first middleware in the chain will be the outermost wrapper
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		// Apply middleware in reverse order so the first middleware
		// becomes the outermost wrapper
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Logger creates a middleware that logs HTTP requests
// It logs the request method, path, and timestamp for each request
// The enabled parameter allows configurable logging enable/disable functionality
func Logger(enabled bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled {
				start := time.Now()
				log.Printf("[%s] %s %s", 
					start.Format("2006-01-02 15:04:05"), 
					r.Method, 
					r.URL.Path)
			}
			next.ServeHTTP(w, r)
		})
	}
}