// Package main is the entry point for the viber application.
package main

import (
	"log"
	"os"

	"github.com/grzadr/viber/internal/config"
	"github.com/grzadr/viber/internal/files"
)

func main() {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("Paths to process: %v\n", cfg.Paths)

	for audioFile, err := range files.GetAudioPaths(cfg.Paths) {
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		log.Println(audioFile)
	}
}
