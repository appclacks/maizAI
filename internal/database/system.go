package database

import (
	"context"
	"fmt"

	"github.com/appclacks/maizai/internal/database/queries"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/jackc/pgx/v5"
	er "github.com/mcorbin/corbierror"
)

func (c *Database) CreateSystemPrompt(ctx context.Context, prompt shared.SystemPrompt) error {
	err := c.queries.CreateSystemPrompt(ctx, queries.CreateSystemPromptParams{
		ID:          pgxID(prompt.ID),
		Name:        prompt.Name,
		Description: pgxText(prompt.Description),
		Content:     prompt.Content,
		CreatedAt:   pgxTime(prompt.CreatedAt),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) GetSystemPrompt(ctx context.Context, id string) (*shared.SystemPrompt, error) {
	prompt, err := c.queries.GetSystemPrompt(ctx, pgxID(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, er.Newf("system prompt %s doesn't exist", er.NotFound, true, id)
		}
		return nil, err
	}
	return &shared.SystemPrompt{
		ID:          prompt.ID.String(),
		Name:        prompt.Name,
		Description: prompt.Description.String,
		Content:     prompt.Content,
		CreatedAt:   prompt.CreatedAt.Time,
	}, nil
}

func (c *Database) GetSystemPromptByName(ctx context.Context, name string) (*shared.SystemPrompt, error) {
	prompt, err := c.queries.GetSystemPromptByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, er.Newf("system prompt with name %s doesn't exist", er.NotFound, true, name)
		}
		return nil, err
	}
	return &shared.SystemPrompt{
		ID:          prompt.ID.String(),
		Name:        prompt.Name,
		Description: prompt.Description.String,
		Content:     prompt.Content,
		CreatedAt:   prompt.CreatedAt.Time,
	}, nil
}

func (c *Database) SystemPromptExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := c.queries.GetSystemPromptByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("fail to check system prompt %s: %w", name, err)
	}
	return true, nil
}

func (c *Database) ListSystemPrompts(ctx context.Context) ([]shared.SystemPrompt, error) {
	prompts, err := c.queries.ListSystemPrompts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]shared.SystemPrompt, 0, len(prompts))
	for _, p := range prompts {
		result = append(result, shared.SystemPrompt{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description.String,
			CreatedAt:   p.CreatedAt.Time,
		})
	}
	return result, nil
}

func (c *Database) UpdateSystemPrompt(ctx context.Context, id string, content string) error {
	err := c.queries.UpdateSystemPrompt(ctx, queries.UpdateSystemPromptParams{
		ID:      pgxID(id),
		Content: content,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return er.Newf("system prompt %s doesn't exist", er.NotFound, true, id)
		}
		return err
	}
	return nil
}

func (c *Database) DeleteSystemPrompt(ctx context.Context, id string) error {
	err := c.queries.DeleteSystemPrompt(ctx, pgxID(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return er.Newf("system prompt %s doesn't exist", er.NotFound, true, id)
		}
		return err
	}
	return nil
}
