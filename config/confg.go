package config

import (
	"context"

	"github.com/appclacks/maizai/internal/database"
	"github.com/appclacks/maizai/internal/http"
	"github.com/sethvargo/go-envconfig"
)

type StoreConfiguration struct {
	Type       string `env:"MAIZAI_STORE_TYPE, default=memory"`
	PostgreSQL database.Configuration
}

type ProvidersConfiguration struct {
	Anthropic string `env:"MAIZAI_ANTHROPIC_API_KEY"`
	Mistral   string `env:"MAIZAI_MISTRAL_API_KEY"`
}

type Configuration struct {
	Providers ProvidersConfiguration
	Store     StoreConfiguration
	HTTP      http.Configuration
}

func Load() (*Configuration, error) {
	var c Configuration
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
