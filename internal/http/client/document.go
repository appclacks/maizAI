package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Document struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created-at"`
}

type DocumentChunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document-id,omitempty"`
	Fragment   string    `json:"fragment"`
	Embedding  []float32 `json:"-"`
	CreatedAt  time.Time `json:"created-at"`
}

type ListDocumentChunksForDocumentInput struct {
	DocumentID string `param:"id" path:"id"`
}

type GetDocumentInput struct {
	ID string `param:"id" path:"id"`
}

type DeleteDocumentInput struct {
	ID string `param:"id" path:"id"`
}

type DeleteDocumentChunkInput struct {
	ID string `param:"id" path:"id"`
}

type CreateDocumentInput struct {
	Name        string `json:"name" required:"true"`
	Description string `json:"description"`
}

type EmbedDocumentInput struct {
	DocumentID string `json:"-" param:"document-id" path:"document-id"`
	Model      string `json:"model" required:"true"`
	Input      string `json:"input" required:"true"`
	Provider   string `json:"provider" required:"true"`
}

type ListDocumentsOutput struct {
	Documents []Document `json:"documents"`
}

type ListDocumentChunksOutput struct {
	Chunks []DocumentChunk `json:"chunks"`
}

type RagSearchQuery struct {
	Input    string `json:"input" required:"true"`
	Model    string `json:"model" required:"true"`
	Provider string `json:"provider" required:"true"`
	Limit    int32  `json:"limit" required:"true"`
}

func (c *Client) ListDocuments(ctx context.Context) (*ListDocumentsOutput, error) {
	var result ListDocumentsOutput
	_, err := c.sendRequest(ctx, "/api/v1/document", http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateDocument(ctx context.Context, input CreateDocumentInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, "/api/v1/document", http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) EmbedDocument(ctx context.Context, input EmbedDocumentInput) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/document/%s", input.DocumentID), http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDocument(ctx context.Context, id string) (*Document, error) {
	var result Document
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/document/%s", id), http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDocument(ctx context.Context, id string) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/document/%s", id), http.MethodDelete, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDocumentChunk(ctx context.Context, id string) (*Response, error) {
	var result Response
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/document-chunk/%s", id), http.MethodDelete, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) MatchChunk(ctx context.Context, input RagSearchQuery) (*ListDocumentChunksOutput, error) {
	var result ListDocumentChunksOutput
	_, err := c.sendRequest(ctx, "/api/v1/document-chunk", http.MethodPut, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListDocumentsChunkForDocument(ctx context.Context, id string) (*ListDocumentChunksOutput, error) {
	var result ListDocumentChunksOutput
	_, err := c.sendRequest(ctx, fmt.Sprintf("/api/v1/document/%s/chunks", id), http.MethodGet, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
