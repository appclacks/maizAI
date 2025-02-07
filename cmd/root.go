package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

var logLevel string
var logFormat string

func printJson(result any) {
	b, err := json.Marshal(result)
	exitIfError(err)
	fmt.Println(string(b))
}

func Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		return err
	}

	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("maizai"),
	)

	shutdownFn := func() {}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" {
		tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp), trace.WithResource(r))
		otel.SetTracerProvider(tracerProvider)
		shutdownFn = func() {
			err := tracerProvider.Shutdown(context.Background())
			if err != nil {
				panic(err)
			}
		}
	}
	defer shutdownFn()

	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Root command",
	}

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Logger log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Logger logs format (text, json)")

	contextCmd := &cobra.Command{
		Use:   "context",
		Short: "Context subcommands",
	}
	contextMessageCmd := &cobra.Command{
		Use:   "message",
		Short: "Context messages subcommands",
	}
	contextSourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Context source subcommands",
	}
	documentCmd := &cobra.Command{
		Use:   "document",
		Short: "Document subcommands",
	}
	documentChunkCmd := &cobra.Command{
		Use:   "document-chunk",
		Short: "Document chunk subcommands",
	}
	embeddingCmd := &cobra.Command{
		Use:   "embedding",
		Short: "Embedding commands",
	}
	serverCmd := buildServerCmd()
	embeddingCmd.AddCommand(embeddingMatchCmd())
	documentCmd.AddCommand(documentListCmd())
	documentCmd.AddCommand(documentCreateCmd())
	documentCmd.AddCommand(documentEmbedCmd())
	documentCmd.AddCommand(documentDeleteCmd())
	documentCmd.AddCommand(documentGetCmd())
	documentCmd.AddCommand(documentChunkListCmd())
	documentChunkCmd.AddCommand(documentChunkDeleteCmd())
	contextCmd.AddCommand(contextMessageCmd)
	contextCmd.AddCommand(contextSourceCmd)
	contextCmd.AddCommand(contextCreateCmd())
	contextCmd.AddCommand(contextListCmd())
	contextCmd.AddCommand(contextDeleteCmd())
	contextCmd.AddCommand(contextGetCmd())
	contextMessageCmd.AddCommand(addMessagesToContextCmd())
	contextMessageCmd.AddCommand(messageUpdateCmd())
	contextMessageCmd.AddCommand(deleteContextMessageCmd())
	contextSourceCmd.AddCommand(contextSourceContextDeleteCmd())
	contextSourceCmd.AddCommand(contextSourceContextAddCmd())

	conversationCmd := buildConversationCmd()
	rootCmd.AddCommand(embeddingCmd)
	rootCmd.AddCommand(documentCmd)
	rootCmd.AddCommand(documentChunkCmd)
	rootCmd.AddCommand(conversationCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(serverCmd)
	shutdown, err := initOpentelemetry()
	if err != nil {
		return err
	}
	defer shutdown()
	return rootCmd.Execute()
}

func exitIfError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
