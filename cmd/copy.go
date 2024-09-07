/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"muxic/musicutils"
	"path/filepath"

	"github.com/spf13/cobra"
)

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
		destructive := cmd.Flag("destructive").Value.String() == "true"
		allFiles := musicutils.GetAllMusicFiles(sourceFolder)

		// Print all the files
		for _, file := range allFiles {
			// Make a target path name out of the file tag
			targetPathName := musicutils.GetTargetPathName(file)
			targetPathName = filepath.Join(targetFolder, targetPathName)

			// Check to see if the target file already exists
			// If it does, skip the file
			if musicutils.FileExists(targetPathName) {
				fmt.Printf("EXISTS: %s, skipping it...\n", targetPathName)
				continue
			} else {
				fmt.Printf("Copying %s to %s\n", file, targetPathName)
				musicutils.CopyFile(file, targetPathName, destructive)
			}
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
	copyCmd.Flags().String("source", "s", "The source folder name")
	copyCmd.Flags().String("target", "t", "The destination folder name")
	copyCmd.Flags().BoolP("destructive", "d", false, "Delete source file once target is verified")
}
