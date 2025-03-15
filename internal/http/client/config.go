package client

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Configuration struct {
	Endpoint string `env:"MAIZAI_HTTP_ENDPOINT, default=http://0.0.0.0:3333"`
	Key      string `env:"MAIZAI_HTTP_TLS_KEY_PATH"`
	Cert     string `env:"MAIZAI_HTTP_TLS_CERT_PATH"`
	Cacert   string `env:"MAIZAI_HTTP_TLS_CACERT_PATH"`
	Insecure bool   `env:"MAIZAI_HTTP_TLS_INSECURE"`
}

func Load() (*Configuration, error) {
	var c Configuration
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return nil, err
	}
	return &c, nil
}
