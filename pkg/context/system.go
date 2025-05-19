package context

import (
	"context"
	"errors"

	"github.com/appclacks/maizai/internal/id"
	"github.com/appclacks/maizai/pkg/shared"
)

func (c *ContextManager) CreateSystemPrompt(ctx context.Context, prompt shared.SystemPrompt) error {
	err := prompt.Validate()
	if err != nil {
		return err
	}
	return c.store.CreateSystemPrompt(ctx, prompt)
}

func (c *ContextManager) GetSystemPrompt(ctx context.Context, promptID string) (*shared.SystemPrompt, error) {
	if err := id.Validate(promptID, "Invalid system prompt ID"); err != nil {
		return nil, err
	}
	return c.store.GetSystemPrompt(ctx, promptID)
}

func (c *ContextManager) ListSystemPrompts(ctx context.Context) ([]shared.SystemPrompt, error) {
	return c.store.ListSystemPrompts(ctx)
}

func (c *ContextManager) DeleteSystemPrompt(ctx context.Context, promptID string) error {
	if err := id.Validate(promptID, "Invalid system prompt ID"); err != nil {
		return err
	}
	return c.store.DeleteSystemPrompt(ctx, promptID)
}

func (c *ContextManager) UpdateSystemPrompt(ctx context.Context, promptID string, content string) error {
	if err := id.Validate(promptID, "Invalid system prompt ID"); err != nil {
		return err
	}
	if content == "" {
		return errors.New("System prompt content is empty")
	}
	return c.store.UpdateSystemPrompt(ctx, promptID, content)
}
