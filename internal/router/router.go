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
	chatRepo repository.ChatRepository,
) http.Handler {
	mux := http.NewServeMux()
	healthH := handler.NewHealthHandler(cfg)
	projectH := handler.NewProjectHandler(projectRepo)
	chatH := handler.NewChatHandler(chatRepo)

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
				"chat_stream": "GET/POST /api/v1/chat/stream"
			}
		}`))
	})

	// API v1 routes
	mux.HandleFunc("GET /api/v1/health", healthH.Check)
	// Projects
	mux.HandleFunc("GET /api/v1/projects", projectH.List)
	// AI Chatbot SSE Streaming
	mux.HandleFunc("GET /api/v1/chat/stream", chatH.Stream)
	mux.HandleFunc("POST /api/v1/chat/stream", chatH.Stream)

	// Wrap mux with global middleware chain
	var wrappedHandler http.Handler = mux
	wrappedHandler = middleware.CORS(cfg)(wrappedHandler)
	wrappedHandler = middleware.Recovery(wrappedHandler)
	wrappedHandler = middleware.Logger(wrappedHandler)

	return wrappedHandler
}
