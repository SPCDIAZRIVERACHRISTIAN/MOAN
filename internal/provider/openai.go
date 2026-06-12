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

type OpenAIProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIProvider(baseURL, apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (p *OpenAIProvider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("build test request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to OpenAI at %s: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("OpenAI rejected the API key (status 401). Check the key with: moan config show")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("OpenAI returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (p *OpenAIProvider) Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error) {
	messages := make([]openaiMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, msg := range req.Messages {
		messages = append(messages, openaiMessage{Role: msg.Role, Content: msg.Content})
	}

	payload := openaiChatRequest{
		Model:    req.Model,
		Messages: messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal OpenAI request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build OpenAI request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("send OpenAI request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("read OpenAI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ReviewResponse{}, fmt.Errorf("OpenAI returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed openaiChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ReviewResponse{}, fmt.Errorf("parse OpenAI response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return ReviewResponse{}, fmt.Errorf("OpenAI response contained no choices")
	}

	return ReviewResponse{
		Content: parsed.Choices[0].Message.Content,
		Raw:     string(respBody),
	}, nil
}

type openaiChatRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
