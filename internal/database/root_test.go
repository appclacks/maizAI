package database_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/appclacks/maizai/internal/database"
)

var TestComponent *database.Database

func InitTestDB(logger *slog.Logger) *database.Database {
	config := database.Configuration{
		Username: "appclacks",
		Password: "appclacks",
		Database: "appclacks",
		Host:     "127.0.0.1",
		Port:     5432,
		SSLMode:  "disable",
	}
	c, err := database.New(config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = Cleanup(c)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("db cleanup done")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	return c
}

func Cleanup(c *database.Database) error {
	for _, query := range database.CleanupQueries {
		_, err := c.Exec(query)
		if err != nil {
			return fmt.Errorf("fail to clean DB on query %s: %w", query, err)
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	logger := slog.Default()
	TestComponent = InitTestDB(logger)
	exitVal := m.Run()
	err := Cleanup(TestComponent)
	if err != nil {
		logger.Error(err.Error())
	}
	os.Exit(exitVal)
}
