// Package filesystem provides utility functions for interacting with the file system.
package filesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings" // Added strings import
)

// FolderExists checks if a folder exists and is a directory.
func FolderExists(folder string) bool {
	info, err := os.Stat(folder)
	if os.IsNotExist(err) {
		return false
	}
	// Ensure it's a directory
	return err == nil && info.IsDir()
}

// IsDirEmpty checks if a directory is empty.
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// FileExists checks if a file exists and is not a directory.
func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// DeleteFileAndPruneParents deletes a file and then recursively deletes parent directories
// if they become empty, stopping at rootDir.
// If dryRun is true, it returns a list of actions that would be taken.
func DeleteFileAndPruneParents(file string, rootDir string, dryRun bool) ([]string, error) {
	var actions []string
	cleanedRootDir := filepath.Clean(rootDir)

	if !FileExists(file) { // Check if file exists before attempting to delete
		if dryRun {
			actions = append(actions, fmt.Sprintf("File not found, would not delete: %s", file))
			return actions, nil // Nothing to do if file doesn't exist
		}
		return nil, fmt.Errorf("file %s does not exist or is a directory", file)
	}

	if dryRun {
		actions = append(actions, fmt.Sprintf("Would delete file: %s", file))
	} else {
		err := os.Remove(file)
		if err != nil {
			log.Println("Error deleting file: ", err)
			return nil, err
		}
	}

	pathBeingConsidered := file // Start with the file itself for parent pruning logic

	for {
		currentDir := filepath.Dir(pathBeingConsidered)

		// Stop conditions
		if currentDir == cleanedRootDir || !strings.HasPrefix(currentDir, cleanedRootDir) || currentDir == "." || currentDir == "/" || filepath.Clean(currentDir) == filepath.VolumeName(currentDir)+string(filepath.Separator) || currentDir == filepath.Dir(cleanedRootDir) {
			break
		}

		if dryRun {
			// Simulate directory emptiness check for dry run
			dirEntries, err := os.ReadDir(currentDir)
			if err != nil {
				actions = append(actions, fmt.Sprintf("Could not read directory %s to determine emptiness due to error: %v. Stopping pruning for this path.", currentDir, err))
				break
			}

			wouldBeEmpty := true
			if len(dirEntries) == 0 { // Already empty (shouldn't happen if we just "deleted" from it, but good check)
				wouldBeEmpty = true
			} else if len(dirEntries) == 1 && filepath.Clean(filepath.Join(currentDir, dirEntries[0].Name())) == filepath.Clean(pathBeingConsidered) {
				// The directory contains only the item we are "deleting" in this dry run iteration
				wouldBeEmpty = true
			} else if len(dirEntries) > 1 {
				// Contains other items beyond the one we are "deleting"
				wouldBeEmpty = false
			} else {
				wouldBeEmpty = false
			}

			if wouldBeEmpty {
				actions = append(actions, fmt.Sprintf("Would delete empty directory: %s", currentDir))
				pathBeingConsidered = currentDir // For the next iteration, check parent of this dir
			} else {
				actions = append(actions, fmt.Sprintf("Directory %s is not empty, would not delete.", currentDir))
				break // Stop pruning if a directory would not be empty
			}
		} else { // Not a dry run
			empty, err := IsDirEmpty(currentDir)
			if err != nil {
				return nil, fmt.Errorf("error checking if directory %s is empty: %w", currentDir, err)
			}

			if empty {
				log.Println("Deleting empty source folder: ", currentDir)
				err = os.Remove(currentDir)
				if err != nil {
					log.Println("Error deleting source folder: ", currentDir, err)
					return nil, err
				}
				pathBeingConsidered = currentDir // Continue to parent of deleted directory
			} else {
				break // Stop if directory is not empty
			}
		}
	}
	return actions, nil
}
