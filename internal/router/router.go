package router

import (
	"net/http"

	"portfolio-backend/internal/config"
	"portfolio-backend/internal/handler"
	"portfolio-backend/internal/middleware"
	"portfolio-backend/internal/repository"
)

// New creates and configures the HTTP router with all routes and middlewares
func New(
	cfg *config.Config,
	projectRepo repository.ProjectRepository,
	contactRepo repository.ContactRepository,
) http.Handler {
	mux := http.NewServeMux()

	healthH := handler.NewHealthHandler(cfg)
	projectH := handler.NewProjectHandler(projectRepo)
	contactH := handler.NewContactHandler(contactRepo)

	// Root index endpoint
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"message": "Welcome to Portfolio REST API",
			"endpoints": {
				"health": "GET /api/v1/health",
				"projects": "GET /api/v1/projects",
				"project_by_id": "GET /api/v1/projects/{id}",
				"create_project": "POST /api/v1/projects",
				"update_project": "PUT /api/v1/projects/{id}",
				"delete_project": "DELETE /api/v1/projects/{id}",
				"submit_contact": "POST /api/v1/contact",
				"list_contact": "GET /api/v1/contact"
			}
		}`))
	})

	// API v1 routes
	mux.HandleFunc("GET /api/v1/health", healthH.Check)

	// Projects
	mux.HandleFunc("GET /api/v1/projects", projectH.List)

	// Contact messages
	mux.HandleFunc("POST /api/v1/contact", contactH.Create)
	mux.HandleFunc("GET /api/v1/contact", contactH.List)

	// Wrap mux with global middleware chain
	var wrappedHandler http.Handler = mux
	wrappedHandler = middleware.CORS(cfg)(wrappedHandler)
	wrappedHandler = middleware.Recovery(wrappedHandler)
	wrappedHandler = middleware.Logger(wrappedHandler)

	return wrappedHandler
}
