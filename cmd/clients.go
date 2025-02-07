package cmd

import (
	"errors"
	"os"

	"github.com/appclacks/maizai/config"
	"github.com/appclacks/maizai/internal/providers/anthropic"
	"github.com/appclacks/maizai/internal/providers/mistral"
	"github.com/appclacks/maizai/pkg/assistant"
)

func BuildProviders(config config.ProvidersConfiguration) (map[string]assistant.Provider, error) {
	clients := make(map[string]assistant.Provider)
	if config.Anthropic != "" {
		anthropic := anthropic.New(anthropic.Config{
			APIKey: config.Anthropic,
		})
		os.Unsetenv("MAIZAI_ANTHROPIC_API_KEY")
		clients["anthropic"] = anthropic
	}
	if config.Mistral != "" {
		mistral := mistral.New(mistral.Config{
			APIKey: config.Mistral,
		})
		os.Unsetenv("MAIZAI_MISTRAL_API_KEY")
		clients["mistral"] = mistral
	}
	if len(clients) == 0 {
		return nil, errors.New("No AI client configured")
	}
	return clients, nil

}
