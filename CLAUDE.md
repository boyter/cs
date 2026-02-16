# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**cs (codespelunker)** is a command-line code search tool written in Go. It searches files recursively using boolean queries, regex, and fuzzy matching, with relevance ranking (BM25/TF-IDF). Three modes: console output, interactive TUI, and HTTP server.

## Build & Test Commands

```bash
go build -o cs                        # Build binary
go test ./...                         # Run all tests
go test -v ./...                      # Verbose tests
go test --tags=integration ./...      # With integration tests (used by check.sh)
go test -run TestPreParseQuery ./...  # Run a single test
```

**Linting (from check.sh):**
```bash
golangci-lint run --enable=gofmt ./...
gofmt -s -w -l .
```

## Architecture

All source files are in the root package `main`. The codebase uses a **concurrent worker pipeline** connected by Go channels:

```
FindFiles() → FileReaderWorker → SearcherWorker → ResultSummarizer/TUI/HTTP
```

1. **File discovery** (`file.go`): `walkFiles()` via gocodewalker, filters binary/minified files, respects .gitignore
2. **File reading** (`file.go`): `FileReaderWorker` reads files in parallel (NumCPU goroutines)
3. **Search** (`searcher.go`): `SearcherWorker` matches content in parallel, supports text/quoted/regex/fuzzy/negation
4. **Query parsing** (`search.go`): `PreParseQuery()` extracts `file:` filters, `ParseQuery()` handles boolean operators, quoted phrases, regex (`/pattern/`), fuzzy (`~1`, `~2`), and negation (`NOT`)
5. **Ranking** (`ranker.go`): `rankResults()` applies BM25 (default), TF-IDF, or simple ranking with location-based boosting for filename matches
6. **Snippet extraction** (`snippet.go`): Sliding window algorithm (`extractRelevantV3`) that finds the most relevant text around matches

**Entry point** (`main.go`): Cobra CLI routes to `StartHttpServer()`, `NewConsoleSearch()`, or `NewTuiSearch()` based on flags/args.

**Global configuration** (`globals.go`): All CLI flags and search parameters are package-level variables set by Cobra.

**Output formats** (`console.go`): text (with ANSI color), JSON, vimgrep.

## Key Conventions

- Dependencies are vendored (`vendor/` directory)
- SPDX license headers on all source files (MIT)
- TUI uses both `tview` and `bubbletea` (migration in progress)
- Releases via GoReleaser (`.goreleaser.yaml`)
