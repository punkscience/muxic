package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"muxic/pkg/dedup"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	targetDir     string
	scorchedEarth bool
)

var musicExtensions = []string{".mp3", ".flac", ".m4a", ".wav"}

var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Find and remove duplicate music files",
	Long: `Scans the target directory for music files with identical binary content.
Only files that share a size are read, and only fully when their first and last 16 KiB match.
Offers interactive or automatic (scorched earth) deletion.`,
	Run: func(cmd *cobra.Command, args []string) {
		if targetDir == "" {
			fmt.Println("Error: --target flag is required")
			os.Exit(1)
		}
		if err := runDedup(targetDir, scorchedEarth, os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dedupCmd)
	dedupCmd.Flags().StringVar(&targetDir, "target", "", "Target directory to scan for duplicates")
	dedupCmd.Flags().BoolVar(&scorchedEarth, "scorchedearth", false, "Automatically delete duplicates, keeping the most recently modified")
}

func runDedup(targetDir string, scorchedEarth bool, stdin io.Reader, stdout io.Writer) error {
	fmt.Fprintf(stdout, "Scanning %s...\n", targetDir)
	result, err := dedup.Find(targetDir, dedup.Options{
		Extensions: musicExtensions,
		Workers:    runtime.NumCPU(),
		Warn:       func(err error) { fmt.Fprintf(stdout, "Warning: %v\n", err) },
	})
	if err != nil {
		return fmt.Errorf("error scanning target directory: %w", err)
	}
	fmt.Fprintf(stdout, "Scan complete. %d music files, %d duplicate sets.\n", result.Scanned, len(result.Groups))

	if len(result.Groups) == 0 {
		fmt.Fprintln(stdout, "No duplicates found.")
		return nil
	}

	reader := bufio.NewReader(stdin)
	var bytesSaved int64
	for _, files := range result.Groups {
		fmt.Fprintln(stdout, "\nDuplicate set found:")
		for i, f := range files {
			fmt.Fprintf(stdout, "%d) %s\n", i+1, f.Path)
		}

		var keepIndex int
		if scorchedEarth {
			keepIndex = newest(files)
			fmt.Fprintf(stdout, "Scorched Earth: keeping %s\n", files[keepIndex].Path)
		} else {
			keepIndex = askKeep(reader, stdout, len(files))
		}
		if keepIndex < 0 {
			continue
		}
		bytesSaved += deleteOthers(files, keepIndex, stdout)
	}

	fmt.Fprintf(stdout, "Cleanup complete. Saved approx %.2f MB\n", float64(bytesSaved)/(1024*1024))
	return nil
}

func newest(files dedup.Group) int {
	keep := 0
	for i, f := range files {
		if f.ModTime.After(files[keep].ModTime) {
			keep = i
		}
	}
	return keep
}

// askKeep returns the index of the file to keep, or -1 to leave the set alone.
func askKeep(reader *bufio.Reader, stdout io.Writer, n int) int {
	for {
		fmt.Fprint(stdout, "Enter number to keep (or 's' to skip, 'a' to keep all): ")
		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "s" || input == "a" || (err != nil && input == "") {
			return -1
		}
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= n {
			return idx - 1
		}
		fmt.Fprintln(stdout, "Invalid input.")
	}
}

func deleteOthers(files dedup.Group, keepIndex int, stdout io.Writer) int64 {
	keeper := files[keepIndex]
	if err := dedup.Verify(keeper); err != nil {
		fmt.Fprintf(stdout, "Skipping set, keeper changed: %v\n", err)
		return 0
	}
	var saved int64
	for i, f := range files {
		if i == keepIndex {
			continue
		}
		fmt.Fprintf(stdout, "Deleting %s... ", f.Path)
		if err := dedup.Verify(f); err != nil {
			fmt.Fprintf(stdout, "Skipped: %v\n", err)
			continue
		}
		if err := os.Remove(f.Path); err != nil {
			fmt.Fprintf(stdout, "Error: %v\n", err)
			continue
		}
		fmt.Fprintln(stdout, "Done.")
		saved += dedup.Reclaims(f)
	}
	return saved
}
