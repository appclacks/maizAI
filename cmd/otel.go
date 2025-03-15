package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initOpentelemetry() (func(), error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" {
		r := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("maizAI"),
		)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		exp, err := otlptracehttp.New(ctx)
		cancel()
		if err != nil {
			return nil, err
		}
		slog.Info("starting opentelemetry traces export")
		tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp), trace.WithResource(r))
		otel.SetTracerProvider(tracerProvider)
		return func() {
			err := tracerProvider.Shutdown(context.Background())
			if err != nil {
				panic(err)
			}
		}, nil
	}
	return func() {}, nil
}
