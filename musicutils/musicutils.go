package musicutils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetAllMusicFiles returns a list of all music files in the specified folder
func GetAllMusicFiles(folder string) []string {
	var files []string
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

			//fmt.Println("Found music file: ", path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking the path %q: %v\n", folder, err)
	}
	return files
}

// GetFilteredMusicFiles returns a list of all music files in the specified folder that match the filter
func GetFilteredMusicFiles(folder string, filter string, maxMB int) []string {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil {

			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".mp3") ||
				strings.HasSuffix(info.Name(), ".flac") ||
				strings.HasSuffix(info.Name(), ".m4a") ||
				strings.HasSuffix(info.Name(), ".wav")) {
				if strings.Contains(strings.ToLower(path), strings.ToLower(filter)) {
					if maxMB > 0 && info.Size() > int64(maxMB*1024*1024) {
						files = append(files, path)
						//fmt.Println("Found music file: ", path)
					} else if maxMB == 0 {
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

func FolderExists(folder string) bool {
	_, err := os.Stat(folder)
	return !os.IsNotExist(err)
}

// Check if a folder is empty
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

func DeleteFile(file string) {
	// If this flag is set, delete the source file
	err := os.Remove(file)
	if err != nil {
		log.Println("Error deleting source file: ", err)
		return
	}

	// Once removed, see if the folder is empty
	// If it is, remove the folders
	dir := filepath.Dir(file)

	// Get the root folder of a path
	root := filepath.VolumeName(dir) + string(filepath.Separator)

	for dir != root {
		empty, err := IsDirEmpty(dir)
		if err == nil {
			if empty {
				log.Println("Deleting empty source folder: ", dir)
				err = os.Remove(dir)
				if err != nil {
					log.Println("Error deleting source folder: ", err)
					return
				}
			}
		} else {
			log.Println("Error checking if folder is empty: ", err)
		}

		dir = filepath.Dir(dir)
	}
}

