package logger

import (
	"log/slog"
	"os"
)

func New(serviceName string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("service", serviceName)
}
