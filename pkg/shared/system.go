package shared

import (
	"errors"
	"time"

	"github.com/appclacks/maizai/internal/id"
)

type SystemPrompt struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created-at"`
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
