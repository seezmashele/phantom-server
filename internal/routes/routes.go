package routes

import (
	"net/http"

	"github.com/rs/cors"
	"phantom-server/internal/config"
	"phantom-server/internal/handlers"
	"phantom-server/internal/middleware"
)

// Router manages HTTP routes and middleware integration
type Router struct {
	mux     *http.ServeMux
	handler *handlers.Handler
}

// NewRouter creates a new Router instance with handler dependency
func NewRouter(handler *handlers.Handler) *Router {
	return &Router{
		mux:     http.NewServeMux(),
		handler: handler,
	}
}

// SetupRoutes configures all routes with middleware and returns the final handler
func (r *Router) SetupRoutes(cfg *config.Config) http.Handler {
	// Register specific routes
	r.mux.HandleFunc("/", r.handler.Home)
	r.mux.HandleFunc("/health", r.handler.Health)

	// Create a wrapper that handles 404s for unregistered routes
	routeHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// For the root path, serve it directly
		if req.URL.Path == "/" {
			r.handler.Home(w, req)
			return
		}
		// For health path, serve it directly
		if req.URL.Path == "/health" {
			r.handler.Health(w, req)
			return
		}
		// For all other paths, return 404
		r.handler.NotFound(w, req)
	})

	// Setup CORS middleware
	corsHandler := r.setupCORS(cfg)

	// Create middleware chain: Logger -> CORS -> Routes
	middlewareChain := middleware.Chain(
		middleware.Logger(cfg.Server.EnableLogging),
	)

	// Apply middleware chain to the route handler, then wrap with CORS
	return corsHandler.Handler(middlewareChain(routeHandler))
}

// setupCORS configures CORS using rs/cors package with config options
func (r *Router) setupCORS(cfg *config.Config) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   cfg.Server.AllowedMethods,
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
}
