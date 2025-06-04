package aggregates

import (
	"errors"

	"github.com/appclacks/maizai/pkg/rag/aggregates"
)

type QueryOptions struct {
	Model       string                 `json:"model"`
	System      string                 `json:"system"`
	Temperature float64                `json:"temperature"`
	MaxTokens   uint64                 `json:"max-tokens"`
	Provider    string                 `json:"provider"`
	RagQuery    aggregates.SearchQuery `json:"rag,omitempty"`
}

func (q QueryOptions) Validate() error {
	if q.Model == "" {
		return errors.New("A model name is mandatory")
	}
	if q.Provider == "" {
		return errors.New("An AI provider name is mandatory")
	}
	return nil
}

type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type ToolUse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	InputJSON string `json:"input-json"`
}

type Result struct {
	Text       string     `json:"text"`
	ToolUse    ToolUse    `json:"tool-use"`
	ToolResult ToolResult `json:"tool-result"`
}

type Answer struct {
	Results      []Result `json:"result"`
	InputTokens  uint64   `json:"input-tokens"`
	OutputTokens uint64   `json:"output-tokens"`
	Context      string   `json:"context"`
}

type Event struct {
	Answer *Answer
	Delta  string
	Error  error
}
