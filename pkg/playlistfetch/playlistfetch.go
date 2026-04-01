// Package playlistfetch provides types and utilities for fetching playlists
// from streaming services and writing them to text files.
package playlistfetch

import (
	"fmt"
	"muxic/pkg/sanitization"
	"os"
	"path/filepath"
)

// Track represents a single track in a playlist.
type Track struct {
	Artist      string
	Album       string
	TrackNumber int
	Title       string
}

// Playlist represents a named playlist containing tracks.
type Playlist struct {
	Name   string
	Tracks []Track
}

// Service is the interface implemented by each streaming service.
type Service interface {
	FetchPlaylists() ([]Playlist, error)
}

// WritePlaylist writes the playlist to <outputDir>/<sanitizedName>.txt.
// Each line is formatted as: "artist - album - trackNumber - title"
func WritePlaylist(pl Playlist, outputDir string) error {
	s := sanitization.NewWindowsSanitizer()
	filename := s.SanitizeFileName(pl.Name) + ".txt"
	path := filepath.Join(outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, t := range pl.Tracks {
		var line string
		if t.Album != "" && t.TrackNumber != 0 {
			line = fmt.Sprintf("%s - %s - %d - %s\n", t.Artist, t.Album, t.TrackNumber, t.Title)
		} else {
			line = fmt.Sprintf("%s - %s\n", t.Artist, t.Title)
		}
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}
