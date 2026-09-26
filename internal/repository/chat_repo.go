package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"portfolio-backend/internal/model"
	"portfolio-backend/internal/prompts"
	"portfolio-backend/internal/tools"
)

// ChatRepository defines the interface for streaming AI chatbot interactions
type ChatRepository interface {
	StreamMessage(ctx context.Context, history []model.ChatMessage, prompt string, onChunk func(chunk string) error) error
	Close() error
}

// GeminiChatRepository streams responses directly from Google Gemini API with dynamic function calling
type GeminiChatRepository struct {
	apiKey       string
	modelName    string
	httpClient   *http.Client
	toolRegistry *tools.Registry
}

// NewGeminiChatRepository initializes the Gemini client with API key, model name, and optional tool registry
func NewGeminiChatRepository(_ context.Context, apiKey, modelName string, toolRegistry *tools.Registry) (*GeminiChatRepository, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = "gemini-1.5-flash"
	}

	return &GeminiChatRepository{
		apiKey:       apiKey,
		modelName:    strings.TrimSpace(modelName),
		toolRegistry: toolRegistry,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Gemini API JSON structures
type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type geminiContent struct {
	Role  string            `json:"role,omitempty"`
	Parts []json.RawMessage `json:"parts"`
}

type geminiToolDeclaration struct {
	FunctionDeclarations []tools.FunctionDeclaration `json:"functionDeclarations"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	Contents          []geminiContent         `json:"contents"`
	Tools             []geminiToolDeclaration `json:"tools,omitempty"`
	GenerationConfig  struct {
		Temperature float64 `json:"temperature"`
	} `json:"generationConfig"`
}

type geminiCandidate struct {
	Content struct {
		Role  string            `json:"role"`
		Parts []json.RawMessage `json:"parts"`
	} `json:"content"`
	FinishReason string `json:"finishReason,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// StreamMessage handles conversation turns, tool calls, and real-time SSE streaming from Gemini
func (r *GeminiChatRepository) StreamMessage(ctx context.Context, history []model.ChatMessage, prompt string, onChunk func(chunk string) error) error {
	var contents []geminiContent

	// 1. Reconstruct conversation history
	for _, h := range history {
		role := "user"
		if strings.EqualFold(h.Role, "model") || strings.EqualFold(h.Role, "assistant") {
			role = "model"
		}
		c, err := makeTextContent(role, h.Content)
		if err != nil {
			return err
		}
		contents = append(contents, c)
	}

	// 2. Add current user prompt
	userContent, err := makeTextContent("user", prompt)
	if err != nil {
		return err
	}
	contents = append(contents, userContent)

	systemPrompt := prompts.NewChatBotPrompt().GetSystemPrompt()

	// 3. If tools are available, run multi-turn function call loop
	if r.toolRegistry != nil && r.toolRegistry.HasTools() {
		const maxToolTurns = 5
		for turn := 0; turn < maxToolTurns; turn++ {
			modelContent, fnCall, err := r.evaluateToolCall(ctx, contents, systemPrompt)
			if err != nil {
				return err
			}

			// If no tool was requested by Gemini, break out to stream final response
			if fnCall == nil {
				break
			}

			// Execute tool dynamically
			toolResult, toolErr := r.toolRegistry.Execute(ctx, fnCall.Name, fnCall.Args)
			var respPayload any
			if toolErr != nil {
				respPayload = map[string]any{"error": toolErr.Error()}
			} else {
				respPayload = toolResult
			}

			// Append raw model turn preserving exact parts (including thought_signature)
			contents = append(contents, *modelContent)

			// Append function response turn with Role 'user' (required by Gemini API)
			fnRespContent, err := makeFunctionResponseContent(fnCall.Name, respPayload)
			if err != nil {
				return fmt.Errorf("failed to marshal function response: %w", err)
			}
			contents = append(contents, fnRespContent)
		}
	}

	// 4. Stream final response via SSE
	return r.streamFinalResponse(ctx, contents, systemPrompt, onChunk)
}

// evaluateToolCall sends a non-streaming generateContent request to check if Gemini invokes a tool
func (r *GeminiChatRepository) evaluateToolCall(ctx context.Context, contents []geminiContent, systemPrompt string) (*geminiContent, *geminiFunctionCall, error) {
	sysContent, err := makeTextContent("", systemPrompt)
	if err != nil {
		return nil, nil, err
	}

	reqBody := geminiRequest{
		SystemInstruction: &sysContent,
		Contents:          contents,
		Tools: []geminiToolDeclaration{
			{FunctionDeclarations: r.toolRegistry.GetDeclarations()},
		},
	}
	reqBody.GenerationConfig.Temperature = 0.7

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", r.modelName, r.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call gemini api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, nil, fmt.Errorf("gemini error: %s", geminiResp.Error.Message)
	}

	for _, cand := range geminiResp.Candidates {
		for _, rawPart := range cand.Content.Parts {
			var partCheck struct {
				FunctionCall *geminiFunctionCall `json:"functionCall"`
			}
			if err := json.Unmarshal(rawPart, &partCheck); err == nil && partCheck.FunctionCall != nil {
				modelContent := &geminiContent{
					Role:  cand.Content.Role,
					Parts: cand.Content.Parts,
				}
				if modelContent.Role == "" {
					modelContent.Role = "model"
				}
				return modelContent, partCheck.FunctionCall, nil
			}
		}
	}

	return nil, nil, nil
}

// streamFinalResponse streams text tokens via Server-Sent Events (SSE)
func (r *GeminiChatRepository) streamFinalResponse(ctx context.Context, contents []geminiContent, systemPrompt string, onChunk func(chunk string) error) error {
	sysContent, err := makeTextContent("", systemPrompt)
	if err != nil {
		return err
	}

	reqBody := geminiRequest{
		SystemInstruction: &sysContent,
		Contents:          contents,
	}

	if r.toolRegistry != nil && r.toolRegistry.HasTools() {
		reqBody.Tools = []geminiToolDeclaration{
			{FunctionDeclarations: r.toolRegistry.GetDeclarations()},
		}
	}

	reqBody.GenerationConfig.Temperature = 0.7

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal stream request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", r.modelName, r.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create stream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to gemini stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini stream error (status %d): %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		jsonData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if jsonData == "" || jsonData == "[DONE]" {
			continue
		}

		var streamResp geminiResponse
		if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
			continue
		}

		if streamResp.Error != nil {
			return fmt.Errorf("gemini error: %s", streamResp.Error.Message)
		}

		for _, cand := range streamResp.Candidates {
			for _, rawPart := range cand.Content.Parts {
				var partText struct {
					Text string `json:"text"`
				}
				if err := json.Unmarshal(rawPart, &partText); err == nil && partText.Text != "" {
					if err := onChunk(partText.Text); err != nil {
						return err
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// Helpers for serializing gemini content parts
func makeTextContent(role, text string) (geminiContent, error) {
	partBytes, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return geminiContent{}, err
	}
	return geminiContent{
		Role:  role,
		Parts: []json.RawMessage{partBytes},
	}, nil
}

func makeFunctionResponseContent(name string, result any) (geminiContent, error) {
	part := map[string]any{
		"functionResponse": map[string]any{
			"name": name,
			"response": map[string]any{
				"result": result,
			},
		},
	}
	partBytes, err := json.Marshal(part)
	if err != nil {
		return geminiContent{}, err
	}
	return geminiContent{
		Role:  "user",
		Parts: []json.RawMessage{partBytes},
	}, nil
}

// Close closes any open repository resources
func (r *GeminiChatRepository) Close() error {
	return nil
}
