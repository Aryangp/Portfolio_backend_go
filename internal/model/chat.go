package model

// ChatMessage represents a single message turn in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "model"
	Content string `json:"content"` // Text message content
}

// ChatRequest represents the payload for starting or continuing a chat session
type ChatRequest struct {
	Message string        `json:"message"`
	History []ChatMessage `json:"history,omitempty"`
}

// ChatStreamEvent represents an SSE event chunk emitted to the client
type ChatStreamEvent struct {
	Chunk string `json:"chunk,omitempty"`
	Done  bool   `json:"done,omitempty"`
	Error string `json:"error,omitempty"`
}
