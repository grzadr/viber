// Package config provides configuration parsing for command-line arguments.
package config

import (
	"errors"
	"flag"
	"fmt"
)

// ErrNoPathsProvided is returned when no path arguments are provided.
var ErrNoPathsProvided = errors.New("at least one path argument is required")

// ArgsParsed contains the parsed command-line arguments.
type ArgsParsed struct {
	Paths []string
}

func (a *ArgsParsed) String() string {
	return fmt.Sprintf("paths: %v", a.Paths)
}

// ParseArgs parses command-line arguments and returns ArgsParsed struct.
func ParseArgs(args []string) (*ArgsParsed, error) {
	flagSet := flag.NewFlagSet("viber", flag.ExitOnError)

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	parsedArgs := flagSet.Args()

	if len(parsedArgs) < 1 {
		return nil, ErrNoPathsProvided
	}

	return &ArgsParsed{
		Paths: parsedArgs,
	}, nil
}
