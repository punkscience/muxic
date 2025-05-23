package movemusic

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"github.com/dhowden/tag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var ErrFileExists = errors.New("file already exists")

func BuildDestinationFileName( sourceFileFullPath string, destFullPath string, useFolders bool ) ( string, error ) {
	// Check if the source file exists
	if _, err := os.Stat(sourceFileFullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("source file does not exist")
	}

	// Get the mp3, flac or wav file details from the file
	file, err := os.Open(sourceFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error opening the file")
	}
	defer file.Close()

	// Get the file extension
	ext := strings.ToLower(filepath.Ext(sourceFileFullPath))
	if ext != ".mp3" && ext != ".flac" && ext != ".wav" {
		return "", fmt.Errorf("unsupported file type")
	}

	// Check if the artist, album, track and track number are empty
	artist := "Unknown"
	album := "Unknown"
	track := strings.TrimSuffix(filepath.Base(sourceFileFullPath), ext)
	trackNumber := 1

	// Open the source file
	m, err := tag.ReadFrom(file)
	if err == nil {
		// Get the artist, album, track and track number

		if m.Artist() != "" {
			artist = m.Artist()
		}

		if m.Album() != "" {
			album = m.Album()
		}

		if m.Title() != "" {
			track = m.Title()
		}

		trackNumber, _ = m.Track()
	}

	// Build a name
	newName := makeFileName(artist, album, track, trackNumber, ext, useFolders)

	// There is a chance this goes off the rails if the tag data was larger than expected
	// and windows has filename limitations to contend with, so let's correct that.
	if len(newName) > 255 {
		log.Println("Filename too long, using the original.")
		newName = filepath.Base(sourceFileFullPath)
	}

	// Build the destination file path
	destFileFullPath := filepath.Join(destFullPath, newName)

	return destFileFullPath, nil
}

func CopyMusic(sourceFileFullPath string, destFolderPath string, useFolders bool) (string, error) {

	// Check if the source file exists
	if _, err := os.Stat(sourceFileFullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("source file does not exist")
	}

	// Check if the destination folder exists
	if _, err := os.Stat(destFolderPath); os.IsNotExist(err) {
		return "", fmt.Errorf("destination folder does not exist")
	}

	destFileFullPath, err := BuildDestinationFileName( sourceFileFullPath, destFolderPath, useFolders )

	if err != nil {
		return "", err
	}

	// Check if the destination file exists
	if _, err := os.Stat(destFileFullPath); err == nil {
		return destFileFullPath, ErrFileExists
	}

	// Copy the file
	sourceFile, err := os.Open(sourceFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error opening the source file: %v", err)
	}
	defer sourceFile.Close()

	// Make sure the destination folder is created
	err = os.MkdirAll(filepath.Dir(destFileFullPath), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating the destination folder: %v", err)
	}

	destFile, err := os.Create(destFileFullPath)
	if err != nil {
		return "", fmt.Errorf("error creating the destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("error copying the file: %v", err)
	}

	return destFileFullPath, nil
}

func makeFileName(artist string, album string, track string, trackNumber int, ext string, useFolders bool) string {
	// Remove the invalid characters from the artist, album and track
	artist = cleanup(artist)
	album = cleanup(album)
	track = cleanup(track)

	// Build the new name
	var newName string

	if useFolders {
		// Build the folder name
		newName = filepath.Join(artist, album)
		newName = filepath.Join(newName, fmt.Sprintf("%02d - %s%s", trackNumber, track, ext))
	} else {
		// Build the folder name
		newName = fmt.Sprintf("%s - %s - %02d - %s%s", artist, album, trackNumber, track, ext)
	}

	// Build the file name
	return newName
}

func cleanup(s string) string {

	s = strings.Trim(s, " \t\n\r\"'")

	// Remove the invalid characters from the artist, album and track
	s = strings.Replace(s, "/", "-", -1)
	s = strings.Replace(s, "\\", "-", -1)
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, "*", "-", -1)
	s = strings.Replace(s, "?", "-", -1)
	s = strings.Replace(s, "\"", "-", -1)
	s = strings.Replace(s, "<", "-", -1)
	s = strings.Replace(s, ">", "-", -1)
	s = strings.Replace(s, "|", "-", -1)
	s = strings.Replace(s, "  ", " ", -1)

	s = strings.Replace(s, "feat.", "ft", -1)
	s = strings.Replace(s, "Feat.", "ft", -1)
	s = strings.Replace(s, "Feat", "ft", -1)
	s = strings.Replace(s, "Featuring", "ft", -1)
	s = strings.Replace(s, "&", "and", -1)

	// Remove any special characters
	s = strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 {
			return r
		}
		return -1
	}, s)

	// Finally, fix the capitalization to proper english title case
	s = cases.Title(language.English).String(s)

	return s
}