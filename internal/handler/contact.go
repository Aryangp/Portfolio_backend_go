package handler

import (
	"net/http"

	"portfolio-backend/internal/model"
	"portfolio-backend/internal/repository"
)

// ContactHandler handles contact form messages
type ContactHandler struct {
	repo repository.ContactRepository
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(repo repository.ContactRepository) *ContactHandler {
	return &ContactHandler{repo: repo}
}

// Create handles POST /api/v1/contact
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateContactRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	msg, err := h.repo.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to submit contact message")
		return
	}

	writeSuccess(w, http.StatusCreated, "Thank you for reaching out! Your message has been received.", msg)
}

// List handles GET /api/v1/contact
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	messages, err := h.repo.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	writeSuccess(w, http.StatusOK, "Contact messages retrieved successfully", messages)
}
