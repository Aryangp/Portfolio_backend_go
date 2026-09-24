package repository

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"portfolio-backend/internal/model"
)

// ContactRepository defines the interface for contact messages data access
type ContactRepository interface {
	Create(req model.CreateContactRequest) (*model.ContactMessage, error)
	GetAll() ([]model.ContactMessage, error)
}

// InMemoryContactRepository is a thread-safe in-memory store for contact form messages
type InMemoryContactRepository struct {
	mu       sync.RWMutex
	messages map[string]model.ContactMessage
	nextID   int
}

// NewInMemoryContactRepository creates a new in-memory contact repository
func NewInMemoryContactRepository() *InMemoryContactRepository {
	return &InMemoryContactRepository{
		messages: make(map[string]model.ContactMessage),
		nextID:   1,
	}
}

// Create stores a new incoming contact message
func (r *InMemoryContactRepository) Create(req model.CreateContactRequest) (*model.ContactMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("msg_%d", r.nextID)
	r.nextID++

	msg := model.ContactMessage{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Subject:   req.Subject,
		Message:   req.Message,
		CreatedAt: time.Now().UTC(),
	}

	r.messages[id] = msg
	return &msg, nil
}

// GetAll returns all saved contact messages ordered chronologically (newest first)
func (r *InMemoryContactRepository) GetAll() ([]model.ContactMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.ContactMessage
	for _, m := range r.messages {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	if result == nil {
		result = []model.ContactMessage{}
	}
	return result, nil
}
