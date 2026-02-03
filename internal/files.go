package internal

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// extensionToType maps file extensions to their corresponding AudioType.
var extensionToType = map[string]AudioType{
	".flac": FLAC,
	".mp3":  MP3,
	".ogg":  OGG,
	".wav":  WAV,
	".m4a":  M4A,
	".aac":  AAC,
	".wma":  WMA,
	".aiff": AIFF,
	".aif":  AIF,
	".opus": OPUS,
	".ape":  APE,
	".wv":   WV,
}

// AudioPath represents a path to an audio file with its type.
type AudioPath struct {
	Path string
	Type AudioType
}

// GetAudioFiles takes a list of paths and returns all supported audio files.
// If a path is a file, it will be included if supported.
// If a path is a directory, it will be traversed recursively.
// Duplicates are removed from the final result.
func GetAudioFiles(paths []string) ([]AudioPath, error) {
	var result []AudioPath

	for _, path := range paths {
		files, err := processPath(path)
		if err != nil {
			return nil, err
		}
		result = append(result, files...)
	}

	return deduplicateAndSort(result), nil
}

// processPath handles a single path, either adding it directly if it's a file,
// or recursively walking it if it's a directory.
func processPath(path string) ([]AudioPath, error) {
	// Get absolute path to avoid duplicates
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	var result []AudioPath

	if info.IsDir() {
		// Walk the directory recursively
		err := filepath.WalkDir(absPath, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if audioType := getAudioType(p); audioType != Unknown {
				result = append(result, AudioPath{
					Path: p,
					Type: audioType,
				})
			}

			return nil
		})
		return result, err
	}

	// It's a file
	if audioType := getAudioType(absPath); audioType != Unknown {
		result = append(result, AudioPath{
			Path: absPath,
			Type: audioType,
		})
	}

	return result, nil
}

// deduplicateAndSort removes duplicate paths and sorts the result by path.
func deduplicateAndSort(files []AudioPath) []AudioPath {
	if len(files) == 0 {
		return files
	}

	// Sort by path
	slices.SortFunc(files, func(a, b AudioPath) int {
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})

	// Deduplicate consecutive entries
	return slices.CompactFunc(files, func(a, b AudioPath) bool {
		return a.Path == b.Path
	})
}

// getAudioType determines the audio type based on file extension.
func getAudioType(path string) AudioType {
	ext := strings.ToLower(filepath.Ext(path))
	if audioType, ok := extensionToType[ext]; ok {
		return audioType
	}
	return Unknown
}
