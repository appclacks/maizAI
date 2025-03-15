package client

import (
	"context"
	"net/http"
)

type QueryOptions struct {
	Model       string         `json:"model"`
	System      string         `json:"system"`
	Temperature float64        `json:"temperature"`
	MaxTokens   uint64         `json:"max-tokens"`
	Provider    string         `json:"provider"`
	RagQuery    RagSearchQuery `json:"rag,omitempty"`
}

type ContextOptions struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Sources     ContextSources `json:"sources"`
}

type CreateConversationInput struct {
	QueryOptions      QueryOptions   `json:"query-options"`
	Prompt            string         `json:"prompt" required:"true"`
	ContextID         string         `json:"context-id,omitempty"`
	NewContextOptions ContextOptions `json:"new-context"`
}

type Result struct {
	Text string `json:"text"`
}

type ConversationAnswer struct {
	Results      []Result `json:"result"`
	InputTokens  uint64   `json:"input-tokens"`
	OutputTokens uint64   `json:"output-tokens"`
	Context      string   `json:"context"`
}

func (c *Client) CreateConversation(ctx context.Context, input CreateConversationInput) (*ConversationAnswer, error) {
	var result ConversationAnswer
	_, err := c.sendRequest(ctx, "/api/v1/conversation", http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
