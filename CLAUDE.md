# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # Build
go test ./...           # Run all tests
go test ./pkg/sanitization/...  # Run tests for a specific package
go run main.go copy --source /src --target /dst --dry-run  # Run directly
```

## Architecture

Muxic is a CLI music organization tool built with Cobra. It copies/moves music files from a source directory into a structured `Artist/Album/TrackNum - Title.ext` hierarchy at the target.

### Command flow (copy/move)

`cmd/copy.go` → fans out to N goroutines (one per CPU core) → each worker:
1. Reads ID3 tags via `pkg/metadata` (uses `github.com/dhowden/tag`)
2. Generates sanitized destination path via `movemusic.SuggestDestinationPath()` → `pkg/sanitization`
3. Copies file via `movemusic.CopyMusic()`, or copies + deletes source via `movemusic.MoveMusic()`

Progress counters use `atomic.AddInt64`; the sanitizer's title-caser is protected by a mutex.

### Key packages

- **`movemusic/`** — core copy/move logic, filename construction, FLAC-over-MP3 upgrade (auto-removes MP3 if FLAC exists at destination)
- **`musicutils/`** — recursive file discovery streaming results over a channel; supports name/size/duration filters
- **`pkg/sanitization/`** — 8-step Windows filesystem sanitization pipeline (Unicode transliteration, reserved char replacement, title casing, etc.)
- **`pkg/dedup/`** — SHA-256 signature cache persisted at `~/.muxic/dedup_cache.json`; used by `cmd/dedup.go`
- **`pkg/playlistfetch/`** — Spotify and YouTube Music playlist export; `Service` interface with OAuth via a local callback server on `:8080`
- **`pkg/config/`** — JSON config at `~/.muxic/config.json` (0600 perms) holding OAuth tokens for streaming services

### dedup command

Walks the target directory, groups files by SHA-256 signature, then either interactively asks which duplicate to keep or auto-deletes in "scorched earth" mode (keeps shortest path).

### playlist-fetch command

Requires credentials in `~/.muxic/config.json` (Spotify) or `~/.muxic/youtube_credentials.json` (YouTube). Writes playlists as `<outputDir>/<sanitizedName>.txt` with one `artist - album - trackNumber - title` line per track.
