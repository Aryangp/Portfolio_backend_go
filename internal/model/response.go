package model

// APIResponse represents a standard success response envelope
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse represents a standard error response envelope
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
