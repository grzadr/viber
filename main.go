// Package main is the entry point for the viber application.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/grzadr/viber/internal/config"
	"github.com/grzadr/viber/internal/files"
	"github.com/grzadr/viber/internal/sub"
)

const DefaultOutputCapacity = 64

type App struct {
	logger *slog.Logger
}

// NewApp initializes the application with a dedicated logger.
func NewApp() *App {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	return &App{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, opts)),
	}
}

func (a *App) DebugLog(msg string, args ...any) {
	a.logger.Debug(msg, args...)
}

type MetaDataRaw struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	SampleRate       int `json:"sample_rate,string"`
	BitsPerRawSample int `json:"bits_per_raw_sample,string"`
}

type Format struct {
	Duration float64           `json:"duration,string"`
	Size     int64             `json:"size,string"`
	Tags     map[string]string `json:"tags"`
}

type MetaData struct {
	Path             string
	Filename         string
	SampleRate       int
	BitsPerRawSample int
	Duration         float64
	Size             int64
}

func NewMetaDataRaw(ch <-chan string) (*MetaDataRaw, error) {
	var builder strings.Builder

	for chunk := range ch {
		builder.WriteString(chunk)
	}

	m := new(MetaDataRaw)
	err := json.Unmarshal([]byte(builder.String()), m)

	return m, err
}

func (m *MetaDataRaw) String() string {
	prettyJSON, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("error formatting MediaData: %v", err)
	}

	return string(prettyJSON)
}

func run(cfg *config.ArgsParsed, app *App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for audioFile, err := range files.GetAudioPaths(cfg.Paths) {
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		app.DebugLog("Processing file", "path", audioFile.Path)

		stdout, cmdResult, cmdErr := sub.StreamCommand(
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

		meta, metaErr := NewMetaDataRaw(stdout)
		if metaErr != nil {
			return fmt.Errorf(
				"error getting metadata %s: %w",
				audioFile.Path,
				metaErr,
			)
		}

		res := <-cmdResult

		if res.Error != nil {
			app.DebugLog("Error: %v\n", res.Error)
			app.DebugLog("Stderr: %v\n", res.Stderr)
			app.DebugLog("ExitCode: %v\n", res.ExitCode)
		}

		app.DebugLog("stdout", "output", meta)

		break
	}

	return nil
}

func main() {
	app := NewApp()

	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	app.DebugLog("loaded configuration", "config", cfg)

	if runErr := run(cfg, app); runErr != nil {
		log.Fatalf("Error: %v\n", runErr)
	}
}
