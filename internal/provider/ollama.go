package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllamaProvider(baseURL string) *OllamaProvider {
	baseURL = strings.TrimRight(baseURL, "/")

	return &OllamaProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OllamaProvider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("build test request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (p *OllamaProvider) Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error) {
	payload := ollamaChatRequest{
		Model:  req.Model,
		Stream: false,
		Messages: append([]ollamaMessage{
			{
				Role:    "system",
				Content: req.SystemPrompt,
			},
		}, toOllamaMessages(req.Messages)...),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("send ollama request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("read ollama response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ReviewResponse{}, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed ollamaChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ReviewResponse{}, fmt.Errorf("parse ollama response: %w", err)
	}

	return ReviewResponse{
		Content: parsed.Message.Content,
		Raw:     string(respBody),
	}, nil
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Model   string        `json:"model"`
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

func toOllamaMessages(messages []Message) []ollamaMessage {
	out := make([]ollamaMessage, 0, len(messages))
	for _, msg := range messages {
		out = append(out, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return out
}
