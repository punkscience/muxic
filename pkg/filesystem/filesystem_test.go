package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Helper to create a structure of directories and a file for testing
func setupTestDirStructure(t *testing.T, root string, structure []string, testFileName string) (string, string, string, string) {
	t.Helper()
	currentPath := root
	for _, dir := range structure {
		currentPath = filepath.Join(currentPath, dir)
		if err := os.Mkdir(currentPath, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", currentPath, err)
		}
	}
	testFilePath := ""
	if testFileName != "" {
		testFilePath = filepath.Join(currentPath, testFileName)
		f, err := os.Create(testFilePath)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", testFilePath, err)
		}
		f.Close()
	}
	// Return root, deepest sub-directory, path to file, and potentially a mid-level directory for rootDir tests
	parentDir := ""
	if len(structure) > 0 {
		parentDir = filepath.Join(root, structure[0])
	}
	return root, currentPath, testFilePath, parentDir
}

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test_fileexists_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "testfile.txt")
	f, _ := os.Create(filePath)
	f.Close()

	dirPath := filepath.Join(tmpDir, "testdir")
	os.Mkdir(dirPath, 0755)

	if !FileExists(filePath) {
		t.Errorf("FileExists returned false for existing file %s", filePath)
	}
	if FileExists(dirPath) {
		t.Errorf("FileExists returned true for directory %s", dirPath)
	}
	if FileExists(filepath.Join(tmpDir, "nonexistent.txt")) {
		t.Errorf("FileExists returned true for non-existent file")
	}
}

func TestFolderExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test_folderexists_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "testfile.txt")
	f, _ := os.Create(filePath)
	f.Close()

	dirPath := filepath.Join(tmpDir, "testdir")
	os.Mkdir(dirPath, 0755)

	if !FolderExists(dirPath) {
		t.Errorf("FolderExists returned false for existing directory %s", dirPath)
	}
	if FolderExists(filePath) {
		t.Errorf("FolderExists returned true for file %s", filePath)
	}
	if FolderExists(filepath.Join(tmpDir, "nonexistent_dir")) {
		t.Errorf("FolderExists returned true for non-existent directory")
	}
}

func TestIsDirEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test_isdirempty_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	emptyDirPath := filepath.Join(tmpDir, "empty_dir")
	os.Mkdir(emptyDirPath, 0755)

	nonEmptyDirPath := filepath.Join(tmpDir, "non_empty_dir")
	os.Mkdir(nonEmptyDirPath, 0755)
	f, _ := os.Create(filepath.Join(nonEmptyDirPath, "file.txt"))
	f.Close()

	empty, err := IsDirEmpty(emptyDirPath)
	if err != nil {
		t.Fatalf("IsDirEmpty for empty dir returned error: %v", err)
	}
	if !empty {
		t.Errorf("IsDirEmpty returned false for empty directory %s", emptyDirPath)
	}

	empty, err = IsDirEmpty(nonEmptyDirPath)
	if err != nil {
		t.Fatalf("IsDirEmpty for non-empty dir returned error: %v", err)
	}
	if empty {
		t.Errorf("IsDirEmpty returned true for non-empty directory %s", nonEmptyDirPath)
	}

	_, err = IsDirEmpty(filepath.Join(tmpDir, "nonexistent_dir"))
	if err == nil {
		t.Errorf("IsDirEmpty did not return error for non-existent directory")
	}
}

