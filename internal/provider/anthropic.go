package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	anthropicVersion   = "2023-06-01"
	anthropicMaxTokens = 4096
)

type AnthropicProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewAnthropicProvider(baseURL, apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (p *AnthropicProvider) setHeaders(req *http.Request) {
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")
}

func (p *AnthropicProvider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/v1/models", nil)
	if err != nil {
		return fmt.Errorf("build test request: %w", err)
	}
	p.setHeaders(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to Anthropic at %s: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Anthropic rejected the API key (status 401). Check the key with: moan config show")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("Anthropic returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (p *AnthropicProvider) Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error) {
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, anthropicMessage{Role: msg.Role, Content: msg.Content})
	}

	payload := anthropicRequest{
		Model:     req.Model,
		MaxTokens: anthropicMaxTokens,
		System:    req.SystemPrompt,
		Messages:  messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal Anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build Anthropic request: %w", err)
	}
	p.setHeaders(httpReq)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("send Anthropic request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("read Anthropic response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ReviewResponse{}, fmt.Errorf("Anthropic returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ReviewResponse{}, fmt.Errorf("parse Anthropic response: %w", err)
	}

	var content strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}

	if content.Len() == 0 {
		return ReviewResponse{}, fmt.Errorf("Anthropic response contained no text content")
	}

	return ReviewResponse{
		Content: content.String(),
		Raw:     string(respBody),
	}, nil
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}
