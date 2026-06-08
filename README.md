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

### Homebrew (macOS/Linux)

After a release tag is published, install using:

`brew install punkscience/muxic/muxic`

### APT / Debian (Linux)

Release tags generate a `.deb` package artifact in GitHub Releases.

Download and install:

`sudo apt install ./muxic_<version>_linux_amd64.deb`

Example:

`sudo apt install ./muxic_1.2.3_linux_amd64.deb`

### From source

To install from source, you need Go installed:

`go install github.com/punkscience/muxic@latest`

## CI/CD pipeline

- Pushes and pull requests run build + test in GitHub Actions (`.github/workflows/ci.yml`).
- Pushing a semver tag like `v1.2.3` triggers the release workflow (`.github/workflows/release.yml`).
- Release automation uses GoReleaser (`.goreleaser.yaml`) to:
  - Build binaries for Linux, macOS, and Windows
  - Publish release archives/checksums to GitHub Releases
  - Create Debian (`.deb`) packages for APT installation
  - Publish/update a Homebrew formula in `punkscience/homebrew-muxic`
- Required repository secret for release publishing:
  - `HOMEBREW_TAP_GITHUB_TOKEN` (PAT with `contents:write` on the tap repository)

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
