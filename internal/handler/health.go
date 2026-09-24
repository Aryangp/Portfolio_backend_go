package handler

import (
	"net/http"
	"time"

	"portfolio-backend/internal/config"
)

// HealthHandler handles health-check requests
type HealthHandler struct {
	cfg       *config.Config
	startTime time.Time
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startTime: time.Now().UTC(),
	}
}

// Check returns server health status, uptime, and metadata
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime).Round(time.Second).String()

	data := map[string]any{
		"status":      "healthy",
		"version":     "1.0.0",
		"environment": h.cfg.AppEnv,
		"uptime":      uptime,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	writeSuccess(w, http.StatusOK, "Portfolio Backend API is running smoothly", data)
}
