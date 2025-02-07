package database

type Configuration struct {
	Username string `env:"MAIZAI_POSTGRESQL_USERNAME"`
	Password string `env:"MAIZAI_POSTGRESQL_PASSWORD"`
	Database string `env:"MAIZAI_POSTGRESQL_DATABASE"`
	Host     string `env:"MAIZAI_POSTGRESQL_HOST"`
	Port     uint   `env:"MAIZAI_POSTGRESQL_PORT"`
	SSLMode  string `env:"MAIZAI_POSTGRESQL_SSL_MODE"`
}
