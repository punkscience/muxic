# Technical Specification: muxic CLI

## 1. Overview

`muxic` is a command-line interface (CLI) application built with Go and the Cobra library. Its primary purpose is to organize music libraries by copying or moving audio files from a source directory to a structured target directory. The organization is based on the metadata (tags) embedded in the music files (e.g., artist, album). The application also performs file name sanitization.

## 2. Core Functionality

The application revolves around the `copy` command, which handles the scanning, processing, and transfer of music files.

### 2.1. `copy` Command

The `copy` command is the main entry point for the application's features.

**Usage:**

```bash
muxic copy --source <source_directory> --target <target_directory> [flags]
```

**Arguments & Flags:**

*   `--source` (string, required): Specifies the path to the source directory containing the music files to be processed.
*   `--target` (string, required): Specifies the path to the destination directory where the organized music files will be stored. If this directory does not exist, the application will attempt to create it.
*   `--filter` (string, optional): Filters the music files to be copied based on a case-insensitive substring match within the file path. Only files matching the filter will be processed.
*   `--over` (int, optional): Filters music files by size. Only files larger than the specified size in megabytes (MB) will be processed. Defaults to 0 (no size restriction).
*   `--move` or `-m` (boolean, optional): If set to `true`, the application will delete the original source file after a successful copy. This effectively "moves" the file. Default is `false` (standard copy).
*   `--verbose` or `-v` (boolean, optional): If set to `true`, the application will output detailed logs of its operations, including each file being processed and any errors encountered. Default is `false`.

**Behavior:**

1.  **Initialization:**
    *   Parses the provided command-line arguments and flags.
    *   Validates the presence of required arguments (`source`, `target`).

2.  **File Scanning:**
    *   Recursively scans the `source` directory for music files.
    *   Supported file types include: MP3, FLAC, M4A, WAV.
    *   If the `--filter` flag is used, only files whose paths contain the filter string (case-insensitive) are included.
    *   If the `--over` flag is used, only files exceeding the specified size (in MB) are included.

3.  **Processing and Organization:**
    *   For each identified music file:
        *   The application attempts to read metadata tags (e.g., artist, album) from the file.
        *   It uses this metadata to determine the appropriate subdirectory structure within the `target` folder (e.g., `<target>/<Artist>/<Album>/<TrackNumber> - <Title>.<extension>`).
        *   File names are sanitized: capitalization is standardized, and special characters are typically removed or replaced.
        *   The target directory structure is created if it doesn't already exist.

4.  **File Transfer:**
    *   The music file is copied from its `source` location to the determined `target` path.
    *   **Error Handling:**
        *   If the target file already exists, the copy operation for that specific file is skipped, and a message is logged.
        *   Other I/O errors during copying are logged, and the application continues with the next file.
    *   **Verbose Logging:** If `--verbose` is enabled, detailed information about each file operation is printed to the console.

5.  **Destructive Mode (`--move`):**
    *   If `--move` is enabled and a file is copied successfully:
        *   The original file at the `source` location is deleted.
        *   The application then attempts to recursively delete any empty parent directories in the `source` path from which the file was moved. This helps keep the source directory clean.

## 3. Music File Identification

The application identifies music files based on their extensions:
*   `.mp3`
*   `.flac`
*   `.m4a`
*   `.wav`

The `musicutils/musicutils.go` file contains the helper functions `GetAllMusicFiles` and `GetFilteredMusicFiles` responsible for locating these files.

## 4. File and Folder Operations

The `musicutils/musicutils.go` package also provides utilities for:

*   `FolderExists(folder string) bool`: Checks if a given folder path exists.
*   `IsDirEmpty(name string) (bool, error)`: Checks if a directory is empty. This is used during the cleanup phase of the `--move` operation.
*   `DeleteFile(file string)`: Deletes the specified file. If the `--move` operation is active, this function is also responsible for attempting to remove empty parent directories from the source.

The actual file copying and metadata-based organization logic leverages the `github.com/punkscience/movemusic` library, specifically the `movemusic.CopyMusic` function.

## 5. Error Handling and Logging

*   The application logs informational messages, warnings, and errors to standard output.
*   Specific errors handled include:
    *   File already exists at the target location (skipped).
    *   Errors creating the target directory.
    *   Errors during file copying.
    *   Errors deleting source files/folders (in `--move` mode).
*   Verbose logging (`--verbose`) provides more detailed insight into the application's operations.
