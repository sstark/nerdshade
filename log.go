package main

import (
	"log/slog"
	"os"
)

func noTimeHandler(debugLevel slog.Level) slog.Handler {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: debugLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return th
}
