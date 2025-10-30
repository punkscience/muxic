package cmd

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestEnvironment creates a temporary directory structure for testing.
// It returns the source and target directory paths and a cleanup function.
func setupTestEnvironment(t *testing.T) (string, string, func()) {
	t.Helper()

	// Create a temporary directory for the test
	tmpDir, err := ioutil.TempDir("", "muxic-test-*")
	assert.NoError(t, err)

	// Create source and target directories
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	assert.NoError(t, os.Mkdir(sourceDir, 0755))
	assert.NoError(t, os.Mkdir(targetDir, 0755))

	// Create dummy music files and copy the tagged file
	createDummyFile(t, sourceDir, "untagged.mp3", 1)      // 1MB
	createDummyFile(t, sourceDir, "another.flac", 2)       // 2MB
	createDummyFile(t, filepath.Join(sourceDir, "subdir"), "sub.m4a", 3) // 3MB
	copyTaggedFile(t, sourceDir, "../testdata/test.mp3")

	// Return a cleanup function to remove the temporary directory
	cleanup := func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}

	return sourceDir, targetDir, cleanup
}

// createDummyFile creates a dummy file of a given size in MB.
func createDummyFile(t *testing.T, dir, name string, sizeMB int) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	assert.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
	content := make([]byte, sizeMB*1024*1024)
	assert.NoError(t, ioutil.WriteFile(filePath, content, 0644))
}

// copyTaggedFile copies a tagged file from the testdata directory to the source directory.
func copyTaggedFile(t *testing.T, sourceDir, sourceFile string) {
	t.Helper()
	content, err := ioutil.ReadFile(sourceFile)
	assert.NoError(t, err)
	destFile := filepath.Join(sourceDir, filepath.Base(sourceFile))
	assert.NoError(t, ioutil.WriteFile(destFile, content, 0644))
}

// setupCobra defines the flags for the copy command.
func setupCobra() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().String("source", "", "The source folder containing music files.")
	copyCmd.Flags().String("target", "", "The destination folder where music files will be organized.")
	copyCmd.Flags().String("filter", "", "Filter files by a string contained in their path (case-insensitive).")
	copyCmd.Flags().Int("over", 0, "Only process files over this size in megabytes (MB).")
	copyCmd.Flags().BoolVarP(&destructive, "move", "m", false, "Move files instead of copying (deletes source files and empty parent dirs).")
	copyCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging for detailed operation output.")
	copyCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate operations without making any changes to the file system.")
}

// executeCommand runs the copy command with the given arguments and returns the log output.
func executeCommand(t *testing.T, args ...string) string {
	t.Helper()

	// Redirect log output to a buffer
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)

	// Reset and re-initialize flags before each execution
	rootCmd.ResetFlags()
	copyCmd.ResetFlags()
	setupCobra()

	// Set up the command with arguments
	rootCmd.SetArgs(append([]string{"copy"}, args...))
	Execute()

	// Restore original log output
	log.SetOutput(os.Stderr)

	return logOutput.String()
}

func TestCopyCommand_DryRun(t *testing.T) {
	sourceDir, targetDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	args := []string{
		"--source", sourceDir,
		"--target", targetDir,
		"--dry-run",
		"--verbose",
	}

	output := executeCommand(t, args...)

	assert.Contains(t, output, "[DRY-RUN] Operation: Copying")
	assert.Contains(t, output, "[DRY-RUN] Would copy")
	assert.NoFileExists(t, filepath.Join(targetDir, "test.mp3"))
}

func TestMoveCommand_DryRun(t *testing.T) {
	sourceDir, targetDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	args := []string{
		"--source", sourceDir,
		"--target", targetDir,
		"--move",
		"--dry-run",
		"--verbose",
	}

	output := executeCommand(t, args...)

	assert.Contains(t, output, "[DRY-RUN] Operation: Moving")
	assert.Contains(t, output, "Simulated delete actions")
	assert.FileExists(t, filepath.Join(sourceDir, "test.mp3"))
}

func TestCopyCommand_Filter(t *testing.T) {
	sourceDir, targetDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	args := []string{
		"--source", sourceDir,
		"--target", targetDir,
		"--filter", "another",
	}

	executeCommand(t, args...)

	assert.NoFileExists(t, filepath.Join(targetDir, "test.mp3"))
	// This assertion is tricky because the destination path is based on metadata, which is missing.
	// We'll check if the log contains the processing of the filtered file.
	output := executeCommand(t, "--source", sourceDir, "--target", targetDir, "--filter", "another", "--dry-run", "--verbose")
	assert.Contains(t, output, "another.flac")
	assert.NotContains(t, output, "test.mp3")
}

func TestCopyCommand_Over(t *testing.T) {
	sourceDir, targetDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	args := []string{
		"--source", sourceDir,
		"--target", targetDir,
		"--over", "2",
	}

	executeCommand(t, args...)

	// Again, checking logs for verification due to metadata dependency.
	output := executeCommand(t, "--source", sourceDir, "--target", targetDir, "--over", "2", "--dry-run", "--verbose")
	assert.Contains(t, output, "sub.m4a")
	assert.Contains(t, output, "another.flac")
	assert.NotContains(t, output, "untagged.mp3")
}

func TestCopyCommand_Metadata(t *testing.T) {
	sourceDir, targetDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	args := []string{
		"--source", sourceDir,
		"--target", targetDir,
	}

	executeCommand(t, args...)

	expectedPath := filepath.Join(targetDir, "Test Artist", "Test Album", "01 - Test Title.mp3")
	assert.FileExists(t, expectedPath)
}
