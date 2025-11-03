# Muxic

Muxic is an opinionated music organization utility designed to simplify the process of organizing your music library. It helps you copy or move music files from a source folder to a destination folder, creating a clean and organized structure based on the files' metadata (ID3 tags).

## Features

- **Organize by Metadata:** Automatically organizes music files into an `Artist/Album/Track` folder structure.
- **File Renaming:** Cleans and standardizes filenames, making your library consistent and easy to navigate.
- **Copy or Move:** Choose to either copy your files to a new location or move them, deleting the source files and cleaning up empty parent directories.
- **Dry-Run Mode:** Simulate any operation without making actual changes to your files, allowing you to preview the outcome.
- **Filtering:** Process only the files you want by filtering by name or file size.
- **Verbose Logging:** Get detailed output on every step of the process for better visibility.

## Installation

To install Muxic, you need to have Go installed on your system. You can then install Muxic using the following command:

`go get github.com/your-username/muxic`

## Usage

The primary command in Muxic is the `copy` command, which handles both copying and moving files.

### Copying Files

To copy files from a source directory to a target directory, use the following command:

`muxic copy --source /path/to/your/music --target /path/to/organized/music`

### Moving Files

To move files instead of copying them, use the `--move` or `-m` flag. This will delete the source files and any empty parent directories after a successful copy.

`muxic copy --source /path/to/your/music --target /path/to/organized/music --move`

### Flags

| Flag | Shorthand | Description |
|---|---|---|
| `--source` | | The source folder containing music files. |
| `--target` | | The destination folder where music files will be organized. |
| `--filter` | | Filter files by a string contained in their path (case-insensitive). |
| `--over` | | Only process files over this size in megabytes (MB). |
| `--move` | `-m` | Move files instead of copying (deletes source files and empty parent dirs). |
| `--verbose` | `-v` | Enable verbose logging for detailed operation output. |
| `--dry-run` | `-n` | Simulate operations without making any changes to the file system. |