func TestDeleteFileAndPruneParents(t *testing.T) {
	baseTmpDir, err := os.MkdirTemp("", "fs_test_delete_*")
	if err != nil {
		t.Fatalf("Failed to create base temp dir: %v", err)
	}
	defer os.RemoveAll(baseTmpDir)

	tests := []struct {
		name                 string
		structure            []string                                      // Directories to create, nested
		fileToCreate         string                                        // File to create in the deepest directory
		rootDirSelector      func(root, parent1, deepestDir string) string // Selects the rootDir for pruning
		dryRun               bool
		expectedActions      func(file, deepestDir, parent1 string) []string
		expectFileDeleted    bool
		expectDirsDeleted    []func(root, parent1, deepestDir string) string // Dirs expected to be deleted
		expectDirsKept       []func(root, parent1, deepestDir string) string // Dirs expected to be kept
		expectErrorSubstring string
	}{
		{
			name:            "DryRun_DeleteFileAndTwoLevelsOfEmptyParents",
			structure:       []string{"parent1", "child1"},
			fileToCreate:    "test.txt",
			rootDirSelector: func(root, p1, dp string) string { return root },
			dryRun:          true,
			expectedActions: func(file, deepestDir, parent1 string) []string {
				return []string{
					fmt.Sprintf("Would delete file: %s", file),
					fmt.Sprintf("Would delete empty directory: %s", deepestDir),
					fmt.Sprintf("Would delete empty directory: %s", parent1),
				}
			},
			expectFileDeleted: false, // Dry run
			expectDirsKept: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return r },
				func(r, p1, dp string) string { return p1 },
				func(r, p1, dp string) string { return dp },
			},
		},
		{
			name:              "ActualRun_DeleteFileAndTwoLevelsOfEmptyParents_StopAtRoot",
			structure:         []string{"parent1", "child1"},
			fileToCreate:      "test.txt",
			rootDirSelector:   func(root, p1, dp string) string { return root },
			dryRun:            false,
			expectFileDeleted: true,
			expectDirsDeleted: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return p1 },
				func(r, p1, dp string) string { return dp },
			},
			expectDirsKept: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return r },
			},
		},
		{
			name:              "ActualRun_DeleteFile_StopAtParent1",
			structure:         []string{"parent1", "child1"},
			fileToCreate:      "test.txt",
			rootDirSelector:   func(root, p1, dp string) string { return p1 }, // rootDir is parent1
			dryRun:            false,
			expectFileDeleted: true,
			expectDirsDeleted: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return dp }, // child1 deleted
			},
			expectDirsKept: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return r },  // root should always be kept
				func(r, p1, dp string) string { return p1 }, // parent1 is rootDir, so kept
			},
		},
		{
			name:              "ActualRun_NonEmptyParent_StopsPruning",
			structure:         []string{"parent1", "child1"},
			fileToCreate:      "test.txt",
			rootDirSelector:   func(root, p1, dp string) string { return root },
			dryRun:            false,
			expectFileDeleted: true,
			// Setup: Create another file in parent1 to make it non-empty after child1 is pruned
			// expectDirsDeleted will only include child1
			// expectDirsKept will include root and parent1
		},
		{
			name:            "DryRun_FileNonExistent",
			structure:       []string{"parent1"},
			fileToCreate:    "", // No file created
			rootDirSelector: func(root, p1, dp string) string { return root },
			dryRun:          true,
			expectedActions: func(file, deepestDir, parent1 string) []string {
				// 'file' here will be the path to the non-existent file
				return []string{fmt.Sprintf("File not found, would not delete: %s", file)}
			},
			expectFileDeleted: false,
			expectDirsKept: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return r },
				func(r, p1, dp string) string { return p1 },
			},
		},
		{
			name:                 "ActualRun_FileNonExistent",
			structure:            []string{"parent1"},
			fileToCreate:         "", // No file created
			rootDirSelector:      func(root, p1, dp string) string { return root },
			dryRun:               false,
			expectErrorSubstring: "does not exist or is a directory",
			expectFileDeleted:    false,
			expectDirsKept: []func(root, p1, dp string) string{
				func(r, p1, dp string) string { return r },
				func(r, p1, dp string) string { return p1 },
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique test root for each test case to avoid interference
			testCaseRoot, err := os.MkdirTemp(baseTmpDir, "testcase_*")
			if err != nil {
				t.Fatalf("Failed to create test case root dir: %v", err)
			}
			// Note: defer os.RemoveAll(testCaseRoot) // This would clean up before checks can be made for actual runs.
			// Cleanup of testCaseRoot will be handled by the cleanup of baseTmpDir.

			actualRoot, deepestDir, filePath, parent1Path := setupTestDirStructure(t, testCaseRoot, tc.structure, tc.fileToCreate)

			if tc.fileToCreate == "" && tc.name == "DryRun_FileNonExistent" { // Special handling for this test case
				filePath = filepath.Join(deepestDir, "non_existent_file.txt")
			}
			if tc.fileToCreate == "" && tc.name == "ActualRun_FileNonExistent" {
				filePath = filepath.Join(deepestDir, "non_existent_file.txt")
			}

			if tc.name == "ActualRun_NonEmptyParent_StopsPruning" {
				// Create a sibling file in parent1 to make it non-empty
				siblingFilePath := filepath.Join(parent1Path, "sibling.txt")
				f, err := os.Create(siblingFilePath)
				if err != nil {
					t.Fatalf("Failed to create sibling file: %v", err)
				}
				f.Close()
			}

			selectedRootDir := tc.rootDirSelector(actualRoot, parent1Path, deepestDir)
			actions, err := DeleteFileAndPruneParents(filePath, selectedRootDir, tc.dryRun)

			if tc.expectErrorSubstring != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.expectErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectErrorSubstring) {
					t.Errorf("Expected error to contain '%s', got '%v'", tc.expectErrorSubstring, err)
				}
			} else if err != nil {
				t.Errorf("Did not expect error, got %v", err)
			}

			if tc.dryRun && tc.expectedActions != nil {
				expected := tc.expectedActions(filePath, deepestDir, parent1Path)
				sort.Strings(actions)
				sort.Strings(expected)
				if !reflect.DeepEqual(actions, expected) {
					t.Errorf("Dry run actions mismatch.\nGot:    %v\nWanted: %v", actions, expected)
				}
			}

			// Verify file deletion status
			if tc.expectFileDeleted {
				if FileExists(filePath) {
					t.Errorf("Expected file %s to be deleted, but it exists.", filePath)
				}
			} else {
				if tc.fileToCreate != "" && !FileExists(filePath) { // Only check if file was meant to exist
					t.Errorf("Expected file %s to exist, but it was deleted.", filePath)
				}
			}

			// Verify directory deletion/retention status
			for _, dirGetter := range tc.expectDirsDeleted {
				dirPath := dirGetter(actualRoot, parent1Path, deepestDir)
				if FolderExists(dirPath) {
					t.Errorf("Expected directory %s to be deleted, but it exists.", dirPath)
				}
			}
			for _, dirGetter := range tc.expectDirsKept {
				dirPath := dirGetter(actualRoot, parent1Path, deepestDir)
				if !FolderExists(dirPath) {
					t.Errorf("Expected directory %s to exist, but it was deleted.", dirPath)
				}
			}

			// Special check for "ActualRun_NonEmptyParent_StopsPruning"
			if tc.name == "ActualRun_NonEmptyParent_StopsPruning" {
				if FolderExists(deepestDir) { // child1 should be deleted
					t.Errorf("Expected directory %s (child1) to be deleted, but it exists.", deepestDir)
				}
				if !FolderExists(parent1Path) { // parent1 should be kept (non-empty)
					t.Errorf("Expected directory %s (parent1) to be kept, but it was deleted.", parent1Path)
				}
				if !FolderExists(actualRoot) { // root should be kept
					t.Errorf("Expected directory %s (root) to be kept, but it was deleted.", actualRoot)
				}
			}
		})
	}
}
