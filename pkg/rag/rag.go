package rag

import (
	"context"
	"fmt"

	"github.com/appclacks/maizai/internal/id"
	"github.com/appclacks/maizai/pkg/rag/aggregates"
)

type Store interface {
	CreateDocument(ctx context.Context, document aggregates.Document) error
	GetDocument(ctx context.Context, id string) (*aggregates.Document, error)
	DeleteDocument(ctx context.Context, id string) error
	DeleteDocumentChunk(ctx context.Context, id string) error
	CreateDocumentChunk(ctx context.Context, documentChunk aggregates.DocumentChunk) error
	ListDocuments(ctx context.Context) ([]aggregates.Document, error)
	FindClosestChunks(ctx context.Context, limit int32, chunk []float32) ([]aggregates.DocumentChunk, error)
	ListDocumentChunksForDocument(ctx context.Context, docID string) ([]aggregates.DocumentChunk, error)
}

type AI interface {
	Embedding(ctx context.Context, query aggregates.EmbeddingQuery) (*aggregates.EmbeddingAnswer, error)
}

type Rag struct {
	store   Store
	clients map[string]AI
}

func New(store Store, clients map[string]AI) *Rag {
	return &Rag{
		store:   store,
		clients: clients,
	}
}

func (r *Rag) Embed(ctx context.Context, docID string, query aggregates.EmbeddingQuery) error {
	if err := id.Validate(docID, "invalid document ID"); err != nil {
		return err
	}
	client, ok := r.clients[query.Provider]
	if !ok {
		return fmt.Errorf("AI client %s not configured", query.Provider)
	}
	answer, err := client.Embedding(ctx, query)
	if err != nil {
		return err
	}
	// TODO check index
	chunk, err := aggregates.NewDocumentChunk(docID, query.Input, answer.Data[0].Embedding)
	if err != nil {
		return err
	}
	err = chunk.Validate()
	if err != nil {
		return err
	}
	err = r.store.CreateDocumentChunk(ctx, *chunk)
	if err != nil {
		return err
	}
	return nil
}

func (r *Rag) CreateDocument(ctx context.Context, document aggregates.Document) error {
	err := document.Validate()
	if err != nil {
		return err
	}
	return r.store.CreateDocument(ctx, document)
}

func (r *Rag) ListDocuments(ctx context.Context) ([]aggregates.Document, error) {
	return r.store.ListDocuments(ctx)
}

func (r *Rag) Match(ctx context.Context, query aggregates.SearchQuery) ([]aggregates.DocumentChunk, error) {
	err := query.Validate()
	if err != nil {
		return nil, err
	}
	client, ok := r.clients[query.Provider]
	if !ok {
		return nil, fmt.Errorf("AI provider %s not configured", query.Provider)
	}
	q := aggregates.EmbeddingQuery{
		Input: query.Input,
		Model: query.Model,
	}
	answer, err := client.Embedding(ctx, q)
	if err != nil {
		return nil, err
	}
	// TODO check index
	return r.store.FindClosestChunks(ctx, query.Limit, answer.Data[0].Embedding)
}

func (r *Rag) GetDocument(ctx context.Context, docID string) (*aggregates.Document, error) {
	if err := id.Validate(docID, "invalid document ID"); err != nil {
		return nil, err
	}
	return r.store.GetDocument(ctx, docID)
}

func (r *Rag) DeleteDocument(ctx context.Context, docID string) error {
	if err := id.Validate(docID, "invalid document ID"); err != nil {
		return err
	}
	return r.store.DeleteDocument(ctx, docID)
}

func (r *Rag) DeleteDocumentChunk(ctx context.Context, chunkID string) error {
	if err := id.Validate(chunkID, "invalid document chunk ID"); err != nil {
		return err
	}
	return r.store.DeleteDocumentChunk(ctx, chunkID)
}

func (r *Rag) ListDocumentChunksForDocument(ctx context.Context, docID string) ([]aggregates.DocumentChunk, error) {
	if err := id.Validate(docID, "invalid document ID"); err != nil {
		return nil, err
	}
	return r.store.ListDocumentChunksForDocument(ctx, docID)
}
