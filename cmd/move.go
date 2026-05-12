package cmd

import "github.com/spf13/cobra"

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Moves music files to a specified destination, organizing them by metadata.",
	Long: `Moves music files from a source folder to a destination folder. It uses
metadata (tags) to create an organized folder layout (Artist/Album/Track).
File names are cleaned and standardized. Source files and empty parent
directories are deleted after a successful move.
The --dry-run flag simulates operations without making changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMusicOperation(cmd, true)
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)

	moveCmd.Flags().String("source", "", "The source folder containing music files.")
	moveCmd.Flags().String("target", "", "The destination folder where music files will be organized.")
	moveCmd.Flags().String("filter", "", "Filter files by a string contained in their path (case-insensitive).")
	moveCmd.Flags().Int("over", 0, "Only process files over this size in megabytes (MB).")
	moveCmd.Flags().Int("duration", 0, "Only process files with a duration in minutes greater than or equal to this value.")

	moveCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging for detailed operation output.")
	moveCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate operations without making any changes to the file system.")
}
