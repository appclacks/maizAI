package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type SystemPrompt struct {
	ID          string    `json:"id" description:"The system prompt ID"`
	Name        string    `json:"name" description:"The system prompt name"`
	Description string    `json:"description,omitempty" description:"The system prompt description"`
	Content     string    `json:"content" description:"The system prompt content"`
	CreatedAt   time.Time `json:"created-at" description:"The system prompt creation date"`
}

type CreateSystemPromptInput struct {
	Name        string `json:"name" required:"true" description:"The system prompt name"`
	Description string `json:"description" description:"The system prompt description"`
	Content     string `json:"content" required:"true" description:"The system prompt content"`
}

type GetSystemPromptInput struct {
	ID string `param:"id" path:"id"`
}

type DeleteSystemPromptInput struct {
	ID string `param:"id" path:"id"`
}

type UpdateSystemPromptInput struct {
	ID      string `json:"-" param:"id" path:"id"`
	Content string `json:"content" required:"true" description:"The system prompt content"`
}

type ListSystemPromptsOutput struct {
	SystemPrompts []SystemPrompt `json:"system_prompts"`
}

func (c *Client) ListSystemPrompts(ctx context.Context) (*ListSystemPromptsOutput, error) {
	var result ListSystemPromptsOutput
	_, err := c.sendRequest(ctx, "/api/v1/system-prompt", http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSystemPrompt(ctx context.Context, id string) (*SystemPrompt, error) {
	var result SystemPrompt
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/system-prompt/%s", id), http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateSystemPrompt(ctx context.Context, input CreateSystemPromptInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, "/api/v1/system-prompt", http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateSystemPrompt(ctx context.Context, input UpdateSystemPromptInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/system-prompt/%s", input.ID), http.MethodPut, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteSystemPrompt(ctx context.Context, id string) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/system-prompt/%s", id), http.MethodDelete, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}