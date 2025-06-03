package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSystemCRUD(t *testing.T) {
	ctx := context.Background()
	prompt := shared.SystemPrompt{
		ID:          uuid.NewString(),
		Name:        "prompt1",
		Content:     "my prompt",
		Description: "abc",
		CreatedAt:   time.Now().UTC(),
	}
	err := TestComponent.CreateSystemPrompt(ctx, prompt)
	assert.NoError(t, err)

	result, err := TestComponent.GetSystemPrompt(ctx, prompt.ID)
	assert.NoError(t, err)
	assert.Equal(t, prompt.Name, result.Name)
	assert.Equal(t, prompt.Content, result.Content)
	assert.Equal(t, prompt.Description, result.Description)

	result, err = TestComponent.GetSystemPromptByName(ctx, prompt.Name)
	assert.NoError(t, err)
	assert.Equal(t, prompt.Name, result.Name)
	assert.Equal(t, prompt.Content, result.Content)
	assert.Equal(t, prompt.Description, result.Description)

	found, err := TestComponent.SystemPromptExistsByName(ctx, prompt.Name)
	assert.NoError(t, err)
	assert.True(t, found)

	found, err = TestComponent.SystemPromptExistsByName(ctx, "unknown")
	assert.NoError(t, err)
	assert.False(t, found)

	err = TestComponent.UpdateSystemPrompt(ctx, prompt.ID, "new content")
	assert.NoError(t, err)

	result, err = TestComponent.GetSystemPrompt(ctx, prompt.ID)
	assert.NoError(t, err)
	assert.Equal(t, prompt.Name, result.Name)
	assert.Equal(t, "new content", result.Content)

	newPrompt := shared.SystemPrompt{
		ID:          uuid.NewString(),
		Name:        "prompt2",
		Content:     "my second prompt",
		Description: "abc",
		CreatedAt:   time.Now().UTC(),
	}
	err = TestComponent.CreateSystemPrompt(ctx, newPrompt)
	assert.NoError(t, err)

	prompts, err := TestComponent.ListSystemPrompts(ctx)
	assert.NoError(t, err)
	assert.Len(t, prompts, 2)

	err = TestComponent.DeleteSystemPrompt(ctx, prompt.ID)
	assert.NoError(t, err)

	_, err = TestComponent.GetSystemPrompt(ctx, prompt.ID)
	assert.ErrorContains(t, err, "doesn't exist")

	err = TestComponent.DeleteSystemPrompt(ctx, newPrompt.ID)
	assert.NoError(t, err)

}
