package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"portfolio-backend/internal/model"
)

type mockChatRepo struct {
	chunks []string
}

func (m *mockChatRepo) StreamMessage(ctx context.Context, history []model.ChatMessage, prompt string, onChunk func(chunk string) error) error {
	for _, chunk := range m.chunks {
		if err := onChunk(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockChatRepo) Close() error {
	return nil
}

func TestChatHandler_Stream(t *testing.T) {
	mockRepo := &mockChatRepo{chunks: []string{"Hello ", "world!"}}
	h := NewChatHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/stream?message=hello", nil)
	rec := httptest.NewRecorder()

	h.Stream(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hello ") || !strings.Contains(body, "world!") {
		t.Errorf("expected streamed chunks in body, got %s", body)
	}
	if !strings.Contains(body, "[DONE]") {
		t.Errorf("expected [DONE] delimiter in body, got %s", body)
	}
}

func TestChatHandler_OfflineFallback(t *testing.T) {
	h := NewChatHandler(nil) // nil repo simulates missing GEMINI_API_KEY

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/stream?message=hello", nil)
	rec := httptest.NewRecorder()

	h.Stream(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "offline") {
		t.Errorf("expected offline notification in body, got %s", body)
	}
}
