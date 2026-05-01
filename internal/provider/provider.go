package provider

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ReviewRequest struct {
	Model        string
	SystemPrompt string
	Messages     []Message
}

type ReviewResponse struct {
	Content string
	Raw     string
}

type Provider interface {
	TestConnection(ctx context.Context) error
	Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error)
}
