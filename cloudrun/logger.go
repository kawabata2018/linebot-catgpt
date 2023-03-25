package main

import (
	"os"

	"golang.org/x/exp/slog"
)

func NewGCPLogger() *slog.Logger {
	return slog.New(
		slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				// Cloud Logging LogSeverity
				// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
				switch {
				case a.Key == slog.MessageKey:
					return slog.String("message", a.Value.String())
				case a.Key == slog.LevelKey && a.Value.String() == slog.LevelWarn.String():
					return slog.String("severity", "WARNING")
				case a.Key == slog.LevelKey:
					return slog.String("severity", a.Value.String())
				}
				return a
			},
		}.NewJSONHandler(os.Stdout),
	)
}
