package musicutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wtolson/go-taglib"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetAllMusicFiles returns a list of all music files in the specified folder
func GetAllMusicFiles(folder string) []string {
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
	if strings.HasSuffix(file, ".mp3") || strings.HasSuffix(file, ".flac") || strings.HasSuffix(file, ".wav") {
		tag, err := taglib.Read(file)
		if err != nil {
			fmt.Printf("error opening file %q: %v\n", file, err)
			return targetPath
		}
		defer tag.Close()

		// Format each string in proper title format
		converter := cases.Title(language.English)
		artist := converter.String(tag.Artist())
		album := converter.String(tag.Album())
		title := converter.String(tag.Title())
		track := fmt.Sprintf("%d", tag.Track())

		// Retrieve the desired tag information, e.g., tag.Title(), tag.Artist(), etc.
		targetPath = fmt.Sprintf("%s/%s/%s - %s", artist, album, track, title)
		targetPath = targetPath + filepath.Ext(file)

		// Return the target path name
		return targetPath
	} else {
		log.Println("Unsupported file type with extension: ", filepath.Ext(file))
	}

	return targetPath
}

// FileExists checks to see if the file exists
func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}
