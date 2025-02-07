package cmd

import (
	"log/slog"
	"os"
)

func buildLogger(level string, format string) *slog.Logger {
	var programLevel = new(slog.LevelVar)
	switch level {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	options := &slog.HandlerOptions{Level: programLevel}
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
}
