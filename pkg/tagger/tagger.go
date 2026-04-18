// Package tagger writes arbitrary word-labels to a custom MUXIC_TAGS metadata field
// on audio files. It uses TagLib's property map, which handles format-specific
// storage automatically (TXXX frames for MP3/WAV, Vorbis comments for FLAC,
// freeform atoms for M4A).
package tagger

import (
	"fmt"
	"strings"

	taglib "go.senan.xyz/taglib"
)

// TagKey is the property map key used to store muxic labels.
const TagKey = "MUXIC_TAGS"

// AddTag adds word to the MUXIC_TAGS field of the audio file at filePath.
// If the word is already present (case-insensitive) it returns false, nil.
// If dryRun is true the file is not modified.
func AddTag(filePath, word string, dryRun bool) (changed bool, err error) {
	tags, err := taglib.ReadTags(filePath)
	if err != nil {
		return false, fmt.Errorf("read tags %q: %w", filePath, err)
	}

	existing := tags[TagKey]
	for _, v := range existing {
		if strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(word)) {
			return false, nil
		}
	}

	if dryRun {
		return true, nil
	}

	updated := append(existing, word)
	if err := taglib.WriteTags(filePath, map[string][]string{TagKey: updated}, 0); err != nil {
		return false, fmt.Errorf("write tags %q: %w", filePath, err)
	}
	return true, nil
}

// ReadTags returns the MUXIC_TAGS words for the audio file at filePath.
func ReadTags(filePath string) ([]string, error) {
	tags, err := taglib.ReadTags(filePath)
	if err != nil {
		return nil, fmt.Errorf("read tags %q: %w", filePath, err)
	}
	return tags[TagKey], nil
}
