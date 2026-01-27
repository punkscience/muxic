// Package movemusic provides the core logic for copying or moving music files
// based on their metadata. It handles filename sanitization, path generation,
// and file operations, utilizing the metadata and filesystem packages.
package movemusic

import (
	"errors"
	"fmt"
	"io"
	"log"
	"muxic/pkg/filesystem"
	"muxic/pkg/metadata"
	"muxic/pkg/sanitization"
	"os"
	"path/filepath"
)

// ErrFileAlreadyExists is returned when the destination file for a copy or move operation already exists.
var ErrFileAlreadyExists = errors.New("file already exists")

// sanitizer is the global sanitizer instance used for cleaning metadata
var sanitizer = sanitization.NewWindowsSanitizer()

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

	// Optimization: The destination folder is already guaranteed to exist by the caller (cmd/copy.go).
	// Removing the redundant os.Stat call to improve performance in the loop.

	destFileFullPath, err := SuggestDestinationPath(destFolderPath, useFolders, trackInfo)
	if err != nil {
		return "", fmt.Errorf("error suggesting destination path: %w", err)
	}

	// Check for self-move: if the source and calculated destination are the same file.
	// This happens when running muxic on an already organized directory.
	cleanSource, _ := filepath.Abs(sourceFileFullPath)
	cleanDest, _ := filepath.Abs(destFileFullPath)
	if cleanSource == cleanDest {
		log.Printf("IDENTICAL: Source and destination are the same, skipping %s", sourceFileFullPath)
		return destFileFullPath, ErrFileAlreadyExists
	}

	if filesystem.FileExists(destFileFullPath) {
		log.Printf("EXISTS: File already exists, skipping %s", sourceFileFullPath)
		return destFileFullPath, ErrFileAlreadyExists
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
		if errors.Is(err, ErrFileAlreadyExists) {
			return copiedFilePath, nil // Not an error for move operation, just skip deleting
		}
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
// It uses the sanitized artist, album, title, track number, and original extension.
// If useFolders is true, the format is "Artist/Album/TrackNum - Title.ext";
// otherwise, it's "Artist - Album - TrackNum - Title.ext".
func makeFileName(trackInfo *metadata.TrackInfo, useFolders bool) string {
	// Use the new sanitization system for Windows filesystem compatibility
	artist, album, title := sanitizer.SanitizeTrackMetadata(
		trackInfo.Artist,
		trackInfo.Album,
		trackInfo.Title,
	)

	var newName string
	if useFolders {
		newName = filepath.Join(artist, album, fmt.Sprintf("%02d - %s%s", trackInfo.TrackNumber, title, trackInfo.OriginalExtension))
	} else {
		newName = fmt.Sprintf("%s - %s - %02d - %s%s", artist, album, trackInfo.TrackNumber, title, trackInfo.OriginalExtension)
	}
	return newName
}
