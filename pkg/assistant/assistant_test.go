package assistant_test

import (
	"context"
	"testing"
	"time"

	"github.com/appclacks/maizai/internal/contextstore/memory"
	mocks "github.com/appclacks/maizai/mocks/github.com/appclacks/maizai/pkg/assistant"
	"github.com/google/uuid"

	"github.com/appclacks/maizai/pkg/assistant"
	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	ct "github.com/appclacks/maizai/pkg/context"
	ragdata "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPipeline(t *testing.T) {
	rag := mocks.NewMockRag(t)
	store := memory.New()
	client := mocks.NewMockProvider(t)
	manager := ct.New(store)

	clients := make(map[string]assistant.Provider)
	clients["test"] = client
	ai := assistant.New(clients, manager, rag)

	ctx := context.Background()

	client.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(
		&aggregates.Answer{
			Results: []aggregates.Result{
				{
					Text: "this is the AI answer",
				},
			},
		}, nil)
	rag.On("Match", mock.Anything, mock.Anything).Return(
		[]ragdata.DocumentChunk{
			{
				Fragment: "fragment from rag",
			},
		},
		nil)
	queryOptions := aggregates.QueryOptions{
		Model:       "corbi-3.5",
		System:      "system prompt",
		Temperature: 1.0,
		MaxTokens:   8000,
		Provider:    "test",
		RagQuery: ragdata.SearchQuery{
			Input:    "rag input",
			Model:    "mistral-embed",
			Provider: "mistral",
			Limit:    1,
		},
	}
	contextOptions := shared.ContextOptions{Name: "foo"}
	messages := []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "message1 {ragdata}",
			CreatedAt: time.Now().UTC(),
		},
	}
	answer, err := ai.Pipeline(ctx, queryOptions, contextOptions, "", messages)
	assert.NoError(t, err)

	assert.Equal(t, "this is the AI answer", answer.Results[0].Text)

	result, err := store.GetContext(ctx, answer.Context)
	assert.NoError(t, err)
	assert.Len(t, result.Messages, 2)
	assert.Equal(t, shared.UserRole, result.Messages[0].Role)
	assert.Equal(t, "message1 fragment from rag", result.Messages[0].Content)
	assert.Equal(t, shared.AssistantRole, result.Messages[1].Role)
	assert.Equal(t, "this is the AI answer", result.Messages[1].Content)

	sentMessages := client.Calls[0].Arguments[1].([]shared.Message)
	assert.Len(t, sentMessages, 1)
	assert.Equal(t, sentMessages[0].Content, "message1 fragment from rag")
}

func TestEnrich(t *testing.T) {
	store := memory.New()
	manager := ct.New(store)

	ai := assistant.New(nil, manager, nil)
	ctx := context.Background()

	context1 := shared.Context{
		ID:        uuid.NewString(),
		Name:      "context1",
		CreatedAt: time.Now().UTC(),
	}
	err := manager.CreateContext(ctx, context1)
	assert.NoError(t, err)
	err = manager.AddMessagesToContext(ctx, context1.ID, []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.AssistantRole,
			Content:   "context 1 content 1",
			CreatedAt: time.Now().UTC(),
		},
	})
	assert.NoError(t, err)
	context2 := shared.Context{
		ID:        uuid.NewString(),
		Name:      "context2",
		CreatedAt: time.Now().UTC(),
	}
	err = manager.CreateContext(ctx, context2)
	assert.NoError(t, err)
	err = manager.AddMessagesToContext(ctx, context2.ID, []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "context 2 content 1",
			CreatedAt: time.Now().UTC(),
		},
	})
	assert.NoError(t, err)
	err = manager.AddMessagesToContext(ctx, context2.ID, []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "context 2 content 2",
			CreatedAt: time.Now().UTC(),
		},
	})
	assert.NoError(t, err)
	context3 := shared.Context{
		ID:        uuid.NewString(),
		Name:      "context3",
		CreatedAt: time.Now().UTC(),
		Sources: shared.ContextSources{
			Contexts: []string{context1.ID, context2.ID},
		},
	}
	err = manager.CreateContext(ctx, context3)
	assert.NoError(t, err)
	err = manager.AddMessagesToContext(ctx, context3.ID, []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "context 3 content 1",
			CreatedAt: time.Now().UTC(),
		},
	})
	assert.NoError(t, err)

	context4 := shared.Context{
		ID:        uuid.NewString(),
		Name:      "context4",
		CreatedAt: time.Now().UTC(),
		Sources: shared.ContextSources{
			Contexts: []string{context3.ID},
		},
	}
	err = manager.CreateContext(ctx, context4)
	assert.NoError(t, err)
	err = manager.AddMessagesToContext(ctx, context4.ID, []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.AssistantRole,
			Content:   "context 4 content 1",
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "context 4 content 2",
			CreatedAt: time.Now().UTC(),
		},
	})
	assert.NoError(t, err)

	messages := []shared.Message{
		{
			ID:        uuid.NewString(),
			Role:      shared.AssistantRole,
			Content:   "final msg 1",
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.NewString(),
			Role:      shared.UserRole,
			Content:   "final msg 2",
			CreatedAt: time.Now().UTC(),
		},
	}
	// context 4 (2 messages) => context 3 (1 message) => [ context 2 (2 message), context 1 (1 message) ]

	context, err := store.GetContext(ctx, context4.ID)
	assert.NoError(t, err)
	result, err := ai.Enrich(ctx, context, messages)
	assert.NoError(t, err)
	assert.Len(t, result, 8)
	assert.Equal(t, "context 1 content 1", result[0].Content)
	assert.Equal(t, shared.AssistantRole, result[0].Role)
	assert.Equal(t, "context 2 content 1", result[1].Content)
	assert.Equal(t, shared.UserRole, result[1].Role)
	assert.Equal(t, "context 2 content 2", result[2].Content)
	assert.Equal(t, shared.UserRole, result[2].Role)
	assert.Equal(t, "context 3 content 1", result[3].Content)
	assert.Equal(t, shared.UserRole, result[3].Role)
	assert.Equal(t, "context 4 content 1", result[4].Content)
	assert.Equal(t, shared.AssistantRole, result[4].Role)
	assert.Equal(t, "context 4 content 2", result[5].Content)
	assert.Equal(t, shared.UserRole, result[5].Role)
	assert.Equal(t, "final msg 1", result[6].Content)
	assert.Equal(t, shared.AssistantRole, result[6].Role)
	assert.Equal(t, "final msg 2", result[7].Content)
	assert.Equal(t, shared.UserRole, result[7].Role)

}
