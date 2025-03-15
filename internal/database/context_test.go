package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContextCRUD(t *testing.T) {
	ctx := context.Background()
	context := shared.Context{
		Name:        "test",
		ID:          uuid.New().String(),
		Description: "foo",
		CreatedAt:   time.Now().UTC(),
		Sources:     shared.ContextSources{},
		Messages: []shared.Message{
			{
				ID:        uuid.New().String(),
				Role:      shared.AssistantRole,
				Content:   "1234",
				CreatedAt: time.Now().UTC(),
			},
		},
	}
	err := TestComponent.CreateContext(ctx, context)
	assert.NoError(t, err)

	get, err := TestComponent.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Equal(t, get.ID, context.ID)
	assert.Equal(t, get.Name, context.Name)
	assert.Equal(t, get.Description, context.Description)
	assert.Len(t, get.Messages, 1)
	assert.Equal(t, get.Messages[0].Content, "1234")
	assert.Equal(t, get.Messages[0].ID, context.Messages[0].ID)
	assert.Equal(t, get.Messages[0].Role, context.Messages[0].Role)

	err = TestComponent.UpdateContextMessage(ctx, context.Messages[0].ID, shared.UserRole, "new message")
	assert.NoError(t, err)

	get, err = TestComponent.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Equal(t, get.Messages[0].Content, "new message")
	assert.Equal(t, get.Messages[0].ID, context.Messages[0].ID)
	assert.Equal(t, get.Messages[0].Role, shared.UserRole)

	listResult, err := TestComponent.ListContexts(ctx)
	assert.NoError(t, err)

	assert.Len(t, listResult, 1)
	assert.Equal(t, listResult[0].ID, context.ID)
	assert.Equal(t, listResult[0].Name, context.Name)
	assert.Equal(t, listResult[0].Description, context.Description)

	contextWithSource := shared.Context{
		Name:        "test2",
		ID:          uuid.New().String(),
		Description: "foo",
		CreatedAt:   time.Now().UTC(),
		Sources: shared.ContextSources{
			Contexts: []string{context.ID},
		},
		Messages: []shared.Message{
			{
				ID:        uuid.New().String(),
				Role:      shared.AssistantRole,
				Content:   "1234",
				CreatedAt: time.Now().UTC(),
			},
			{
				ID:        uuid.New().String(),
				Role:      shared.UserRole,
				Content:   "456",
				CreatedAt: time.Now().UTC(),
			},
		},
	}
	err = TestComponent.CreateContext(ctx, contextWithSource)
	assert.NoError(t, err)

	getSrc, err := TestComponent.GetContext(ctx, contextWithSource.ID)
	assert.NoError(t, err)
	assert.Equal(t, getSrc.ID, contextWithSource.ID)
	assert.Equal(t, getSrc.Name, contextWithSource.Name)
	assert.Equal(t, getSrc.Description, contextWithSource.Description)
	assert.Len(t, getSrc.Sources.Contexts, 1)
	assert.Len(t, getSrc.Messages, 2)
	assert.Equal(t, getSrc.Messages[0].Content, "1234")
	assert.Equal(t, getSrc.Messages[0].ID, contextWithSource.Messages[0].ID)
	assert.Equal(t, getSrc.Messages[0].Role, contextWithSource.Messages[0].Role)
	assert.Equal(t, getSrc.Messages[1].Content, "456")
	assert.Equal(t, getSrc.Messages[1].ID, contextWithSource.Messages[1].ID)
	assert.Equal(t, getSrc.Messages[1].Role, contextWithSource.Messages[1].Role)

	exists, err := TestComponent.ContextExists(ctx, getSrc.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = TestComponent.ContextExists(ctx, uuid.New().String())
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = TestComponent.ContextExistsByName(ctx, getSrc.Name)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = TestComponent.ContextExistsByName(ctx, "azeaajazie")
	assert.NoError(t, err)
	assert.False(t, exists)

	messagesToAdd := []shared.Message{
		{
			ID:        uuid.New().String(),
			Role:      shared.AssistantRole,
			Content:   "9876",
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.New().String(),
			Role:      shared.UserRole,
			Content:   "hello",
			CreatedAt: time.Now().UTC(),
		},
	}
	err = TestComponent.AddMessages(ctx, getSrc.ID, messagesToAdd)
	assert.NoError(t, err)

	getWithMsg, err := TestComponent.GetContext(ctx, contextWithSource.ID)
	assert.NoError(t, err)
	assert.Len(t, getWithMsg.Messages, 4)
	assert.Equal(t, getWithMsg.Messages[0].Content, "1234")
	assert.Equal(t, getWithMsg.Messages[0].ID, contextWithSource.Messages[0].ID)
	assert.Equal(t, getWithMsg.Messages[0].Role, contextWithSource.Messages[0].Role)
	assert.Equal(t, getWithMsg.Messages[1].Content, "456")
	assert.Equal(t, getWithMsg.Messages[1].ID, contextWithSource.Messages[1].ID)
	assert.Equal(t, getWithMsg.Messages[1].Role, contextWithSource.Messages[1].Role)
	assert.Equal(t, getWithMsg.Messages[2].Content, "9876")
	assert.Equal(t, getWithMsg.Messages[2].ID, messagesToAdd[0].ID)
	assert.Equal(t, getWithMsg.Messages[2].Role, messagesToAdd[0].Role)
	assert.Equal(t, getWithMsg.Messages[3].Content, "hello")
	assert.Equal(t, getWithMsg.Messages[3].ID, messagesToAdd[1].ID)
	assert.Equal(t, getWithMsg.Messages[3].Role, messagesToAdd[1].Role)

	listResult, err = TestComponent.ListContexts(ctx)
	assert.NoError(t, err)
	assert.Len(t, listResult, 2)

	err = TestComponent.DeleteContextMessage(ctx, getWithMsg.Messages[0].ID)
	assert.NoError(t, err)
	getWithMsg, err = TestComponent.GetContext(ctx, contextWithSource.ID)
	assert.NoError(t, err)
	assert.Len(t, getWithMsg.Messages, 3)

	err = TestComponent.CreateContextSourceContext(ctx, context.ID, getSrc.ID)
	assert.NoError(t, err)
	get, err = TestComponent.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Equal(t, get.ID, context.ID)
	assert.Len(t, get.Sources.Contexts, 1)
	assert.Equal(t, getSrc.ID, get.Sources.Contexts[0])

	err = TestComponent.DeleteContextSourceContext(ctx, context.ID, getSrc.ID)
	assert.NoError(t, err)
	get, err = TestComponent.GetContext(ctx, context.ID)
	assert.NoError(t, err)
	assert.Equal(t, get.ID, context.ID)
	assert.Len(t, get.Sources.Contexts, 0)

	err = TestComponent.DeleteContext(ctx, context.ID)
	assert.NoError(t, err)
	listResult, err = TestComponent.ListContexts(ctx)
	assert.NoError(t, err)
	assert.Len(t, listResult, 1)
}
