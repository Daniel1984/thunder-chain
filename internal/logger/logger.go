package logger

import (
	"log/slog"
	"os"
)

func WithJSONFormat() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	return logger
}
