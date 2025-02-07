package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/appclacks/maizai/internal/database/queries"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/jackc/pgx/v5"
	er "github.com/mcorbin/corbierror"
)

func (c *Database) CreateContext(ctx context.Context, context shared.Context) error {
	tx, err := c.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	shouldRollback := true
	defer func() {
		if shouldRollback {
			err := tx.Rollback(ctx)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}()
	qtx := c.queries.WithTx(tx)

	_, err = qtx.CreateContext(ctx, queries.CreateContextParams{
		ID:          pgxID(context.ID),
		Name:        context.Name,
		Description: pgxText(context.Description),
		CreatedAt:   pgxTime(context.CreatedAt),
	})
	if err != nil {
		return err
	}
	for _, source := range context.Sources.Contexts {
		err := qtx.CreateContexSource(ctx,
			queries.CreateContexSourceParams{
				SourceContextID: pgxID(source),
				ContextID:       pgxID(context.ID),
			})
		if err != nil {
			return err
		}
	}
	for _, message := range context.Messages {
		_, err := qtx.CreateContextMessage(
			ctx,
			queries.CreateContextMessageParams{
				ID:        pgxID(message.ID),
				Role:      message.Role,
				Content:   message.Content,
				CreatedAt: pgxTime(message.CreatedAt),
				ContextID: pgxID(context.ID),
			})
		if err != nil {
			return err
		}
	}
	shouldRollback = false
	return tx.Commit(ctx)
}

func (c *Database) GetContext(ctx context.Context, id string) (*shared.Context, error) {
	context, err := c.queries.GetContext(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
		return nil, er.Newf("context %s doesn't exist", er.NotFound, true, id)
	}
	result := shared.Context{
		ID:          context.ID.String(),
		Name:        context.Name,
		Description: context.Description.String,
		CreatedAt:   context.CreatedAt.Time,
		Sources: shared.ContextSources{
			Contexts: []string{},
		},
	}
	messages, err := c.queries.GetContextMessages(ctx, pgxID(id))
	if err != nil {
		return nil, err
	}
	for _, message := range messages {
		result.Messages = append(result.Messages, shared.Message{
			ID:        message.ID.String(),
			Role:      message.Role,
			Content:   message.Content,
			CreatedAt: message.CreatedAt.Time,
		})
	}
	sources, err := c.queries.GetContextSourcesForContext(ctx, pgxID(id))
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		result.Sources.Contexts = append(result.Sources.Contexts, source.String())
	}
	return &result, nil

}

func (c *Database) AddMessages(ctx context.Context, id string, messages []shared.Message) error {

	tx, qtx, rollbackFn, err := c.beginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackFn()
	exists, err := c.exists(qtx, ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return er.Newf("context %s doesn't exist", er.NotFound, true, id)
	}
	for _, message := range messages {
		_, err := qtx.CreateContextMessage(
			ctx,
			queries.CreateContextMessageParams{
				ID:        pgxID(message.ID),
				Role:      message.Role,
				Content:   message.Content,
				CreatedAt: pgxTime(message.CreatedAt),
				ContextID: pgxID(id),
			})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (c *Database) DeleteContext(ctx context.Context, id string) error {
	tx, err := c.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint
	qtx := c.queries.WithTx(tx)
	err = qtx.CleanContextSourcesForContext(ctx, pgxID(id))
	if err != nil {
		return err
	}
	err = qtx.DeleteContextMessagesForContext(ctx, pgxID(id))
	if err != nil {
		return err
	}
	err = qtx.DeleteContext(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return err
		}
		return er.Newf("context %s doesn't exist", er.NotFound, true, id)
	}

	return tx.Commit(ctx)
}

func (c *Database) ContextExists(ctx context.Context, id string) (bool, error) {
	return c.exists(c.queries, ctx, id)
}

func (c *Database) exists(queries *queries.Queries, ctx context.Context, id string) (bool, error) {
	_, err := queries.GetContextNameByID(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return false, fmt.Errorf("fail to get context %s: %w", id, err)
		}
		return false, nil
	}
	return true, nil
}

func (c *Database) ContextExistsByName(ctx context.Context, name string) (bool, error) {
	_, err := c.queries.GetContextIDByName(ctx, name)
	if err != nil {
		if err != pgx.ErrNoRows {
			return false, fmt.Errorf("fail to get context %s: %w", name, err)
		}
		return false, nil
	}
	return true, nil
}

func (c *Database) ListContexts(ctx context.Context) ([]shared.ContextMetadata, error) {
	metadata, err := c.queries.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	result := []shared.ContextMetadata{}
	for _, m := range metadata {
		metadata := shared.ContextMetadata{
			ID:          m.ID.String(),
			Name:        m.Name,
			Description: m.Description.String,
			CreatedAt:   m.CreatedAt.Time,
		}
		sources, err := c.queries.GetContextSourcesForContext(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		for _, source := range sources {
			metadata.Sources.Contexts = append(metadata.Sources.Contexts, source.String())
		}
		result = append(result, metadata)
	}
	return result, nil

}

func (c *Database) UpdateContextMessage(ctx context.Context, messageID string, role string, content string) error {
	return c.queries.UpdateContextMessage(ctx, queries.UpdateContextMessageParams{
		ID:      pgxID(messageID),
		Role:    role,
		Content: content,
	})
}

func (c *Database) DeleteContextMessage(ctx context.Context, messageID string) error {
	return c.queries.DeleteContextMessage(ctx, pgxID(messageID))
}

func (c *Database) DeleteContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	return c.queries.DeleteContextSource(ctx, queries.DeleteContextSourceParams{
		ContextID:       pgxID(contextID),
		SourceContextID: pgxID(sourceContextID),
	})
}

func (c *Database) CreateContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	return c.queries.CreateContexSource(ctx, queries.CreateContexSourceParams{
		ContextID:       pgxID(contextID),
		SourceContextID: pgxID(sourceContextID),
	})
}
