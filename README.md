# Muxic: Music Organization Utility

**Muxic** is a powerful and flexible command-line tool designed for organizing your music library efficiently. Built with Go, it offers a consistent, multi-platform solution to manage large music collections by copying or moving files into a structured directory layout based on their metadata (tags).

## Features

*   **Metadata-based Organization:** Automatically organizes music files into `Artist/Album/Track - Title.ext` folder structures using embedded tags.
*   **Flexible File Operations:**
    *   Copy music files to a new organized location.
    *   Move music files, deleting the originals and cleaning up empty source directories.
*   **File Name Sanitization:** Cleans up file names by standardizing capitalization and removing special characters.
*   **Support for Multiple Audio Formats:** Works with MP3, FLAC, M4A, and WAV files.
*   **Filtering Capabilities:**
    *   Filter files by name (case-insensitive substring match).
    *   Filter files by size (only process files larger than a specified MB).
*   **Verbose Logging:** Optional detailed output for monitoring all operations.
*   **Error Handling:** Gracefully handles common issues like existing files, inaccessible directories, and I/O errors.
*   **Cross-Platform:** As a Go application, `muxic` can be compiled to run on various operating systems.

## Getting Started

Currently, `muxic` is run from source or by building the Go project.

*(Further installation instructions, e.g., for binaries or package managers, would go here as the project evolves.)*

## Usage

The primary command for `muxic` is `copy`.

```bash
muxic copy --source <source_directory> --target <target_directory> [flags]
```

### Arguments & Flags:

*   `--source <path>`: **(Required)** Path to the directory containing your unsorted music files.
*   `--target <path>`: **(Required)** Path to the directory where organized music will be stored. `muxic` will create this directory if it doesn't exist.
*   `--filter <text>`: (Optional) Only process files whose path contains the specified text (case-insensitive).
    *   Example: `--filter "Rock"`
*   `--over <MB>`: (Optional) Only process files larger than the specified size in Megabytes.
    *   Example: `--over 5` (processes files > 5MB)
*   `-m`, `--move`: (Optional) Move files instead of copying. This will delete the source file after a successful transfer and attempt to clean up empty source subdirectories.
    *   Example: `muxic copy --source ./unsorted --target ./sorted --move`
*   `-v`, `--verbose`: (Optional) Enable detailed logging of all actions.
    *   Example: `muxic copy --source ./downloads --target ./MusicLib --verbose`
*   `-n`, `--dry-run`: (Optional) Report actions that would be taken without executing them. This allows you to see which files would be copied, moved, or deleted, and where they would go, without making any actual changes to your file system. Useful for previewing operations.

### Examples:

1.  **Basic Copy:**
    Copy all music from `./my_downloads` to `./organized_music_library`:
    ```bash
    muxic copy --source ./my_downloads --target ./organized_music_library
    ```

2.  **Move Files with Verbose Output:**
    Move all music from `./temp_music` to `./permanent_collection`, showing detailed logs:
    ```bash
    muxic copy --source ./temp_music --target ./permanent_collection --move --verbose
    ```

3.  **Copy Only Large FLAC Files:**
    Copy only FLAC files larger than 10MB from `./ripped_cds` to `./lossless_collection`:
    ```bash
    muxic copy --source ./ripped_cds --target ./lossless_collection --filter .flac --over 10
    ```
    *(Note: The current filter is by full path; a dedicated extension filter could be a future enhancement.)*

**Pro Tip:** Add the `--dry-run` or `-n` flag to any of these commands to see what would happen before committing to the changes!

## Contributing

Contributions are welcome! If you have ideas for new features, improvements, or bug fixes, please follow these steps:

1.  **Fork the repository.**
2.  **Create a new branch** for your feature or fix:
    ```bash
    git checkout -b feature/your-feature-name
    ```
3.  **Make your changes.** Ensure you add or update tests as appropriate.
4.  **Commit your changes** with a clear and descriptive commit message.
5.  **Push your branch** to your fork:
    ```bash
    git push origin feature/your-feature-name
    ```
6.  **Submit a Pull Request** for review.

Please ensure your code adheres to the project's coding standards and includes relevant documentation.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
