package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

type MusicFile struct {
	path string
	size int64
}

func main() {
	replacer := strings.NewReplacer("*", "+",
		"http://", "",
		"@", "at",
		"/", "_",
		"\\", "+",
		"?", "",
		"\"", "'",
		":", "-",
		"|", "-",
		"<", "_",
		">", "_",
		"  ", " ",
		"w/", "with",
		"W/", "with",
		"ft.", "featuring",
		"feat.", "featuring",
		"Feat.", "featuring",
		"FEAT.", "featuring",
		"Feat.", "featuring")

	fmt.Println("MusicProc: A Utility by Punk Science Studios Inc.")

	if len(os.Args) < 3 {
		fmt.Println("Please specify a source and target folder (they can be the same).")
		return
	}

	srcFolder := os.Args[1]
	trgFolder := os.Args[2]

	fmt.Println("Reading files from " + srcFolder + "...")

	files := readFiles(srcFolder)
	fmt.Printf("Read %d files.\n", len(files))

	// Now process all the files
	for _, file := range files {
		processFile(file, trgFolder, replacer)
	}

}

func readFiles(src string) []MusicFile {
	files := []MusicFile{}

	err := filepath.Walk(src,
		func(path string, info os.FileInfo, err error) error {
			if info.Name()[0] == '.' || info.IsDir() == true || strings.ToLower(filepath.Ext(path)) != ".mp3" {
				return nil
			}

			if err != nil {
				fmt.Println(err)
				return err
			}

			musicFile := MusicFile{path: path, size: info.Size()}
			files = append(files, musicFile)

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return files
}

func processFile(file MusicFile, trgPath string, repl *strings.Replacer) error {
	f, err := os.Open(file.path)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Try to read the tag
	m, err := tag.ReadFrom(f)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure the folders exist
	artistPath := path.Join(trgPath, cleanupSymbols(m.Artist(), repl))
	albumPath := path.Join(artistPath, cleanupSymbols(m.Album(), repl))

	if _, err := os.Stat(albumPath); err != nil {
		err := os.MkdirAll(albumPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	newFullPath := buildNewPath(file, trgPath, m, repl)

	/// If it already exists, delete it
	if _, err := os.Stat(newFullPath); err == nil { // TODO: Check file size too

		// Make sure we're not cleaning up a single folder, in which case we'd be deleting everything
		if strings.ToLower(newFullPath) != strings.ToLower(file.path) {
			// fmt.Println("SRC: " + file.path)
			// fmt.Println("DST: " + newFullPath)
			fmt.Println("DELETING " + file.path) // The detected format.
			os.Remove(file.path)
		}
	} else {
		fmt.Printf("MOVING:\n%s\n%s----\n", file.path, newFullPath) // The detected format.
		copyFile(file.path, newFullPath)
		os.Remove(file.path)
	}

	return nil
}

func buildNewPath(file MusicFile, root string, tag tag.Metadata, repl *strings.Replacer) string {
	newPath := root
	extension := filepath.Ext(file.path)

	trackNo, _ := tag.Track()

	artist := cleanupSymbols(tag.Artist(), repl)
	album := cleanupSymbols(tag.Album(), repl)
	title := cleanupSymbols(tag.Title(), repl)
	title = strings.Replace(title, ".mp3", "", -1)

	newPath = path.Join(root, artist)
	newPath = path.Join(newPath, album)
	newPath = path.Join(newPath, strconv.Itoa(trackNo)+" - "+title+extension)

	return newPath
}

func copyFile(src string, dst string) {
	from, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupSymbols(str string, repl *strings.Replacer) string {

	if len(str) == 0 {
		return str
	}

	//fmt.Println(str)

	newString := str

	newString = repl.Replace(newString)
	newString = strings.Trim(newString, "\t\n ")

	// Make sure the string is not empty at this point
	if newString == "" {
		newString = "(untitled)"
	} else if newString[0] == ' ' {
		newString = "(spaces)"
	}

	// Crave off the last character if it's a '.' -- Ubuntu apparently doesn't like this.
	for newString[len(newString)-1] == '.' {
		newString = newString[:len(newString)-1]
		if newString == "" {
			newString = "dot"
		}
	}

	// Print out the change the was made for debugging purposes
	if str != newString {
		fmt.Printf("\"%s\" -> \"%s\"\n", str, newString)
	}

	return newString
}
