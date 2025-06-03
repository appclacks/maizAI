package shared

import (
	"errors"
	"time"

	"github.com/appclacks/maizai/internal/id"
	"github.com/google/uuid"
)

type SystemPrompt struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created-at"`
}

func NewSystemPrompt(name, description, content string) (*SystemPrompt, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	return &SystemPrompt{
		ID:          id.String(),
		Name:        name,
		Description: description,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (s SystemPrompt) Validate() error {
	if err := id.Validate(s.ID, "Invalid system prompt ID"); err != nil {
		return err
	}
	if s.Name == "" {
		return errors.New("A system prompt name is mandatory")
	}
	if s.Content == "" {
		return errors.New("System prompt content is empty")
	}
	if s.CreatedAt.IsZero() {
		return errors.New("A system prompt should have a creation date")
	}
	return nil
}
