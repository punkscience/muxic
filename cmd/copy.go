/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"muxic/musicutils"

	"os"

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
		sourceFolder := cmd.Flag("source").Value.String()
		targetFolder := cmd.Flag("target").Value.String()

		allFiles := musicutils.GetAllMusicFiles(sourceFolder)

		// Print all the files
		for _, file := range allFiles {
			if destructive {
				fmt.Println("Moving file: ", file)
			} else {
				fmt.Println("Copying file: ", file)
			}

			resultFileName, err := movemusic.CopyMusic(file, targetFolder, true)

			// Check if the file is the same as the result file
			sameFile := resultFileName == file

			if err != nil {
				if err == movemusic.ErrFileExists {
					fmt.Println("File already exists, skipping.")

					if destructive && !sameFile {
						// Delete the source file
						fmt.Println("Deleting source file: ", file)
						err := os.Remove(file)

						if err != nil {
							println("Error deleting file: ", err)
						}
					}
				} else {
					log.Println("Error copying file: ", err)
				}

				continue
			} else if destructive && !sameFile {

				// Delete the source file
				fmt.Println("Deleting source file: ", file)
				err := os.Remove(file)

				if err != nil {
					println("Error deleting file: ", err)
				}
			}

			println("Finished: ", resultFileName)
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
	copyCmd.Flags().BoolVarP(&destructive, "move", "m", false, "Delete the source file after copying")
}
