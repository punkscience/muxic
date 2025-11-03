// Package musicutils provides utility functions for music file discovery.
// It currently includes functions to get all music files in a folder
// and to filter them based on name and size.
package musicutils

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	taglib "go.senan.xyz/taglib"
)

// GetAllMusicFiles returns a list of all music files in the specified folder.
// It supports .mp3, .flac, .m4a, and .wav files.
func GetAllMusicFiles(folder string) []string {
	files := make([]string, 0) // Initialize as non-nil empty slice
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".mp3") ||
			strings.HasSuffix(info.Name(), ".flac") ||
			strings.HasSuffix(info.Name(), ".m4a") ||
			strings.HasSuffix(info.Name(), ".wav")) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking the path %q: %v\n", folder, err)
	}
	return files
}

// hasSufficientDuration checks if a music file's duration is greater than or equal to a minimum value.
// It returns true if the duration is sufficient or if minDuration is 0 (no filter).
func hasSufficientDuration(path string, minDuration int) bool {
	if minDuration <= 0 {
		return true // No duration filter, so always pass
	}

	properties, err := taglib.ReadProperties(path)
	if err != nil {
		log.Printf("Could not get duration for %s: %v", path, err)
		return false // Exclude files where duration can't be determined
	}

	return int(properties.Length.Minutes()) >= minDuration
}

// GetFilteredMusicFiles returns a list of all music files in the specified folder
// that match the filter string (case-insensitive, path-based) and are larger than maxMB (if maxMB > 0).
// It supports .mp3, .flac, .m4a, and .wav files.
func GetFilteredMusicFiles(folder string, filter string, maxMB int, minDuration int) []string {
	files := make([]string, 0) // Initialize as non-nil empty slice
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".mp3") ||
				strings.HasSuffix(info.Name(), ".flac") ||
				strings.HasSuffix(info.Name(), ".m4a") ||
				strings.HasSuffix(info.Name(), ".wav")) {
				if !strings.Contains(strings.ToLower(path), strings.ToLower(filter)) {
					return nil
				}
				if maxMB > 0 && info.Size() < int64(maxMB*1024*1024) {
					return nil
				}
				if !hasSufficientDuration(path, minDuration) {
					return nil
				}
				files = append(files, path)
			}
		}
		return err
	})
	if err != nil {
		log.Printf("Error walking the path %q: %v\n", folder, err)
	}
	return files
}
