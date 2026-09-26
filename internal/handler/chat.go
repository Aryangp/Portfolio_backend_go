package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"portfolio-backend/internal/model"
	"portfolio-backend/internal/repository"
)

// ChatHandler handles real-time SSE AI chatbot streaming requests
type ChatHandler struct {
	chatRepo repository.ChatRepository
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(chatRepo repository.ChatRepository) *ChatHandler {
	return &ChatHandler{chatRepo: chatRepo}
}

// Stream handles GET and POST /api/v1/chat/stream Server-Sent Events
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)

	// Set Server-Sent Events (SSE) headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_ = rc.Flush()

	// Extract message and history
	var message string
	var history []model.ChatMessage

	if r.Method == http.MethodPost {
		var req model.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			message = strings.TrimSpace(req.Message)
			history = req.History
		}
	}

	if message == "" {
		message = strings.TrimSpace(r.URL.Query().Get("message"))
	}

	if message == "" {
		_ = emitSSE(w, rc, model.ChatStreamEvent{Error: "Message parameter is required", Done: true})
		fmt.Fprintf(w, "data: [DONE]\n\n")
		_ = rc.Flush()
		return
	}

	// If Gemini client is not initialized (missing API key), provide a friendly fallback
	if h.chatRepo == nil {
		_ = emitSSE(w, rc, model.ChatStreamEvent{
			Chunk: "👋 Hi! The AI Chatbot is currently offline because no GEMINI_API_KEY was provided in the server .env. You can still explore all of Aryan's projects on the site!",
		})
		_ = emitSSE(w, rc, model.ChatStreamEvent{Done: true})
		fmt.Fprintf(w, "data: [DONE]\n\n")
		_ = rc.Flush()
		return
	}

	log.Printf("[Chat] Streaming prompt: %q (history: %d turns)", message, len(history))

	// Stream tokens from Gemini
	chunkCount := 0
	err := h.chatRepo.StreamMessage(r.Context(), history, message, func(chunk string) error {
		chunkCount++
		return emitSSE(w, rc, model.ChatStreamEvent{Chunk: chunk})
	})

	if err != nil {
		log.Printf("❌ [Chat] Gemini stream error: %v", err)
		_ = emitSSE(w, rc, model.ChatStreamEvent{Error: err.Error(), Done: true})
	} else {
		log.Printf("✅ [Chat] Stream completed successfully (%d chunks sent)", chunkCount)
		_ = emitSSE(w, rc, model.ChatStreamEvent{Done: true})
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	_ = rc.Flush()
}

func emitSSE(w http.ResponseWriter, rc *http.ResponseController, event model.ChatStreamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return rc.Flush()
}
