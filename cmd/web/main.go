package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"

	"github.com/0xrinful/Zenq/internal/requester"
	"github.com/0xrinful/Zenq/internal/requester/flare"
	"github.com/0xrinful/Zenq/internal/sources"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	}))
	slog.SetDefault(logger)

	solver := flare.NewSolver()

	requester := requester.New(solver, sources.Config{NeedsFlare: false})

	resp, err := requester.Get(
		context.Background(),
		"https://olympustaff.com/series/dcmkmk/171",
	)
	if err != nil {
		slog.Error("request failed", "err", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("non-200 response",
			"status", resp.StatusCode,
			"status_text", resp.Status,
		)
		return
	}

	slog.Info("request ok", "status", resp.StatusCode)
}
