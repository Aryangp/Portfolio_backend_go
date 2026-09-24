package handler

import (
	"net/http"
	"portfolio-backend/internal/repository"
	"strconv"
)

// ProjectHandler handles HTTP requests related to projects
type ProjectHandler struct {
	repo repository.ProjectRepository
}

// NewProjectHandler returns a new ProjectHandler
func NewProjectHandler(repo repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

// List handles GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")

	var featured *bool
	if featuredStr := r.URL.Query().Get("featured"); featuredStr != "" {
		if val, err := strconv.ParseBool(featuredStr); err == nil {
			featured = &val
		}
	}

	projects, err := h.repo.GetAll(category, tag, featured)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	writeSuccess(w, http.StatusOK, "Projects retrieved successfully", projects)
}
