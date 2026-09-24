package model

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// ContactMessage represents a message sent via the portfolio contact form
type ContactMessage struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateContactRequest contains payload for submitting a contact message
type CreateContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message"`
}

// Validate validates the CreateContactRequest input
func (r *CreateContactRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(r.Email)); err != nil {
		return errors.New("invalid email address format")
	}
	if strings.TrimSpace(r.Message) == "" {
		return errors.New("message is required")
	}
	if len(strings.TrimSpace(r.Message)) < 5 {
		return errors.New("message must be at least 5 characters long")
	}
	return nil
}
