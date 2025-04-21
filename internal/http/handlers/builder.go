package handlers

import (
	"context"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	rag "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
)

type Assistant interface {
	Pipeline(ctx context.Context, options aggregates.QueryOptions, contextOptions shared.ContextOptions, context string, messages []shared.Message) (*aggregates.Answer, error)
	StreamPipeline(ctx context.Context, options aggregates.QueryOptions, contextOptions shared.ContextOptions, contextID string, messages []shared.Message) (<-chan aggregates.Event, error)
}

type ContextManager interface {
	ListContexts(ctx context.Context) ([]shared.ContextMetadata, error)
	CreateContext(ctx context.Context, context shared.Context) error
	GetContext(ctx context.Context, id string) (*shared.Context, error)
	DeleteContext(ctx context.Context, id string) error
	AddMessagesToContext(ctx context.Context, id string, messages []shared.Message) error
	DeleteContextMessage(ctx context.Context, id string) error
	UpdateContextMessage(ctx context.Context, messageID string, role string, content string) error
	DeleteContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error
	CreateContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error
	DeleteContextMessages(ctx context.Context, contextID string) error
}

type Rag interface {
	GetDocument(ctx context.Context, id string) (*rag.Document, error)
	DeleteDocument(ctx context.Context, id string) error
	DeleteDocumentChunk(ctx context.Context, id string) error
	CreateDocument(ctx context.Context, document rag.Document) error
	ListDocuments(ctx context.Context) ([]rag.Document, error)
	Embed(ctx context.Context, docID string, query rag.EmbeddingQuery) error
	Match(ctx context.Context, query rag.SearchQuery) ([]rag.DocumentChunk, error)
	ListDocumentChunksForDocument(ctx context.Context, id string) ([]rag.DocumentChunk, error)
}

func newResponse(messages ...string) client.Response {
	return client.Response{
		Messages: messages,
	}
}

type Builder struct {
	assistant  Assistant
	ctxManager ContextManager
	ragManager Rag
}

func NewBuilder(assistant Assistant, ctxManager ContextManager, ragManager Rag) *Builder {
	return &Builder{
		assistant:  assistant,
		ctxManager: ctxManager,
		ragManager: ragManager,
	}
}
