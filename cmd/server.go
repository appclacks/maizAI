package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/appclacks/maizai/config"
	"github.com/appclacks/maizai/internal/database"
	"github.com/appclacks/maizai/internal/http"
	"github.com/appclacks/maizai/internal/http/handlers"
	"github.com/appclacks/maizai/pkg/assistant"
	ct "github.com/appclacks/maizai/pkg/context"
	"github.com/appclacks/maizai/pkg/rag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

func buildServerCmd() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Runs the HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			logger := buildLogger(logLevel, logFormat)
			err := RunServer()
			if err != nil {
				logger.Error(err.Error())
				os.Exit(2)
			}

		},
	}
	return serverCmd
}

func RunServer() error {
	registry := prometheus.DefaultRegisterer.(*prometheus.Registry)
	config, err := config.Load()
	exitIfError(err)
	db, err := database.New(config.Store.PostgreSQL)
	exitIfError(err)
	clients, err := BuildProviders(config.Providers)
	exitIfError(err)
	embeddingProviders := map[string]rag.AI{}
	if mistral, ok := clients["mistral"]; ok {
		embeddingProviders["mistral"] = mistral.(rag.AI)
	}
	manager := ct.New(db)

	rag := rag.New(db, embeddingProviders)
	ai := assistant.New(clients, manager, rag)

	handlersBuilder := handlers.NewBuilder(ai, manager, rag)
	server, err := http.New(config.HTTP, registry, handlersBuilder)
	if err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	errChan := make(chan error)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM)

	err = server.Start()
	if err != nil {
		return err
	}
	go func() {
		for sig := range signals {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				slog.Info(fmt.Sprintf("received signal %s, starting shutdown", sig))
				signal.Stop(signals)
				err := server.Stop()
				if err != nil {
					errChan <- err
				}
				errChan <- nil
			}

		}
	}()
	exitErr := <-errChan
	return exitErr

}
