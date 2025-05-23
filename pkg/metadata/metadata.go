// Package metadata provides types and functions for reading metadata from music files.
package metadata

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

// TrackInfo holds metadata extracted from a music file.
type TrackInfo struct {
	Artist            string
	Album             string
	Title             string
	TrackNumber       int
	OriginalExtension string
	SourcePath        string
	Genre             string
	Year              int
}

// ReadTrackInfo extracts metadata from the given audio file.
// It returns a TrackInfo struct populated with available metadata or defaults.
// An error is returned for issues like file not existing or unsupported file type.
// Failure to read tags is logged as a warning but does not return an error,
// allowing the function to proceed with defaults.
func ReadTrackInfo(filePath string) (*TrackInfo, error) {
	// Check if the source file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	} else if err != nil {
		return nil, fmt.Errorf("error checking file %s: %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".m4a":  true,
		".txt":  true, // For testing purposes
	}
	if !supportedExtensions[ext] {
		return nil, fmt.Errorf("unsupported file type: %s for file %s", ext, filePath)
	}

	// Initialize TrackInfo with defaults
	trackInfo := &TrackInfo{
		SourcePath:        filePath,
		OriginalExtension: ext,
		Artist:            "Unknown",
		Album:             "Unknown",
		Title:             strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		TrackNumber:       1,
		Genre:             "Unknown",
		Year:              0,
	}

	// Read metadata tags
	m, err := tag.ReadFrom(file)
	if err != nil {
		// Log warning but proceed with defaults; this is not a fatal error for this function.
		log.Printf("Warning: could not read tags from %s: %v", filePath, err)
	} else {
		// Populate from tags if available
		if m.Artist() != "" {
			trackInfo.Artist = m.Artist()
		}
		if m.Album() != "" {
			trackInfo.Album = m.Album()
		}
		if m.Title() != "" {
			trackInfo.Title = m.Title()
		}
		if trackNum, _ := m.Track(); trackNum > 0 {
			trackInfo.TrackNumber = trackNum
		}
		if m.Genre() != "" {
			trackInfo.Genre = m.Genre()
		}
		if m.Year() > 0 { // Year can be 0 if not set, so only update if positive.
			trackInfo.Year = m.Year()
		}
	}

	return trackInfo, nil
}
