package files

import (
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
)

// AudioType represents the type of audio file based on its extension.
//
//go:generate stringer -type=AudioType
type AudioType int

const (
	// Unknown represents an unsupported or unknown audio file type.
	Unknown AudioType = iota
	FLAC
	MP3
	OGG
	WAV
	M4A
	AAC
	WMA
	AIFF
	AIF
	OPUS
	APE
	WV
)

// extensionToType maps file extensions to their corresponding AudioType.
func extensionToType(ext string) AudioType {
	ext = strings.ToLower(ext)
	switch ext {
	case ".flac":
		return FLAC
	case ".mp3":
		return MP3
	case ".ogg":
		return OGG
	case ".wav":
		return WAV
	case ".m4a":
		return M4A
	case ".aac":
		return AAC
	case ".wma":
		return WMA
	case ".aiff":
		return AIFF
	case ".aif":
		return AIF
	case ".opus":
		return OPUS
	case ".ape":
		return APE
	case ".wv":
		return WV
	default:
		return Unknown
	}
}

// AudioPath represents a path to an audio file with its type.
type AudioPath struct {
	Path string
	Type AudioType
}

type AudioPathError struct {
	Path string
	Err  error
}

// GetAudioPaths returns an iterator that yields AudioPath objects and errors
// found while traversing the provided list of root paths.
func GetAudioPaths(paths []string) iter.Seq2[AudioPath, error] {
	return func(yield func(AudioPath, error) bool) {
		for _, root := range paths {
			if !walkSingleRoot(root, yield) {
				return
			}
		}
	}
}

// walkSingleRoot processes a single root path, converting it to an absolute
// path
// and walking its directory tree to find supported audio files.
func walkSingleRoot(root string, yield func(AudioPath, error) bool) bool {
	absPath, err := filepath.Abs(root)
	if err != nil {
		// Wrap error with the path context
		return yield(
			AudioPath{},
			fmt.Errorf("absolute path failed for %s: %w", root, err),
		)
	}

	err = filepath.WalkDir(
		absPath,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Wrap walk error with path context
				if !yield(
					AudioPath{},
					fmt.Errorf("error at %s: %w", path, err),
				) {
					return filepath.SkipAll
				}
				return nil
			}

			// Pruning: Skip hidden directories
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			// Process only regular files
			if !d.Type().IsRegular() {
				return nil
			}

			if aType := extensionToType(filepath.Ext(path)); aType != Unknown {
				if !yield(AudioPath{Path: path, Type: aType}, nil) {
					return filepath.SkipAll
				}
			}
			return nil
		},
	)

	return err == nil || errors.Is(err, filepath.SkipAll)
}
