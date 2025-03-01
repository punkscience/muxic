/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"muxic/musicutils"
	"strings"

	"github.com/punkscience/movemusic"
	"github.com/spf13/cobra"
)

var destructive bool

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies all music files in a specified folder to a specified destination",
	Long: `Copies all music files from a specified folder into a destination file folder using their
mp3 tag information to create the appropriate folder layout. It also cleans up the capitalization and 
removes any special characters from the file names.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the complete list of files from the source folder

		sourceFolder := strings.Trim(cmd.Flag("source").Value.String(), " ")
		targetFolder := strings.Trim(cmd.Flag("target").Value.String(), " ")
		destructive := cmd.Flag("move").Value.String() == "true"
		verbose := cmd.Flag("verbose").Value.String() == "true"

		// Turn the log off // TODO: Fix this to use a propper logger, this doesn't work.
		if !verbose {
			log.SetOutput(io.Discard)
		}

		if destructive {
			fmt.Println("Muxic: Destructive mode is on. Source files will be deleted after copying.")
		}

		fmt.Println("Muxic: Scanning all files from: ", sourceFolder)

		allFiles := musicutils.GetAllMusicFiles(sourceFolder)

		fmt.Println("Muxic: Found ", len(allFiles), " music files. Processing...")

		// Print all the files
		for _, file := range allFiles {

			if verbose {
				log.Println("Copying file: ", file)
			}

			resultFileName, err := movemusic.CopyMusic(file, targetFolder, true)

			if err != nil {
				if err == movemusic.ErrFileExists {
					log.Println("EXISTS: File already exists, skipping ", file)
				} else {
					log.Println("Error copying file: ", err)
					continue
				}
			}

			if destructive {
				// Only delete if they are not the same file entirely.
				if !strings.EqualFold(file, resultFileName) {
					// Delete the source file
					log.Println("Deleting source file: ", file)
					musicutils.DeleteFile(file)
				}
			}

			log.Println("Finished: ", resultFileName)
		}
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	copyCmd.Flags().String("source", "", "The source folder name")
	copyCmd.Flags().String("target", "", "The destination folder name")

	copyCmd.Flags().BoolVarP(&destructive, "move", "m", false, "Move, don't copy -- delete the source file after copying")
	copyCmd.Flags().BoolVarP(&destructive, "verbose", "v", false, "Log everything.")
}
