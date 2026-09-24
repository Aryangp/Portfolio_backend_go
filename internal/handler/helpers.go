package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"portfolio-backend/internal/model"
)

// writeJSON marshals data and sends it with the specified HTTP status code
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"success":false,"error":"Failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

// writeSuccess sends a standard success response envelope
func writeSuccess(w http.ResponseWriter, status int, message string, data any) {
	writeJSON(w, status, model.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// writeError sends a standard error response envelope
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// readJSON decodes the request body into target struct with max size limit
func readJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body cannot be empty")
	}
	defer r.Body.Close()

	// Max 1MB body limit to prevent memory exhaustion
	r.Body = http.MaxBytesReader(nil, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(target); err != nil {
		return err
	}

	// Ensure there is only a single JSON object in the body
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}
