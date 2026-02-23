// Package main is the entry point for the viber application.
package main

import (
	"context"
	"log"
	"os"

	"github.com/grzadr/viber/internal/config"
	"github.com/grzadr/viber/internal/files"
	"github.com/grzadr/viber/internal/sub"
)

func main() {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("Paths to process: %v\n", cfg.Paths)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for audioFile, err := range files.GetAudioPaths(cfg.Paths) {
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		log.Println(audioFile)

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
			log.Printf("Error: %v\n", cmdErr)
		}

		for line := range stdout {
			log.Println(line)
		}

		for line := range stderr {
			log.Println(line)
		}

		break
	}
}
