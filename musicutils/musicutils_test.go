package musicutils

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// Helper function to create dummy files for testing Get(All|Filtered)MusicFiles
func createMusicTestFiles(t *testing.T, rootDir string, fileNames []string) {
	t.Helper()
	for _, name := range fileNames {
		filePath := filepath.Join(rootDir, name)
		// Create parent dirs if they don't exist (for nested test cases)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create parent directory for %s: %v", filePath, err)
		}
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create dummy file %s: %v", filePath, err)
		}
		// Write some minimal data if needed by size filters, otherwise empty is fine
		if filepath.Ext(name) == ".mp3" { // example for size
			f.Write(make([]byte, 1024*1024)) // 1MB
		}
		f.Close()
	}
}

func TestGetAllMusicFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "musicutils_getall_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories for more complex scanning
	subDir1 := filepath.Join(tmpDir, "subdir1")
	os.Mkdir(subDir1, 0755)
	subDir2 := filepath.Join(tmpDir, "subdir1", "subdir2")
	os.Mkdir(subDir2, 0755)

	testFiles := []string{
		"song1.mp3",
		"song2.flac",
		"song3.m4a",
		"song4.wav",
		"notsong.txt",
		filepath.Join("subdir1", "song5.mp3"),
		filepath.Join("subdir1", "subdir2", "song6.flac"),
		filepath.Join("subdir1", "notsong.doc"),
	}
	createMusicTestFiles(t, tmpDir, testFiles)

	expectedFiles := []string{
		filepath.Join(tmpDir, "song1.mp3"),
		filepath.Join(tmpDir, "song2.flac"),
		filepath.Join(tmpDir, "song3.m4a"),
		filepath.Join(tmpDir, "song4.wav"),
		filepath.Join(tmpDir, "subdir1", "song5.mp3"),
		filepath.Join(tmpDir, "subdir1", "subdir2", "song6.flac"),
	}

	actualFiles := GetAllMusicFiles(tmpDir)

	// Sort both slices for consistent comparison
	sort.Strings(actualFiles)
	sort.Strings(expectedFiles)

	if !reflect.DeepEqual(actualFiles, expectedFiles) {
		t.Errorf("GetAllMusicFiles() mismatch.\nGot:    %v\nWanted: %v", actualFiles, expectedFiles)
	}

	// Test with a non-existent folder
	emptyResult := GetAllMusicFiles(filepath.Join(tmpDir, "non_existent_folder_123"))
	if len(emptyResult) != 0 {
		t.Errorf("GetAllMusicFiles() on non-existent folder should return empty slice, got %v", emptyResult)
	}
}

func TestGetFilteredMusicFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "musicutils_getfiltered_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "rock_band")
	os.Mkdir(subDir, 0755)

	testFiles := []string{
		"Artist - SongA.mp3",                        // Matches "artist", size 1MB
		"Artist - SongB.flac",                       // Matches "artist"
		"Another Artist - SongC.m4a",                // Does not match "artist"
		"ARTIST - SongD.wav",                        // Matches "artist" (case-insensitive path)
		filepath.Join(subDir, "Artist - SongE.mp3"), // Matches "artist", size 1MB
		"BigFileSong.mp3",                           // Size 1MB
		"SmallFile.mp3",                             // Size 0MB (or very small)
	}
	// Create files, ensuring "BigFileSong.mp3" and "Artist - SongA.mp3" are >0MB for size filter tests
	// The helper `createMusicTestFiles` makes .mp3 files 1MB.
	createMusicTestFiles(t, tmpDir, testFiles)

	// Manually create SmallFile.mp3 with 0 size
	smallFilePath := filepath.Join(tmpDir, "SmallFile.mp3")
	sf, _ := os.Create(smallFilePath)
	sf.Close()

	testCases := []struct {
		name          string
		folder        string
		filter        string
		maxMB         int
		minDuration   int
		expectedFiles []string
	}{
		{
			name:   "Filter by artist name",
			folder: tmpDir,
			filter: "artist",
			maxMB:  0, // No size limit
			minDuration: 0,
			expectedFiles: []string{
				filepath.Join(tmpDir, "Artist - SongA.mp3"),
				filepath.Join(tmpDir, "Artist - SongB.flac"),
				filepath.Join(tmpDir, "ARTIST - SongD.wav"),
				filepath.Join(tmpDir, subDir, "Artist - SongE.mp3"),
				filepath.Join(tmpDir, "Another Artist - SongC.m4a"), // This was missing
			},
		},
		{
			name:   "Filter by file extension",
			folder: tmpDir,
			filter: ".flac",
			maxMB:  0,
			minDuration: 0,
			expectedFiles: []string{
				filepath.Join(tmpDir, "Artist - SongB.flac"),
			},
		},
		{
			name:   "Filter by subfolder name",
			folder: tmpDir,
			filter: "rock_band",
			maxMB:  0,
			minDuration: 0,
			expectedFiles: []string{
				filepath.Join(tmpDir, subDir, "Artist - SongE.mp3"),
			},
		},
		{
			name:   "Filter with size limit (no specific name filter)",
			folder: tmpDir,
			filter: "", // No name filter
			maxMB:  1,  // Files > 1MB (actually >= 1MB due to helper creating 1MB files)
			minDuration: 0,
			expectedFiles: []string{
				filepath.Join(tmpDir, "Artist - SongA.mp3"),
				filepath.Join(tmpDir, subDir, "Artist - SongE.mp3"),
				filepath.Join(tmpDir, "BigFileSong.mp3"),
			},
		},
		{
			name:   "Filter by name and size limit",
			folder: tmpDir,
			filter: "artist",
			maxMB:  1, // Only "Artist" files that are >= 1MB
			minDuration: 0,
			expectedFiles: []string{
				filepath.Join(tmpDir, "Artist - SongA.mp3"),
				filepath.Join(tmpDir, subDir, "Artist - SongE.mp3"),
			},
		},
		{
			name:          "No matching filter",
			folder:        tmpDir,
			filter:        "nonexistent_filter_term",
			maxMB:         0,
			minDuration: 0,
			expectedFiles: []string{},
		},
		{
			name:          "Non-existent folder",
			folder:        filepath.Join(tmpDir, "non_existent_folder_XYZ"),
			filter:        "",
			maxMB:         0,
			minDuration: 0,
			expectedFiles: []string{},
		},
		{
			name:          "Size filter excludes all",
			folder:        tmpDir,
			filter:        "",
			maxMB:         2, // All test .mp3 files are 1MB
			minDuration: 0,
			expectedFiles: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualFiles := GetFilteredMusicFiles(tc.folder, tc.filter, tc.maxMB, tc.minDuration)
			sort.Strings(actualFiles)
			sort.Strings(tc.expectedFiles)
			if !reflect.DeepEqual(actualFiles, tc.expectedFiles) {
				t.Errorf("GetFilteredMusicFiles() with filter '%s', maxMB %d mismatch.\nGot:    %v\nWanted: %v", tc.filter, tc.maxMB, actualFiles, tc.expectedFiles)
			}
		})
	}
}
