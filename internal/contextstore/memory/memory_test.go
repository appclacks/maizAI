package memory_test

import (
	"context"
	"testing"

	"github.com/appclacks/maizai/internal/contextstore/memory"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	store := memory.New()
	ctx := context.Background()
	contexts, err := store.ListContexts(ctx)
	assert.NoError(t, err)
	assert.Len(t, contexts, 0)

	context := shared.Context{
		ID:          uuid.New().String(),
		Name:        "foo",
		Description: "bar",
		Sources: shared.ContextSources{
			Contexts: []string{uuid.NewString()},
		},
		Messages: []shared.Message{
			{
				ID:      uuid.NewString(),
				Role:    shared.UserRole,
				Content: "hello world",
			},
		},
	}

	exists, err := store.ContextExists(ctx, context.ID)
	assert.NoError(t, err)
	assert.False(t, exists)

	_, err = store.GetContext(ctx, context.ID)
	assert.Error(t, err)

	err = store.CreateContext(ctx, context)
	assert.NoError(t, err)

	contexts, err = store.ListContexts(ctx)
	assert.NoError(t, err)
	assert.Len(t, contexts, 1)
	assert.Equal(t, contexts[0].ID, context.ID)
	assert.Equal(t, contexts[0].Name, context.Name)
	assert.Equal(t, contexts[0].Sources.Contexts[0], context.Sources.Contexts[0])
	assert.Equal(t, contexts[0].Description, context.Description)

	exists, err = store.ContextExists(ctx, context.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	result, err := store.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Equal(t, context.Description, result.Description)
	assert.Len(t, context.Messages, 1)

	newMessages := []shared.Message{
		{
			ID:      uuid.NewString(),
			Role:    shared.UserRole,
			Content: "message2",
		},
		{
			ID:      uuid.NewString(),
			Role:    shared.AssistantRole,
			Content: "message3",
		},
	}
	err = store.AddMessages(ctx, context.ID, newMessages)
	assert.NoError(t, err)

	result, err = store.GetContext(ctx, context.ID)
	assert.Equal(t, context.Description, result.Description)
	assert.NoError(t, err)
	assert.Len(t, result.Messages, 3)

	err = store.DeleteContextSourceContext(ctx, context.ID, context.Sources.Contexts[0])
	assert.NoError(t, err)

	err = store.UpdateContextMessage(ctx, context.Messages[0].ID, shared.AssistantRole, "updated content")
	assert.NoError(t, err)

	result, err = store.GetContext(ctx, context.ID)
	assert.Equal(t, context.Description, result.Description)
	assert.NoError(t, err)
	assert.Len(t, result.Messages, 3)
	assert.Equal(t, result.Messages[0].Content, "updated content")
	assert.Equal(t, result.Messages[0].Role, shared.AssistantRole)

	err = store.DeleteContextMessage(ctx, context.Messages[0].ID)
	assert.NoError(t, err)
	result, err = store.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Len(t, result.Messages, 2)

	err = store.DeleteContext(ctx, context.ID)
	assert.NoError(t, err)

	contexts, err = store.ListContexts(ctx)
	assert.NoError(t, err)
	assert.Len(t, contexts, 0)
}
