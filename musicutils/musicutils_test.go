package musicutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists_ExistingFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_existing_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	if !FileExists(tmpFile.Name()) {
		t.Errorf("FileExists(%q) = false, want true", tmpFile.Name())
	}
	tmpFile.Close() // Close the file
}

func TestFileExists_NonExistingFile(t *testing.T) {
	nonExistentFilePath := filepath.Join(os.TempDir(), "non_existent_file_12345.txt")
	if FileExists(nonExistentFilePath) {
		t.Errorf("FileExists(%q) = true, want false", nonExistentFilePath)
	}
}

func TestFileExists_ExistingDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_existing_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up

	if FileExists(tmpDir) {
		t.Errorf("FileExists(%q) = true, want false for directory", tmpDir)
	}
}

func TestFolderExists_ExistingFolder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_existing_folder_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if !FolderExists(tmpDir) {
		t.Errorf("FolderExists(%q) = false, want true", tmpDir)
	}
}

func TestFolderExists_NonExistingFolder(t *testing.T) {
	nonExistentFolderPath := filepath.Join(os.TempDir(), "non_existent_folder_12345")
	if FolderExists(nonExistentFolderPath) {
		t.Errorf("FolderExists(%q) = true, want false", nonExistentFolderPath)
	}
}

func TestFolderExists_PathIsFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_folder_is_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if FolderExists(tmpFile.Name()) {
		t.Errorf("FolderExists(%q) = true for a file, want false", tmpFile.Name())
	}
}

func TestIsDirEmpty_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_empty_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	empty, err := IsDirEmpty(tmpDir)
	if err != nil {
		t.Fatalf("IsDirEmpty(%q) returned error: %v", tmpDir, err)
	}
	if !empty {
		t.Errorf("IsDirEmpty(%q) = false, want true", tmpDir)
	}
}

func TestIsDirEmpty_NonEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_non_empty_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = os.CreateTemp(tmpDir, "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file in dir: %v", err)
	}

	empty, err := IsDirEmpty(tmpDir)
	if err != nil {
		t.Fatalf("IsDirEmpty(%q) returned error: %v", tmpDir, err)
	}
	if empty {
		t.Errorf("IsDirEmpty(%q) = true for non-empty dir, want false", tmpDir)
	}
}

func TestIsDirEmpty_NonExistentDir(t *testing.T) {
	nonExistentPath := filepath.Join(os.TempDir(), "non_existent_dir_for_isempty_test")
	_, err := IsDirEmpty(nonExistentPath)
	if err == nil {
		t.Errorf("IsDirEmpty(%q) did not return error for non-existent dir, want error", nonExistentPath)
	}
}

func TestDeleteFile_DeletesFileAndEmptyParentDirs(t *testing.T) {
	// Create nested temp directories: rootTmpDir/parentDir/childDir/testfile.txt
	rootTmpDir, err := os.MkdirTemp("", "test_delete_root_*")
	if err != nil {
		t.Fatalf("Failed to create root temp dir: %v", err)
	}
	defer os.RemoveAll(rootTmpDir) // Ensure root is cleaned even if test fails midway

	parentDir := filepath.Join(rootTmpDir, "parentDir")
	err = os.Mkdir(parentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create parentDir: %v", err)
	}

	childDir := filepath.Join(parentDir, "childDir")
	err = os.Mkdir(childDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create childDir: %v", err)
	}

	tmpFile, err := os.Create(filepath.Join(childDir, "testfile.txt"))
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	filePath := tmpFile.Name()
	tmpFile.Close()

	// Call DeleteFile
	DeleteFile(filePath)

	// Assert file is deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File %q was not deleted", filePath)
	}

	// Assert childDir is deleted (because it became empty)
	if _, err := os.Stat(childDir); !os.IsNotExist(err) {
		t.Errorf("Empty child directory %q was not deleted", childDir)
	}

	// Assert parentDir is deleted (because it became empty)
	if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
		t.Errorf("Empty parent directory %q was not deleted", parentDir)
	}

	// Assert rootTmpDir still exists (as it's the root of the deletion logic for this test)
	if _, err := os.Stat(rootTmpDir); os.IsNotExist(err) {
		t.Errorf("Root temp directory %q was unexpectedly deleted", rootTmpDir)
	}
}

func TestDeleteFile_DoesNotDeleteNonEmptyParentDir(t *testing.T) {
	rootTmpDir, err := os.MkdirTemp("", "test_delete_nonempty_root_*")
	if err != nil {
		t.Fatalf("Failed to create root temp dir: %v", err)
	}
	defer os.RemoveAll(rootTmpDir)

	parentDir := filepath.Join(rootTmpDir, "parentDir")
	err = os.Mkdir(parentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create parentDir: %v", err)
	}

	// Create the file to be deleted
	fileToDelete, err := os.Create(filepath.Join(parentDir, "fileToDelete.txt"))
	if err != nil {
		t.Fatalf("Failed to create fileToDelete: %v", err)
	}
	filePathToDelete := fileToDelete.Name()
	fileToDelete.Close()

	// Create another file in parentDir so it's not empty after fileToDelete is removed
	siblingFile, err := os.Create(filepath.Join(parentDir, "siblingFile.txt"))
	if err != nil {
		t.Fatalf("Failed to create siblingFile: %v", err)
	}
	siblingFile.Close()

	DeleteFile(filePathToDelete)

	if _, err := os.Stat(filePathToDelete); !os.IsNotExist(err) {
		t.Errorf("File %q was not deleted", filePathToDelete)
	}
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Non-empty parent directory %q was deleted", parentDir)
	}
	if _, err := os.Stat(siblingFile.Name()); os.IsNotExist(err) {
		t.Errorf("Sibling file %q was unexpectedly deleted", siblingFile.Name())
	}
}
