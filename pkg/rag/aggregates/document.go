package aggregates

import (
	"errors"
	"time"

	"github.com/appclacks/maizai/internal/id"
	"github.com/google/uuid"
)

type Document struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created-at"`
}

func (d Document) Validate() error {
	if err := id.Validate(d.ID, "invalid document ID"); err != nil {
		return err
	}
	if d.Name == "" {
		return errors.New("Invalid document name")
	}
	if d.CreatedAt.IsZero() {
		return errors.New("Invalid creation date")
	}
	return nil
}

func NewDocument(name string, description string) (*Document, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	return &Document{
		ID:          id.String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

type DocumentChunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document-id,omitempty"`
	Fragment   string    `json:"fragment"`
	Embedding  []float32 `json:"-"`
	CreatedAt  time.Time `json:"created-at"`
}

func (d DocumentChunk) Validate() error {
	if err := id.Validate(d.ID, "invalid document chunk ID"); err != nil {
		return err
	}
	if err := id.Validate(d.DocumentID, "invalid document ID"); err != nil {
		return err
	}
	if d.Fragment == "" {
		return errors.New("Invalid fragment")
	}
	if len(d.Embedding) == 0 {
		return errors.New("Invalid embedding")
	}
	if d.CreatedAt.IsZero() {
		return errors.New("Invalid creation date")
	}
	return nil
}

func NewDocumentChunk(docID string, fragment string, embedding []float32) (*DocumentChunk, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	return &DocumentChunk{
		ID:         id.String(),
		DocumentID: docID,
		Fragment:   fragment,
		Embedding:  embedding,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

type EmbeddingQuery struct {
	Input    string `json:"input"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

type SearchQuery struct {
	Input    string `json:"input"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Limit    int32  `json:"limit"`
}

func (s SearchQuery) Validate() error {
	if s.Input == "" {
		return errors.New("Invalid input field")
	}
	if s.Model == "" {
		return errors.New("Invalid model")
	}
	if s.Provider == "" {
		return errors.New("Invalid provider")
	}
	if s.Limit == 0 {
		return errors.New("Invalid limti")
	}
	return nil
}

type Embedding struct {
	Embedding []float32
}

type EmbeddingAnswer struct {
	InputTokens  uint64 `json:"input-tokens"`
	OutputTokens uint64 `json:"output-tokens"`
	Data         []Embedding
}
