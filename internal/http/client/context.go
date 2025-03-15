package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/appclacks/maizai/pkg/shared"
)

type Context struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Sources     ContextSources `json:"sources"`
	Messages    []Message      `json:"messages,omitempty"`
	CreatedAt   time.Time      `json:"created-at"`
}

type ContextMetadata struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created-at"`
	Sources     ContextSources `json:"sources"`
}

type ContextSources struct {
	Contexts []string `json:"contexts,omitempty"`
}

type GetContextInput struct {
	ID string `param:"id" path:"id"`
}

type DeleteContextInput struct {
	ID string `param:"id" path:"id"`
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created-at"`
}

type CreateContextMessage struct {
	Role    string `json:"role" required:"true"`
	Content string `json:"content" required:"true"`
}

type AddMessagesToContextInput struct {
	ID       string `json:"-" param:"id" path:"id"`
	Messages []CreateContextMessage
}

type UpdateContextMessageInput struct {
	ID      string `json:"-" param:"id" path:"id"`
	Role    string `json:"role" required:"true"`
	Content string `json:"content" required:"true"`
}

type DeleteContextMessageInput struct {
	ID string `json:"-" param:"id" path:"id"`
}

type CreateContextInput struct {
	Name        string                 `json:"name" required:"true"`
	Description string                 `json:"description"`
	Sources     shared.ContextSources  `json:"sources"`
	Messages    []CreateContextMessage `json:"messages"`
}

type DeleteContextSourceContextInput struct {
	ID              string `json:"-" param:"id" path:"id"`
	SourceContextID string `json:"-" param:"source-context-id" path:"source-context-id"`
}

type CreateContextSourceContextInput struct {
	ID              string `json:"-" param:"id" path:"id"`
	SourceContextID string `json:"-" param:"source-context-id" path:"source-context-id"`
}

type ListContextOutput struct {
	Contexts []ContextMetadata `json:"contexts"`
}

func (c *Client) ListContexts(ctx context.Context) (*ListContextOutput, error) {
	var result ListContextOutput
	_, err := c.sendRequest(ctx, "/api/v1/context", http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetContext(ctx context.Context, id string) (*Context, error) {
	var result Context
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/context/%s", id), http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteContext(ctx context.Context, id string) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/context/%s", id), http.MethodDelete, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateContext(ctx context.Context, input CreateContextInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, "/api/v1/context", http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) AddMessagesToContext(ctx context.Context, input AddMessagesToContextInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/context/%s/message", input.ID), http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateContextMessage(ctx context.Context, input UpdateContextMessageInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/message/%s", input.ID), http.MethodPut, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteContextMessage(ctx context.Context, input DeleteContextMessageInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/message/%s", input.ID), http.MethodDelete, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteContextSourceContext(ctx context.Context, input DeleteContextSourceContextInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/context/%s/sources/context/%s", input.ID, input.SourceContextID), http.MethodDelete, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateContextSourceContext(ctx context.Context, input CreateContextSourceContextInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/context/%s/sources/context/%s", input.ID, input.SourceContextID), http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
