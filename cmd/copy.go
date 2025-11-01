// Package cmd implements the command-line interface for the Muxic application.
// It uses the Cobra library for command structure and parsing.
package cmd

import (
	"errors"
	"log"
	"muxic/movemusic"
	"muxic/musicutils"
	"muxic/pkg/filesystem"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var destructive bool
var verbose bool
var dryRun bool

// copyCmd represents the copy command, which handles both copying and moving of music files.
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies or moves music files to a specified destination, organizing them by metadata.",
	Long: `Copies (or moves if --move is specified) music files from a source folder
to a destination folder. It uses metadata (tags) to create an organized folder
layout (Artist/Album/Track). File names are cleaned and standardized.
The --dry-run flag simulates operations without making changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		sourceFolder := strings.Trim(cmd.Flag("source").Value.String(), " ")
		targetFolder := strings.Trim(cmd.Flag("target").Value.String(), " ")
		destructive = cmd.Flag("move").Value.String() == "true"
		verbose = cmd.Flag("verbose").Value.String() == "true"
		dryRun = cmd.Flag("dry-run").Value.String() == "true"
		filter := strings.Trim(cmd.Flag("filter").Value.String(), " ")
		maxMB, _ := strconv.Atoi(cmd.Flag("over").Value.String())
		minDuration, _ := strconv.Atoi(cmd.Flag("duration").Value.String())

		operationType := "Copying"
		if destructive {
			operationType = "Moving"
		}

		if dryRun {
			log.Println("Muxic: Dry-run mode enabled. No actual changes will be made.")
			if verbose {
				log.Printf("[DRY-RUN] Operation: %s", operationType)
				log.Printf("[DRY-RUN] Source: %s", sourceFolder)
				log.Printf("[DRY-RUN] Target: %s", targetFolder)
				if filter != "" {
					log.Printf("[DRY-RUN] Filter: %s", filter)
				}
				if maxMB > 0 {
					log.Printf("[DRY-RUN] Over %d MB", maxMB)
				}
				if minDuration > 0 {
					log.Printf("[DRY-RUN] Duration >= %d minutes", minDuration)
				}
			}
		} else {
			log.Printf("Muxic: %s files from '%s' to '%s'.", operationType, sourceFolder, targetFolder)
		}

		if !filesystem.FolderExists(targetFolder) {
			if dryRun {
				log.Printf("[DRY-RUN] Base target folder '%s' does not exist. Would attempt to create it.", targetFolder)
			} else {
				log.Printf("Base target folder '%s' does not exist. Creating it.", targetFolder)
				if err := os.MkdirAll(targetFolder, os.ModePerm); err != nil {
					log.Fatalf("Failed to create base target folder '%s': %v. Aborting.", targetFolder, err)
					return
				}
			}
		}

		var allFiles []string
		if filter != "" || maxMB > 0 || minDuration > 0 {
			log.Printf("Muxic: Filtering files (filter: '%s', size > %dMB, duration >= %d minutes)...", filter, maxMB, minDuration)
			allFiles = musicutils.GetFilteredMusicFiles(sourceFolder, filter, maxMB, minDuration)
		} else {
			log.Println("Muxic: Scanning all music files...")
			allFiles = musicutils.GetAllMusicFiles(sourceFolder)
		}

		log.Printf("Muxic: Found %d music files. Processing...", len(allFiles))

		processedCount := 0
		errorCount := 0

		for _, file := range allFiles {
			if verbose {
				if dryRun {
					log.Printf("[DRY-RUN] Processing file: %s", file)
				} else {
					log.Printf("Processing file: %s", file)
				}
			}

			var resultFileName string
			var err error

			useFolders := true // TODO: Consider making this a command-line flag if flexibility is needed.

			if destructive {
				resultFileName, err = movemusic.MoveMusic(file, targetFolder, useFolders, dryRun, sourceFolder)
			} else {
				resultFileName, err = movemusic.CopyMusic(file, targetFolder, useFolders, dryRun)
			}

			if err != nil {
				if errors.Is(err, movemusic.ErrFileAlreadyExists) {
					// This is not a critical error, just a skip. Do not increment errorCount.
				} else {
					log.Printf("Error processing file %s: %v", file, err)
					errorCount++
				}
				continue
			}

			if !dryRun {
				log.Printf("Finished %s: %s -> %s", operationType, file, resultFileName)
			} else {
				log.Printf("[DRY-RUN] Simulated %s for: %s -> %s", strings.ToLower(operationType), file, resultFileName)
			}
			processedCount++
		}
		log.Printf("Muxic: Processing complete. %d files processed, %d errors.", processedCount, errorCount)
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	copyCmd.Flags().String("source", "", "The source folder containing music files.")
	copyCmd.Flags().String("target", "", "The destination folder where music files will be organized.")
	copyCmd.Flags().String("filter", "", "Filter files by a string contained in their path (case-insensitive).")
	copyCmd.Flags().Int("over", 0, "Only process files over this size in megabytes (MB).")
	copyCmd.Flags().Int("duration", 0, "Only process files with a duration in minutes greater than or equal to this value.")

	copyCmd.Flags().BoolVarP(&destructive, "move", "m", false, "Move files instead of copying (deletes source files and empty parent dirs).")
	copyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging for detailed operation output.")
	copyCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate operations without making any changes to the file system.")
}
