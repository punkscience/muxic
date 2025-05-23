package cmd

import (
	"fmt"
	"log"
	"muxic/musicutils"
	"muxic/movemusic"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var destructive bool
var verbose bool
var dryRun bool

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
		destructive = cmd.Flag("move").Value.String() == "true"
		verbose = cmd.Flag("verbose").Value.String() == "true"
		dryRun = cmd.Flag("dry-run").Value.String() == "true"
		filter := strings.Trim(cmd.Flag("filter").Value.String(), " ")
		maxMB, _ := strconv.Atoi(cmd.Flag("over").Value.String())

		// Convert that to an int

		if dryRun {
			fmt.Println("Muxic: Dry-run mode enabled. No actual changes will be made.")
		}

		if destructive && !dryRun { // Only print if not a dry run, dry run has its own message
			fmt.Println("Muxic: Destructive mode is on. Source files will be deleted after copying.")
		}

		fmt.Println("Muxic: Scanning all files from: ", sourceFolder)

		var allFiles []string

		if filter != "" || maxMB > 0 {
			fmt.Println("Muxic: Filtering files to those containing: ", filter, " and size > ", maxMB, "MB")
			allFiles = musicutils.GetFilteredMusicFiles(sourceFolder, filter, maxMB)
		} else {
			allFiles = musicutils.GetAllMusicFiles(sourceFolder)
		}

		fmt.Println("Muxic: Found ", len(allFiles), " music files. Processing...")

		// Print all the files
		for _, file := range allFiles {

			if verbose || dryRun {
				log.Println("Processing file: ", file)
			}

			// Check to see if the target folder exists and if not, create it
			if !musicutils.FolderExists(targetFolder) {
				if dryRun {
					log.Println("[DRY-RUN] Would create target folder: ", targetFolder)
				} else {
					log.Println("Creating target folder: ", targetFolder)
					err := os.MkdirAll(targetFolder, os.ModePerm)
					if err != nil {
						log.Println("Error creating target folder: ", err)
						continue
					}
				}
			}

			var resultFileName string
			var err error

			if dryRun {
				log.Printf("[DRY-RUN] Would attempt to process/copy music file '%s' to target folder '%s'\n", file, targetFolder)
				resultFileName, err = movemusic.BuildDestinationFileName( file, targetFolder, true )

				if err != nil {
					log.Printf( "[DRY-RUN] There was an error building the file name.")
				}

				log.Printf("[DRY-RUN] Simulated target path would be: %s\n", resultFileName)
				// Simulate checking for existing file - for now, assume it doesn't exist to show full dry-run path
				// if musicutils.FileExists(resultFileName) { err = movemusic.ErrFileExists } else { err = nil }
				err = nil
			} else {
				resultFileName, err = movemusic.CopyMusic(file, targetFolder, true)
			}

			if err != nil {
				if err == movemusic.ErrFileExists {
					log.Println("EXISTS: File already exists, skipping ", file)
				} else {
					log.Println("Error copying file: ", err)
				}
				continue
			}

			if !dryRun {
				log.Println("Finished: ", resultFileName)
			}

			if destructive {
				if dryRun {
					if !strings.EqualFold(file, resultFileName) {
						log.Println("[DRY-RUN] Would delete source file: ", file)
						// Simulate the empty folder deletion logic of musicutils.DeleteFile
						log.Println("[DRY-RUN] Would then check parent directories of", file, "for emptiness and potential deletion.")
						// Simplified simulation of parent directory cleanup
						// currentPath := file
						// for {
						// 	parentDir := filepath.Dir(currentPath)
						// 	if parentDir == "." || parentDir == "/" || parentDir == filepath.Dir(sourceFolder) {
						// 		break
						// 	}
						//  isEmpty, _ := musicutils.IsDirEmpty(parentDir) // This would be a real check
						// 	log.Printf("[DRY-RUN] Would check if directory %s is empty. Simulated: true\n", parentDir)
						// 	log.Printf("[DRY-RUN] Assuming directory %s is empty, would delete it.\n", parentDir)
						// 	currentPath = parentDir
						// 	if strings.EqualFold(parentDir, sourceFolder){ // Stop if we reach the source folder itself
						//      break
						//  }
						// }
					} else {
						log.Println("[DRY-RUN] Source and (simulated) target are the same, would not delete: ", file)
					}
				} else {
					// Original destructive logic
					if !strings.EqualFold(file, resultFileName) {
						log.Println("Deleting source file: ", file)
						musicutils.DeleteFile(file) // This still performs actual deletions
					}
				}
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
	copyCmd.Flags().String("source", "", "The source folder name")
	copyCmd.Flags().String("target", "", "The destination folder name")
	copyCmd.Flags().String("filter", "", "Filter the files to copy by name. Case insensitive.")
	copyCmd.Flags().Int("over", 0, "Only copy files over this size.")

	copyCmd.Flags().BoolVarP(&destructive, "move", "m", false, "Move, don't copy -- delete the source file after copying")
	copyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log everything.")
	copyCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Report actions that would be taken without executing them.")

}
