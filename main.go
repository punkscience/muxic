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
	ext string 
}

func main() {
	// Set up all the things we need to filter out for proper filenames
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
		"Ft.", "featuring",
		"feat.", "featuring",
		"Feat.", "featuring",
		"FEAT.", "featuring",
		"Feat.", "featuring",
		"12\"", "12 Inch",
		"E.P.", "EP")

	fmt.Printf("muxic v1.0.0\nA Utility by Punk Science Studios Inc.\n\n")

	if len(os.Args) < 3 {
		fmt.Println("Please specify a source and target folder (they can be the same).")
		return
	}

	srcFolder := os.Args[1]
	trgFolder := os.Args[2]

	bNonDestructive := false
	if len( os.Args ) >= 4 {
		bNonDestructive = os.Args[3] == "-n"
	}

	fmt.Printf("Reading files from " + srcFolder )
	if bNonDestructive == true {
		fmt.Printf(" in non-destructive mode...\n")
	} else {
		fmt.Printf("...\n")
	}

	files := scanFiles(srcFolder)
	fmt.Printf("Read %d files.\n", len(files))

	// Now process all the files
	for _, file := range files {
		processFile(file, trgFolder, replacer, bNonDestructive )
	}

	fmt.Println("You're all set. Enjoy.")

}

// Scans all files in the folder and returns a list in MusicFile format.
func scanFiles(src string) []MusicFile {
	files := []MusicFile{}

	err := filepath.Walk(src,
		func(path string, info os.FileInfo, err error) error {
			// Grab the current folder name
			currentFolder := filepath.Base(path)

			// If it's a folder, or a hidden folder, we don't care about it
			if info.IsDir() == true  {
				if currentFolder[0] == '.' {
					fmt.Println("Ignoring hidden folder " + currentFolder )
					return filepath.SkipDir
				}
				
				return nil
			}

			// If it's a hidden file, we don't care about it
			if info.Name()[0] == '.' {
				return nil
			}

			if err != nil {
				fmt.Println(err)
				return err
			}

			// We're only worried about the files we support
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mp3" || ext == ".flac" || ext == "m4a" {
				musicFile := MusicFile{path: path, size: info.Size(), ext: ext}
				files = append(files, musicFile)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return files
}

func processFile(file MusicFile, trgPath string, repl *strings.Replacer, bNonDestructive bool ) error {
	f, err := os.Open(file.path)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Try to read the tag
	m, err := tag.ReadFrom(f)
	if err != nil {
		// If we fail to get the tags, let's just move it as is
		log.Println("No tags found for " + file.path + ". Moving as is.")
		trg := trgPath
		trg = filepath.Join(trg, filepath.Base(file.path))
		copyFile(file.path, trg)

		if bNonDestructive == false {
			os.Remove(file.path)
		}
		
		return nil
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
			
			if bNonDestructive == false {
				fmt.Println("DELETING " + file.path) // The detected format.
				os.Remove(file.path)			
			} else {
				fmt.Println("WOULD HAVE DELETED " + file.path) // The detected format.
			}
			
		}
	} else {
		fmt.Printf("MOVING:\n%s\n%s\n----\n", file.path, newFullPath) // The detected format.
		copyFile(file.path, newFullPath)

		if bNonDestructive == false {
			os.Remove(file.path)
		}		
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
	title = strings.Replace(title, ".flac", "", -1)
	title = strings.Replace(title, ".m4a", "", -1)

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
	// if str != newString {
	// 	fmt.Printf("\"%s\" -> \"%s\"\n", str, newString)
	// }

	return newString
}
