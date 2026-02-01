package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"muxic/pkg/dedup"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	targetDir     string
	scorchedEarth bool
)

// dedupCmd represents the dedup command
var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Find and remove duplicate music files",
	Long: `Scans the target directory for duplicate music files based on exact binary content.
Maintains a local cache to speed up subsequent scans.
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
	dedupCmd.Flags().BoolVar(&scorchedEarth, "scorchedearth", false, "Automatically delete duplicates, keeping the one with shortest path")
}

func runDedup(targetDir string, scorchedEarth bool, stdin io.Reader, stdout io.Writer) error {
	// Resolve user home directory for cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %v", err)
	}
	cachePath := filepath.Join(homeDir, ".muxic", "dedup_cache.json")

	fmt.Fprintln(stdout, "Loading cache from", cachePath)
	cache, err := dedup.LoadCache(cachePath)
	if err != nil {
		fmt.Fprintf(stdout, "Warning: Could not load cache: %v. Starting fresh.\n", err)
		cache = make(dedup.Cache)
	}

	fmt.Fprintf(stdout, "Scanning %s...\n", targetDir)

	filesBySig := make(map[string][]string)

	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Simple extension check
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp3" && ext != ".flac" && ext != ".m4a" && ext != ".wav" {
			return nil
		}

		sig, updated, err := dedup.UpdateEntry(path, info, cache, nil)
		if err != nil {
			fmt.Fprintf(stdout, "Error processing %s: %v\n", path, err)
			return nil
		}
		if updated {
			// optional: fmt.Fprintf(stdout, ".")
		}

		filesBySig[sig] = append(filesBySig[sig], path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking target directory: %v", err)
	}
	fmt.Fprintln(stdout, "\nScan complete.")

	// Process duplicates
	reader := bufio.NewReader(stdin)
	duplicatesFound := 0
	bytesSaved := int64(0)

	// Create a list of signatures to iterate deterministically
	var sigs []string
	for sig, files := range filesBySig {
		if len(files) > 1 {
			sigs = append(sigs, sig)
		}
	}
	sort.Strings(sigs)

	for _, sig := range sigs {
		files := filesBySig[sig]
		duplicatesFound++

		fmt.Fprintf(stdout, "\nDuplicate set found (Signature: %s...):\n", sig[:8])

		// Sort files to ensure deterministic order (e.g. by path length then name)
		sort.Slice(files, func(i, j int) bool {
			if len(files[i]) != len(files[j]) {
				return len(files[i]) < len(files[j]) // Prefer shorter paths
			}
			return files[i] < files[j]
		})

		for i, f := range files {
			fmt.Fprintf(stdout, "%d) %s\n", i+1, f)
		}

		var keepIndex int = -1

		if scorchedEarth {
			keepIndex = 0 // Keep the first one (shortest path)
			fmt.Fprintf(stdout, "Scorched Earth: keeping %s\n", files[0])
		} else {
			for {
				fmt.Fprint(stdout, "Enter number to keep (or 's' to skip, 'a' to keep all): ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input == "s" || input == "a" {
					keepIndex = -1
					break
				}

				var idx int
				if _, err := fmt.Sscanf(input, "%d", &idx); err == nil {
					if idx >= 1 && idx <= len(files) {
						keepIndex = idx - 1
						break
					}
				}
				fmt.Fprintln(stdout, "Invalid input.")
			}
		}

		if keepIndex != -1 {
			// Delete others
			for i, f := range files {
				if i == keepIndex {
					continue
				}

				fmt.Fprintf(stdout, "Deleting %s... ", f)
				if err := os.Remove(f); err != nil {
					fmt.Fprintf(stdout, "Error: %v\n", err)
				} else {
					fmt.Fprintln(stdout, "Done.")
					// Remove from cache
					delete(cache, f)

					if entry, ok := cache[files[i]]; ok {
						bytesSaved += entry.Size
					}
				}
			}
		}
	}

	fmt.Fprintln(stdout, "Pruning cache...")
	for path := range cache {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			delete(cache, path)
		}
	}

	if err := dedup.SaveCache(cachePath, cache); err != nil {
		fmt.Fprintf(stdout, "Error saving cache: %v\n", err)
	} else {
		fmt.Fprintln(stdout, "Cache saved.")
	}

	if duplicatesFound == 0 {
		fmt.Fprintln(stdout, "No duplicates found.")
	} else {
		fmt.Fprintf(stdout, "Cleanup complete. Saved approx %.2f MB\n", float64(bytesSaved)/(1024*1024))
	}

	return nil
}
