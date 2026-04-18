package cmd

import (
	"log"
	"muxic/musicutils"
	"muxic/pkg/tagger"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag <word>",
	Short: "Tag audio files with a word stored in the MUXIC_TAGS custom metadata field.",
	Long: `Writes a word (e.g. "dnb", "workout") to the MUXIC_TAGS custom metadata field
of every audio file under --source. The operation is idempotent: if the word is
already present on a file it is silently skipped.

The tag is stored in a format-native custom field so that standard tags such as
Genre are left untouched.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		word := strings.TrimSpace(args[0])
		source := strings.TrimSpace(cmd.Flag("source").Value.String())
		dryRun := cmd.Flag("dry-run").Value.String() == "true"
		verbose := cmd.Flag("verbose").Value.String() == "true"

		if source == "" {
			log.Fatal("Muxic: --source flag is required.")
		}
		if word == "" {
			log.Fatal("Muxic: tag word must not be empty.")
		}

		if dryRun {
			log.Printf("Muxic: Dry-run mode. No files will be modified.")
		}
		log.Printf("Muxic: Tagging files under %q with %q.", source, word)

		filesChan := musicutils.GetAllMusicFiles(source)

		var changedCount int64
		var errorCount int64
		var wg sync.WaitGroup

		numWorkers := runtime.NumCPU()
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for file := range filesChan {
					changed, err := tagger.AddTag(file, word, dryRun)
					if err != nil {
						log.Printf("Error tagging %s: %v", file, err)
						atomic.AddInt64(&errorCount, 1)
						continue
					}
					if changed {
						if verbose {
							if dryRun {
								log.Printf("[DRY-RUN] Would tag: %s", file)
							} else {
								log.Printf("Tagged: %s", file)
							}
						}
						atomic.AddInt64(&changedCount, 1)
					}
				}
			}()
		}

		wg.Wait()
		log.Printf("Muxic: Done. %d files tagged, %d errors.", changedCount, errorCount)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().String("source", "", "File or directory containing audio files to tag.")
	tagCmd.Flags().BoolP("dry-run", "n", false, "Simulate tagging without modifying any files.")
	tagCmd.Flags().BoolP("verbose", "v", false, "Print each file as it is tagged.")
}
