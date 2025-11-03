package musicutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The functions createMusicTestFiles and TestGetAllMusicFiles are already in musicutils_test.go
// If this file is in the same package, you don't need to redefine them.
// However, for isolated testing, let's assume this is a separate test build.

func createTestFileWithDuration(t *testing.T, dir, name string, duration time.Duration) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("Failed to create parent directory for %s: %v", filePath, err)
	}

	// This is a mock. In a real scenario, you would need a file with actual audio data.
	// For the purpose of this test, we create a dummy file and will mock the duration checking.
	// The `audioduration` library actually reads file content, so a simple mock is not enough.
	// We will need to use real audio files with known durations.
	// Let's copy a test file from testdata.
	testDataSource := "../testdata/test.mp3" // Assuming this file has a known duration
	content, err := os.ReadFile(testDataSource)
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
}

func TestGetFilteredMusicFiles_Duration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "musicutils_duration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// This test requires a file with a known duration.
	// The test.mp3 from testdata has a duration of approximately 5 seconds.
	// Let's create a few copies of it.
	testFiles := []string{
		"short_song.mp3", // ~5 seconds
		"another_short_song.mp3",
	}

	for _, name := range testFiles {
		createTestFileWithDuration(t, tmpDir, name, 5*time.Second)
	}

	testCases := []struct {
		name          string
		minDuration   int // in minutes
		expectedCount int
	}{
		{
			name:          "Duration filter matches (0 minutes)",
			minDuration:   0,
			expectedCount: 2,
		},
		{
			name:          "Duration filter excludes (1 minute)",
			minDuration:   1,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualFiles := GetFilteredMusicFiles(tmpDir, "", 0, tc.minDuration)
			if len(actualFiles) != tc.expectedCount {
				t.Errorf("Expected %d files, but got %d for minDuration %d", tc.expectedCount, len(actualFiles), tc.minDuration)
			}
		})
	}
}
