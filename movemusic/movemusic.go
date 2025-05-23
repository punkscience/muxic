// Package movemusic provides the core logic for copying or moving music files
// based on their metadata. It handles filename sanitization, path generation,
// and file operations, utilizing the metadata and filesystem packages.
package movemusic

import (
	"fmt"
	"io"
	"log"
	"muxic/pkg/filesystem"
	"muxic/pkg/metadata"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// specificSubstitutions holds rules for specific string replacements.
// These are applied before general cleanup and title casing.
var specificSubstitutions = map[string]string{
	"feat.":     "ft",
	"Feat.":     "ft",
	"Feat":      "ft",
	"Featuring": "ft",
	"&":         "and",
}

// SuggestDestinationPath suggests a destination path for a music file based on its metadata.
// It uses trackInfo to generate a filename and combines it with the destBaseFolder.
// useFolders determines if the path includes Artist/Album subdirectories.
// Filenames longer than 255 characters are truncated to the base name of the original source file.
func SuggestDestinationPath(destBaseFolder string, useFolders bool, trackInfo *metadata.TrackInfo) (string, error) {
	if trackInfo == nil {
		return "", fmt.Errorf("trackInfo cannot be nil")
	}

	newName := makeFileName(trackInfo, useFolders)

	if len(newName) > 255 {
		log.Println("Warning: Generated filename too long, using original base filename from source path.")
		newName = filepath.Base(trackInfo.SourcePath)
	}

	destFileFullPath := filepath.Join(destBaseFolder, newName)
	return destFileFullPath, nil
}

// CopyMusic copies a music file from sourceFileFullPath to a new location within destFolderPath.
// The new location is determined by the file's metadata and the useFolders flag.
// If dryRun is true, it logs the intended operation without performing file system changes.
func CopyMusic(sourceFileFullPath string, destFolderPath string, useFolders bool, dryRun bool) (string, error) {
	trackInfo, err := metadata.ReadTrackInfo(sourceFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error reading track info for %s: %w", sourceFileFullPath, err)
	}

	if _, statErr := os.Stat(destFolderPath); os.IsNotExist(statErr) {
		return "", fmt.Errorf("destination folder does not exist: %s", destFolderPath)
	} else if statErr != nil {
		return "", fmt.Errorf("error checking destination folder %s: %w", destFolderPath, statErr)
	}

	destFileFullPath, err := SuggestDestinationPath(destFolderPath, useFolders, trackInfo)
	if err != nil {
		return "", fmt.Errorf("error suggesting destination path: %w", err)
	}

	if dryRun {
		log.Printf("[DRY-RUN] Would copy %s to %s", sourceFileFullPath, destFileFullPath)
		return destFileFullPath, nil
	}

	sourceFile, err := os.Open(sourceFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error opening the source file %s: %w", sourceFileFullPath, err)
	}
	defer sourceFile.Close()

	if err = os.MkdirAll(filepath.Dir(destFileFullPath), os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating destination folder structure %s: %w", filepath.Dir(destFileFullPath), err)
	}

	destFile, err := os.Create(destFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error creating destination file %s: %w", destFileFullPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		if removeErr := os.Remove(destFileFullPath); removeErr != nil {
			log.Printf("Warning: failed to remove partially written file %s after copy error: %v", destFileFullPath, removeErr)
		}
		return "", fmt.Errorf("error copying data from %s to %s: %w", sourceFileFullPath, destFileFullPath, err)
	}

	return destFileFullPath, nil
}

// MoveMusic copies a music file to a new location and then deletes the source file and prunes empty parent directories.
// sourceLibraryRootDir specifies the root directory up to which parent directories of the source file may be pruned.
// If dryRun is true, operations are logged but not executed.
func MoveMusic(sourceFileFullPath string, destFolderPath string, useFolders bool, dryRun bool, sourceLibraryRootDir string) (string, error) {
	copiedFilePath, err := CopyMusic(sourceFileFullPath, destFolderPath, useFolders, dryRun)
	if err != nil {
		return copiedFilePath, err
	}

	if dryRun {
		log.Printf("[DRY-RUN] Source file %s would be deleted. Parent directory pruning up to %s would be simulated.", sourceFileFullPath, sourceLibraryRootDir)
		actions, simErr := filesystem.DeleteFileAndPruneParents(sourceFileFullPath, sourceLibraryRootDir, true)
		if simErr != nil {
			log.Printf("[DRY-RUN] Error during simulated deletion of %s: %v", sourceFileFullPath, simErr)
		}
		if len(actions) > 0 {
			log.Println("[DRY-RUN] Simulated delete actions:", actions)
		} else if simErr == nil {
			log.Println("[DRY-RUN] No simulated delete actions for", sourceFileFullPath)
		}
	} else {
		log.Printf("Deleting source file %s and pruning parent directories up to %s.", sourceFileFullPath, sourceLibraryRootDir)
		_, delErr := filesystem.DeleteFileAndPruneParents(sourceFileFullPath, sourceLibraryRootDir, false)
		if delErr != nil {
			log.Printf("Error deleting source file %s or pruning parents: %v. The file was successfully copied to %s.", sourceFileFullPath, delErr, copiedFilePath)
		}
	}
	return copiedFilePath, nil
}

// makeFileName generates a filename string based on track metadata.
// It uses the cleaned artist, album, title, track number, and original extension.
// If useFolders is true, the format is "Artist/Album/TrackNum - Title.ext";
// otherwise, it's "Artist - Album - TrackNum - Title.ext".
func makeFileName(trackInfo *metadata.TrackInfo, useFolders bool) string {
	artist := cleanup(trackInfo.Artist)
	album := cleanup(trackInfo.Album)
	title := cleanup(trackInfo.Title)

	var newName string
	if useFolders {
		newName = filepath.Join(artist, album, fmt.Sprintf("%02d - %s%s", trackInfo.TrackNumber, title, trackInfo.OriginalExtension))
	} else {
		newName = fmt.Sprintf("%s - %s - %02d - %s%s", artist, album, trackInfo.TrackNumber, title, trackInfo.OriginalExtension)
	}
	return newName
}

// cleanup sanitizes a string for use in file or directory names.
// It trims whitespace, replaces reserved characters, performs specific substitutions (e.g., "feat." to "ft"),
// removes non-printable ASCII characters, and applies title casing.
func cleanup(s string) string {
	s = strings.TrimSpace(s)

	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		s = strings.ReplaceAll(s, char, "-")
	}

	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	for key, value := range specificSubstitutions {
		s = strings.ReplaceAll(s, key, value)
	}

	s = strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 { // Printable ASCII range
			return r
		}
		return -1
	}, s)

	s = cases.Title(language.English).String(s)
	return s
}
