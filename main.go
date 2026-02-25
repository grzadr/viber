// Package main is the entry point for the viber application.
package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/grzadr/viber/internal/config"
	"github.com/grzadr/viber/internal/files"
	"github.com/grzadr/viber/internal/sub"
)

type App struct {
	logger *slog.Logger
}

// NewApp initializes the application with a dedicated logger.
func NewApp(logger *slog.Logger) *App {
	return &App{
		logger: logger,
	}
}

func (a *App) DebugLog(msg string, args ...any) {
	a.logger.Debug(msg, args...)
}

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	// 2. Inject it into your application.
	app := NewApp(logger)

	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	app.DebugLog("loaded configuration", "config", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for audioFile, err := range files.GetAudioPaths(cfg.Paths) {
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		app.DebugLog("Processing file", "path", audioFile.Path)

		stdout, stderr, cmdErr := sub.StreamCommand(
			ctx,
			"ffprobe",
			"-v",
			"error",
			"-hide_banner",
			"-select_streams",
			"a:0",
			"-show_entries",
			"format=size,duration:format_tags=ARTIST,TITLE,ALBUM,track,disc,TRACKTOTAL,DISCTOTAL:stream=sample_rate,bits_per_raw_sample",
			"-print_format",
			"json",
			audioFile.Path,
		)
		if cmdErr != nil {
			app.DebugLog("Error: %v\n", cmdErr)
		}

		for line := range stdout {
			app.DebugLog("stdout", "line", line)
		}

		for line := range stderr {
			app.DebugLog("stderr", "line", line)
		}

		break
	}
}
