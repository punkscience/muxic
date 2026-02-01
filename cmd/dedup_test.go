package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestFile(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return path
}

func TestDedupScorchedEarth(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "dedup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create music files
	// file1 and file2 are duplicates
	createTestFile(t, tempDir, "song1.mp3", "duplicate_content")
	createTestFile(t, tempDir, "song1_copy.mp3", "duplicate_content")
	// file3 is unique
	createTestFile(t, tempDir, "song2.mp3", "unique_content")

	// Verify they exist
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 3, len(files))

	// Run dedup with scorched earth
	var stdin bytes.Buffer // No input needed for scorched earth
	var stdout bytes.Buffer

	err = runDedup(tempDir, true, &stdin, &stdout)
	assert.NoError(t, err)

	// Verify output contains expected messages
	output := stdout.String()
	assert.Contains(t, output, "Scan complete")
	assert.Contains(t, output, "Scorched Earth: keeping")
	assert.Contains(t, output, "Deleting")
	assert.Contains(t, output, "Cleanup complete")

	// Verify files remaining
	// Should satisfy: song2.mp3 exists. One of song1.mp3 or song1_copy.mp3 exists.
	// song1.mp3 is shorter than song1_copy.mp3 (length 9 vs 14), so song1.mp3 shout kept.

	remainingFiles, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(remainingFiles))

	_, err = os.Stat(filepath.Join(tempDir, "song2.mp3"))
	assert.NoError(t, err, "Unique file should persist")

	_, err1 := os.Stat(filepath.Join(tempDir, "song1.mp3"))
	_, err2 := os.Stat(filepath.Join(tempDir, "song1_copy.mp3"))

	// One should exist, one should not
	assert.True(t, (err1 == nil && os.IsNotExist(err2)) || (os.IsNotExist(err1) && err2 == nil))

	// Specifically, with our sorting logic (shortest path first), song1.mp3 should be kept
	assert.NoError(t, err1, "song1.mp3 should be kept (shorter name)")
	assert.True(t, os.IsNotExist(err2), "song1_copy.mp3 should be deleted")
}

func TestDedupNoDuplicates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dedup_test_nodup")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createTestFile(t, tempDir, "song1.mp3", "content1")
	createTestFile(t, tempDir, "song2.mp3", "content2")

	var stdin bytes.Buffer
	var stdout bytes.Buffer

	err = runDedup(tempDir, true, &stdin, &stdout)
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "No duplicates found")

	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}
