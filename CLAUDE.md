# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build                  # Build the binary
go run main.go [command]  # Run without building
go test ./...             # Run all tests
go test ./pkg/metadata/...  # Run tests in a single package
```

## Architecture

`muxic` is a Go CLI tool (Cobra) for organizing music files by reading ID3 tags and copying/moving files into a structured directory layout.

**Entry point:** `main.go` → `cmd.Execute()`

**Data flow for `copy`/`move`:**
1. `cmd/copy.go` parses flags and spawns a worker pool (`runtime.NumCPU()` goroutines)
2. `musicutils.GetAllMusicFiles()` / `GetFilteredMusicFiles()` streams file paths via a goroutine channel
3. `movemusic.CopyMusic()` / `MoveMusic()` handles each file:
   - `pkg/metadata.ReadTrackInfo()` extracts ID3 tags (artist, album, title, track number)
   - `pkg/sanitization.WindowsSanitizer` cleans metadata for Windows filesystem safety (Unicode transliteration → substitutions → prohibited chars → title casing)
   - `movemusic.SuggestDestinationPath()` builds the output path as `Artist/Album/NN - Title.ext`
   - `pkg/filesystem` handles actual file I/O and empty-directory pruning

**`dedup` command:** `cmd/dedup.go` → `pkg/dedup` uses SHA256 content hashing with a persistent JSON cache at `~/.muxic/dedup_cache.json` (keyed by path, invalidated by ModTime/Size).

## Key design notes

- **Sanitization is multi-step and Windows-first** — `pkg/sanitization/sanitization.go` is the most complex module. It handles edge cases like AC/DC-style acronyms and Unicode transliteration before title casing.
- **`ErrFileAlreadyExists`** is a sentinel error in `movemusic` — it's not counted as a failure and only logged under `--verbose`.
- **Dry-run** is threaded throughout via a `dryRun bool` flag passed into core functions; no actual file writes occur when true.
- The tech spec at `docs/technical_specification.md` tracks planned vs. completed features with ✅ markers — update it when implementing new features.
