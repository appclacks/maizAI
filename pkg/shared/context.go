package shared

import (
	"errors"
	"fmt"
	"time"

	"github.com/appclacks/maizai/internal/id"
	"github.com/google/uuid"
)

const UserRole = "user"
const AssistantRole = "assistant"

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created-at"`
}

func NewMessage(role string, content string) (*Message, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:        id.String(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewUserMessages(content string) ([]Message, error) {
	msg, err := NewMessage(UserRole, content)
	if err != nil {
		return nil, err
	}
	return []Message{*msg}, nil
}

func (m Message) Validate() error {
	if err := id.Validate(m.ID, "Invalid message ID"); err != nil {
		return err
	}
	if m.Role == "" {
		return errors.New("A role is mandatory for the message")
	}
	if m.Role != UserRole && m.Role != AssistantRole {
		return fmt.Errorf("Invalid value for role %s: The message role should be %s or %s", m.Role, UserRole, AssistantRole)
	}
	if m.Content == "" {
		return errors.New("Message content can't be empty")
	}
	return nil
}

type ContextSources struct {
	Contexts []string `json:"contexts,omitempty"`
}

type Context struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Sources     ContextSources `json:"sources"`
	Messages    []Message      `json:"messages"`
	CreatedAt   time.Time      `json:"created-at"`
}

type ContextMetadata struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created-at"`
	Sources     ContextSources `json:"sources"`
}

func (c Context) Validate() error {
	if err := id.Validate(c.ID, "invalid context ID"); err != nil {
		return err
	}
	if c.Name == "" {
		return errors.New("A context name is mandatory")
	}
	if c.CreatedAt.IsZero() {
		return errors.New("A context should have a creation date")
	}
	for _, message := range c.Messages {
		err := message.Validate()
		if err != nil {
			return err
		}
	}
	for _, sourceCtx := range c.Sources.Contexts {
		err := uuid.Validate(sourceCtx)
		if err != nil {
			return fmt.Errorf("Invalid source context: '%s' is not a valid context uuid", sourceCtx)
		}
	}
	return nil
}

type ContextOptions struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Sources     ContextSources `json:"sources"`
}

func (o *ContextOptions) Validate() error {
	if o.Name == "" {
		return errors.New("A context name is mandatory")
	}
	return nil
}
