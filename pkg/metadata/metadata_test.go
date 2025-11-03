package metadata

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper function to create a dummy file.
func createDummyFile(t *testing.T, dir string, fileName string, content []byte) string {
	t.Helper()
	filePath := filepath.Join(dir, fileName)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file %s: %v", filePath, err)
	}
	return filePath
}

func TestReadTrackInfo(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "metadata_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// --- Test Cases ---

	// 1. File does not exist
	t.Run("FileDoesNotExist", func(t *testing.T) {
		_, err := ReadTrackInfo(filepath.Join(tmpDir, "nonexistent.mp3"))
		if err == nil {
			t.Errorf("Expected error for non-existent file, got nil")
		} else {
			if !strings.Contains(err.Error(), "file does not exist") {
				t.Errorf("Expected error to contain 'file does not exist', got: %v", err)
			}
		}
	})

	// 2. Unsupported file extension
	t.Run("UnsupportedExtension", func(t *testing.T) {
		unsupportedFilePath := createDummyFile(t, tmpDir, "song.unsupported", []byte("dummy"))
		_, err := ReadTrackInfo(unsupportedFilePath)
		if err == nil {
			t.Errorf("Expected error for unsupported file type, got nil")
		} else if !strings.Contains(err.Error(), "unsupported file type: .unsupported") {
			t.Errorf("Expected error message for unsupported type, got '%v'", err)
		}
	})

	// 3. TXT file (simulates tag reading error, checks defaults)
	t.Run("TXTFileDefaults", func(t *testing.T) {
		txtFilePath := createDummyFile(t, tmpDir, "My Song Title.txt", []byte("this is not an audio file"))
		info, err := ReadTrackInfo(txtFilePath)
		if err != nil {
			t.Fatalf("ReadTrackInfo failed for .txt file: %v", err) // Should not fail, only log warning
		}

		expectedInfo := &TrackInfo{
			SourcePath:        txtFilePath,
			OriginalExtension: ".txt",
			Artist:            "Unknown",
			Album:             "Unknown",
			Title:             "My Song Title", // Derived from filename
			TrackNumber:       1,
			Genre:             "Unknown",
			Year:              0,
		}

		if !reflect.DeepEqual(info, expectedInfo) {
			t.Errorf("TrackInfo mismatch for .txt file.\nGot:  %+v\nWant: %+v", info, expectedInfo)
		}
	})

	// 4. Empty MP3 file (checks defaults and graceful handling of tag reading for empty/malformed files)
	t.Run("EmptyMP3Defaults", func(t *testing.T) {
		mp3FilePath := createDummyFile(t, tmpDir, "Empty Audio.mp3", []byte{}) // Empty file
		info, err := ReadTrackInfo(mp3FilePath)
		if err != nil {
			t.Fatalf("ReadTrackInfo failed for empty .mp3 file: %v", err)
		}

		expectedInfo := &TrackInfo{
			SourcePath:        mp3FilePath,
			OriginalExtension: ".mp3",
			Artist:            "Unknown",
			Album:             "Unknown",
			Title:             "Empty Audio", // Derived from filename
			TrackNumber:       1,
			Genre:             "Unknown",
			Year:              0,
		}

		if !reflect.DeepEqual(info, expectedInfo) {
			t.Errorf("TrackInfo mismatch for empty .mp3 file.\nGot:  %+v\nWant: %+v", info, expectedInfo)
		}
	})

	// 5. Empty FLAC file
	t.Run("EmptyFLACDefaults", func(t *testing.T) {
		flacFilePath := createDummyFile(t, tmpDir, "Silent Sound.flac", []byte{}) // Empty file
		info, err := ReadTrackInfo(flacFilePath)
		if err != nil {
			t.Fatalf("ReadTrackInfo failed for empty .flac file: %v", err)
		}
		expectedInfo := &TrackInfo{
			SourcePath:        flacFilePath,
			OriginalExtension: ".flac",
			Artist:            "Unknown",
			Album:             "Unknown",
			Title:             "Silent Sound",
			TrackNumber:       1,
			Genre:             "Unknown",
			Year:              0,
		}
		if !reflect.DeepEqual(info, expectedInfo) {
			t.Errorf("TrackInfo mismatch for empty .flac file.\nGot:  %+v\nWant: %+v", info, expectedInfo)
		}
	})

	// 6. Empty M4A file
	t.Run("EmptyM4ADefaults", func(t *testing.T) {
		m4aFilePath := createDummyFile(t, tmpDir, "Muted Melody.m4a", []byte{}) // Empty file
		info, err := ReadTrackInfo(m4aFilePath)
		if err != nil {
			t.Fatalf("ReadTrackInfo failed for empty .m4a file: %v", err)
		}
		expectedInfo := &TrackInfo{
			SourcePath:        m4aFilePath,
			OriginalExtension: ".m4a",
			Artist:            "Unknown",
			Album:             "Unknown",
			Title:             "Muted Melody",
			TrackNumber:       1,
			Genre:             "Unknown",
			Year:              0,
		}
		if !reflect.DeepEqual(info, expectedInfo) {
			t.Errorf("TrackInfo mismatch for empty .m4a file.\nGot:  %+v\nWant: %+v", info, expectedInfo)
		}
	})

	// 7. Empty WAV file
	t.Run("EmptyWAVDefaults", func(t *testing.T) {
		wavFilePath := createDummyFile(t, tmpDir, "Quiet Wave.wav", []byte{}) // Empty file
		info, err := ReadTrackInfo(wavFilePath)
		if err != nil {
			t.Fatalf("ReadTrackInfo failed for empty .wav file: %v", err)
		}
		expectedInfo := &TrackInfo{
			SourcePath:        wavFilePath,
			OriginalExtension: ".wav",
			Artist:            "Unknown",
			Album:             "Unknown",
			Title:             "Quiet Wave",
			TrackNumber:       1,
			Genre:             "Unknown",
			Year:              0,
		}
		if !reflect.DeepEqual(info, expectedInfo) {
			t.Errorf("TrackInfo mismatch for empty .wav file.\nGot:  %+v\nWant: %+v", info, expectedInfo)
		}
	})

	// Note: Testing with actual valid tagged files is more involved as it requires
	// either pre-existing tagged files or a library to write tags.
	// The current tests focus on error handling and default value population,
	// especially when tag reading fails or provides no data, which `tag.ReadFrom`
	// might do for empty/malformed files.
	// A more complete test suite would mock `tag.ReadFrom` or use sample audio files.

	// Example of how one might test with a mock if we could inject `tag.ReadFrom`
	// or if `tag.ReadFrom` could be easily controlled for a specific dummy file.
	// For now, we rely on the behavior of `tag.ReadFrom` with empty/txt files to test the default paths.
}

// Example of a more advanced test if we could easily create a file that `tag.ReadFrom` processes
// and returns specific metadata. For now, this is illustrative.
/*
func TestReadTrackInfo_WithTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "metadata_test_with_tags_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// This would require a helper function or a pre-made file
	// that is known to have specific tags readable by `tag.ReadFrom`.
	// For example, if `createTaggedDummyFile` could create such a file:
	// taggedFilePath := createTaggedDummyFile(t, tmpDir, "Tagged Song.mp3",
	// 	map[string]string{"artist": "Test Artist", "album": "Test Album", "title": "Test Title", "genre": "Test Genre"}, 2, 2023)

	// _, err = ReadTrackInfo(taggedFilePath)
	// ... assertions ...
}
*/

func TestMain(m *testing.M) {
	// Setup code, if any, can go here.
	// For example, ensuring any external dependencies for testing are available.
	// For `tag.ReadFrom`, it relies on the file content.

	// Run the tests
	exitCode := m.Run()

	// Teardown code, if any, can go here.
	os.Exit(exitCode)
}
