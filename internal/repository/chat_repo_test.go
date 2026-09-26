package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"portfolio-backend/internal/model"
	"portfolio-backend/internal/tools"
)

func TestGeminiChatRepository_WithTools(t *testing.T) {
	// Mock tool
	mockTool := &mockCustomTool{
		name:        "get_github_repositories",
		description: "Fetches github repos",
		decl: tools.FunctionDeclaration{
			Name:        "get_github_repositories",
			Description: "Fetches github repos",
		},
		result: map[string]any{"repos": []string{"gpzer", "nlp-indexing"}},
	}

	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(mockTool)

	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.Contains(r.URL.Path, ":generateContent") {
			// First turn: model requests tool call
			resp := map[string]any{
				"candidates": []map[string]any{
					{
						"content": map[string]any{
							"role": "model",
							"parts": []map[string]any{
								{
									"functionCall": map[string]any{
										"name": "get_github_repositories",
										"args": map[string]any{"limit": 5},
									},
								},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, ":streamGenerateContent") {
			// Final stream response turn
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Aryan has repositories like gpzer! Pika pika ⚡\"}]}}]}\n\n"))
			return
		}
	}))
	defer server.Close()

	repo := &GeminiChatRepository{
		apiKey:       "test-key",
		modelName:    "gemini-1.5-flash",
		httpClient:   server.Client(),
		toolRegistry: toolRegistry,
	}

	// Override base url by rewriting requests in custom transport
	repo.httpClient.Transport = &rewriteTransport{
		targetBaseURL: server.URL,
		base:          http.DefaultTransport,
	}

	var chunks []string
	err := repo.StreamMessage(context.Background(), []model.ChatMessage{}, "What are Aryan's repositories?", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error streaming message: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatalf("expected chunks to be received, got 0")
	}

	combined := strings.Join(chunks, "")
	if !strings.Contains(combined, "gpzer") || !strings.Contains(combined, "Pika pika") {
		t.Errorf("expected Pikachu response with repo details, got %q", combined)
	}
}

type mockCustomTool struct {
	name        string
	description string
	decl        tools.FunctionDeclaration
	result      any
}

func (m *mockCustomTool) Name() string                     { return m.name }
func (m *mockCustomTool) Description() string              { return m.description }
func (m *mockCustomTool) Declaration() tools.FunctionDeclaration { return m.decl }
func (m *mockCustomTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	return m.result, nil
}

type rewriteTransport struct {
	targetBaseURL string
	base          http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())
	reqCopy.URL.Scheme = "http"
	serverHost := strings.TrimPrefix(t.targetBaseURL, "http://")
	reqCopy.URL.Host = serverHost
	return t.base.RoundTrip(reqCopy)
}
