// Package musicutils provides utility functions for music file discovery.
// It currently includes functions to get all music files in a folder
// and to filter them based on name and size.
package musicutils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
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

// GetFilteredMusicFiles returns a list of all music files in the specified folder
// that match the filter string (case-insensitive, path-based) and are larger than maxMB (if maxMB > 0).
// It supports .mp3, .flac, .m4a, and .wav files.
func GetFilteredMusicFiles(folder string, filter string, maxMB int) []string {
	files := make([]string, 0) // Initialize as non-nil empty slice
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".mp3") ||
				strings.HasSuffix(info.Name(), ".flac") ||
				strings.HasSuffix(info.Name(), ".m4a") ||
				strings.HasSuffix(info.Name(), ".wav")) {
				if strings.Contains(strings.ToLower(path), strings.ToLower(filter)) {
					if maxMB > 0 {
						if info.Size() >= int64(maxMB*1024*1024) {
							files = append(files, path)
						}
					} else { // maxMB == 0, so no size filtering
						files = append(files, path)
					}
				}
			}
		}
		return err
	})
	if err != nil {
		log.Printf("Error walking the path %q: %v\n", folder, err)
	}
	return files
}
