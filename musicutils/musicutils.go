package musicutils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetAllMusicFiles returns a list of all music files in the specified folder
func GetAllMusicFiles(folder string) []string {
	fmt.Printf("Scanning all music files in folder %s ...\n", folder)
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error accessing path %q: %v\n", path, err)
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
		fmt.Printf("error walking the path %q: %v\n", folder, err)
	}
	return files
}

// GetTargetPathName returns the target path name for the file
func GetTargetPathName(file string) string {
	targetPath := ""

	// Get the music file tag info
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("error opening file %q: %v\n", file, err)
		return targetPath
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		fmt.Printf("error reading tags from file %q: %v\n", file, err)
		return targetPath
	}

	// Format each string in proper title format
	converter := cases.Title(language.English)
	artist := converter.String(m.Artist())
	album := converter.String(m.Album())
	title := converter.String(m.Title())

	trackNo, _ := m.Track()
	track := fmt.Sprintf("%d", trackNo)

	// Retrieve the desired tag information, e.g., tag.Title(), tag.Artist(), etc.
	targetPath = fmt.Sprintf("%s/%s/%s - %s", artist, album, track, title)
	targetPath = targetPath + filepath.Ext(file)

	// Return the target path name
	return targetPath
}

// FileExists checks to see if the file exists
func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
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
	fmt.Println("Deleting source file: ", file)
	err := os.Remove(file)
	if err != nil {
		log.Println("Error deleting source file: ", err)
		return
	}

	// Once removed, see if the folder is empty
	// If it is, remove the folder
	dir := filepath.Dir(file)
	empty, err := IsDirEmpty(dir)
	if err != nil {
		if empty {
			fmt.Println("Deleting empty source folder: ", dir)
			err = os.Remove(dir)
			if err != nil {
				log.Println("Error deleting source folder: ", err)
				return
			}

			// Now see if the parent artist folder is empty
			dir = filepath.Dir(dir)
			empty, err = IsDirEmpty(dir)

			if err != nil {
				if empty {
					fmt.Println("Deleting empty source folder: ", dir)
					err = os.Remove(dir)
					if err != nil {
						log.Println("Error deleting source artist folder: ", err)
						return
					}
				}
			}
		}
	}
}
