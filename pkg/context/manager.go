package context

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/appclacks/maizai/internal/id"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
)

type ContextStore interface {
	GetContext(ctx context.Context, id string) (*shared.Context, error)
	CreateContext(ctx context.Context, context shared.Context) error
	ContextExists(ctx context.Context, id string) (bool, error)
	ContextExistsByName(ctx context.Context, name string) (bool, error)
	DeleteContext(ctx context.Context, id string) error
	ListContexts(ctx context.Context) ([]shared.ContextMetadata, error)
	AddMessages(ctx context.Context, id string, messages []shared.Message) error
	DeleteContextMessage(ctx context.Context, id string) error
	UpdateContextMessage(ctx context.Context, messageID string, role string, content string) error
	DeleteContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error
	CreateContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error
	DeleteContextMessages(ctx context.Context, contextID string) error
}

type ContextManager struct {
	store ContextStore
}

func New(store ContextStore) *ContextManager {
	return &ContextManager{
		store: store,
	}
}

func NewContext(options shared.ContextOptions) (*shared.Context, error) {
	err := options.Validate()
	if err != nil {
		return nil, err
	}

	uuid, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	return &shared.Context{
		ID:          uuid.String(),
		Sources:     options.Sources,
		Name:        options.Name,
		Description: options.Description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (c *ContextManager) CreateContext(ctx context.Context, context shared.Context) error {
	err := context.Validate()
	if err != nil {
		return err
	}
	exists, err := c.store.ContextExistsByName(ctx, context.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("A context with name %s already exists", context.Name)
	}
	return c.store.CreateContext(ctx, context)
}

func (c *ContextManager) CreateOrGetContext(ctx context.Context, contextID string, options shared.ContextOptions) (*shared.Context, error) {
	if contextID != "" {
		if err := id.Validate(contextID, "Invalid context ID"); err != nil {
			return nil, err
		}
		return c.store.GetContext(ctx, contextID)
	}
	context, err := NewContext(options)
	if err != nil {
		return nil, err
	}
	err = c.CreateContext(ctx, *context)
	if err != nil {
		return nil, err
	}
	return context, nil
}

func (c *ContextManager) GetContext(ctx context.Context, contextID string) (*shared.Context, error) {
	if err := id.Validate(contextID, "Invalid context ID"); err != nil {
		return nil, err
	}
	return c.store.GetContext(ctx, contextID)
}

func (c *ContextManager) DeleteContext(ctx context.Context, contextID string) error {
	if err := id.Validate(contextID, "Invalid context ID"); err != nil {
		return err
	}
	return c.store.DeleteContext(ctx, contextID)
}

func (c *ContextManager) AddMessagesToContext(ctx context.Context, contextID string, messages []shared.Message) error {
	if err := id.Validate(contextID, "Invalid context ID"); err != nil {
		return err
	}
	if len(messages) == 0 {
		return errors.New("You need at least one message to add to the context")
	}
	return c.store.AddMessages(ctx, contextID, messages)
}

func (c *ContextManager) DeleteContextMessage(ctx context.Context, messageID string) error {
	if err := id.Validate(messageID, "Invalid message ID"); err != nil {
		return err
	}
	return c.store.DeleteContextMessage(ctx, messageID)
}

func (c *ContextManager) ListContexts(ctx context.Context) ([]shared.ContextMetadata, error) {
	return c.store.ListContexts(ctx)
}

func (c *ContextManager) UpdateContextMessage(ctx context.Context, messageID string, role string, content string) error {
	if err := id.Validate(messageID, "Invalid message ID"); err != nil {
		return err
	}
	if content == "" {
		return errors.New("Empty content for message")
	}
	if role != shared.UserRole && role != shared.AssistantRole {
		return errors.New("Invalid message role")
	}
	return c.store.UpdateContextMessage(ctx, messageID, role, content)
}

func (c *ContextManager) DeleteContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	if err := id.Validate(contextID, "Invalid context ID"); err != nil {
		return err
	}
	if err := id.Validate(sourceContextID, "Invalid source context ID"); err != nil {
		return err
	}

	return c.store.DeleteContextSourceContext(ctx, contextID, sourceContextID)
}

func (c *ContextManager) CreateContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	if err := id.Validate(contextID, "Invalid context ID"); err != nil {
		return err
	}
	if err := id.Validate(sourceContextID, "Invalid source context ID"); err != nil {
		return err
	}

	return c.store.CreateContextSourceContext(ctx, contextID, sourceContextID)
}

func (c *ContextManager) DeleteContextMessages(ctx context.Context, contextID string) error {
	return c.store.DeleteContextMessages(ctx, contextID)
}
