# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o muxic ./...

# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/metadata/...
go test ./movemusic/...

# Run a single test
go test -run TestFunctionName ./pkg/sanitization/...

# Install locally
go install ./...
```

## Architecture

**Muxic** is a Go CLI tool that organizes music libraries by reading ID3 metadata and restructuring files into `Artist/Album/TrackNum - Title.ext` paths.

### Package Layout

- **`cmd/`** — Cobra command definitions. `copy.go` handles the primary copy/move workflow; `dedup.go` finds and removes duplicates.
- **`movemusic/`** — Core file operation logic. `SuggestDestinationPath()` generates the target path from metadata; `CopyMusic()` and `MoveMusic()` execute the operation.
- **`musicutils/`** — Discovers music files (`.mp3`, `.flac`, `.m4a`, `.wav`) via `GetAllMusicFiles()` and `GetFilteredMusicFiles()` which return channels.
- **`pkg/metadata/`** — Reads ID3/audio tags using `dhowden/tag`.
- **`pkg/sanitization/`** — Sanitizes metadata strings for Windows filesystem compatibility (removes illegal chars, handles unicode via `gounidecode`).
- **`pkg/filesystem/`** — Low-level file operations: existence checks, deletion, empty-directory pruning.
- **`pkg/dedup/`** — SHA-256 based duplicate detection with a cache for performance.

### Key Data Flow (copy command)

1. `cmd/copy.go` validates flags and calls `musicutils.GetAllMusicFiles()` or `GetFilteredMusicFiles()`, which returns a `<-chan string`.
2. A worker pool (`runtime.NumCPU()` goroutines) consumes the channel.
3. Each worker calls `movemusic.CopyMusic()` or `MoveMusic()`, which internally calls `metadata.ReadMetadata()` → `SuggestDestinationPath()` → filesystem operations.
4. `MoveMusic()` also prunes empty source directories after a successful copy+delete.

### Conventions

- CLI commands use **Cobra** (`github.com/spf13/cobra`). New commands should be added to `cmd/` and registered in `init()`.
- All new functionality should have accompanying tests using `testify` (`github.com/stretchr/testify`).
- The sanitization package has a `Sanitizer` interface — use it when adding alternative sanitization strategies.
- `tech-spec.md` (not currently present) is referenced in `.github/copilot/rules.md` as the source of truth for planned features; mark items complete there when implementing them.
