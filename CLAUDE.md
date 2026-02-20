# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**cs (codespelunker)** is a command-line code search tool written in Go. It searches files recursively using boolean queries, regex, and fuzzy matching, with relevance ranking (BM25/TF-IDF). Four modes: console output, interactive TUI, HTTP server, and MCP server.

## Build & Test Commands

```bash
go build -o cs                        # Build binary
go test ./...                         # Run all tests
go test -v ./...                      # Verbose tests
go test -run TestPreParseQuery ./...  # Run a single test
```

**Linting:**
```bash
golangci-lint run --enable=gofmt ./...
gofmt -s -w -l .
```

## Architecture

Root package `main` plus subpackages under `pkg/`. The search pipeline is orchestrated by `DoSearch()` in `search.go`:

```
FileWalker → [read + filter files] → AST query evaluation → matched FileJobs channel
```

### Query Pipeline (`pkg/search/`)

Queries are parsed into an AST and evaluated against each file:

```
Lexer → Parser → Transformer → Planner → Executor
```

- **Lexer** (`lexer.go`): tokenizes query string (terms, quotes, regex, operators, fuzzy markers)
- **Parser** (`parser.go`): builds AST with boolean logic (AND/OR/NOT), quoted phrases, regex `/pattern/`, fuzzy `~1`/`~2`
- **Transformer** (`transformer.go`): semantic rewrites (e.g. `complexity=high`)
- **Planner** (`planner.go`): query optimization
- **Executor** (`executor.go`): evaluates AST against file content, returns match locations
- **Extractor** (`extractor.go`): extracts matched terms for highlighting
- **AST** (`ast.go`): AST node type definitions (AndNode, OrNode, NotNode, KeywordNode, PhraseNode, RegexNode, FuzzyNode, FilterNode)
- **Document** (`document.go`): Document and SearchResult structs

See `pkg/search/README.md` for detailed query syntax documentation.

### Core Root Files

- **`main.go`**: Cobra CLI entry point. Routes to `ConsoleSearch()`, TUI (`initialModel()`), `StartHttpServer()`, or `StartMcpServer()` based on flags/args
- **`config.go`**: `Config` struct with all CLI-configurable fields. `DefaultConfig()` provides sensible defaults
- **`search.go`**: `DoSearch()` — orchestrates the full search pipeline: file walking, reading, binary/minified filtering, AST evaluation, and result streaming via channels. Uses `SearchCache` for prefix-based caching
- **`console.go`**: `ConsoleSearch()` — collects results, ranks, and outputs in text/JSON/vimgrep format
- **`tui.go`**: bubbletea TUI with lipgloss styling. Debounced search input, incremental result streaming, syntax-highlighted preview
- **`http.go`**: HTTP server with embedded templates, search/display endpoints, theme support (dark/light/bare), custom template overrides
- **`cache.go`**: `SearchCache` — LRU cache with TTL for query results; supports prefix matching for progressive refinement in TUI
- **`syntax.go`**: Syntax highlighting with keyword tables for 80+ languages
- **`language.go`**: Language detection via scc processor's language database
- **`mcp.go`**: MCP (Model Context Protocol) server over stdio; exposes search as a tool for AI agents
- **`templates.go`**: Template loading with `//go:embed` for built-in HTML templates; supports custom template paths

### Subpackages

- **`pkg/common/`**: `FileJob` struct — the core data type passed through the pipeline
- **`pkg/ranker/`**: `RankResults()` — BM25 (default), TF-IDF, or simple ranking with location-based boosting. Includes declaration detection (`declarations.go`), deduplication (`dedup.go`), and stopword filtering (`stopwords.go`)
- **`pkg/snippet/`**: Snippet extraction (sliding window) and line-based extraction; auto-selects mode by file type

## Key Conventions

- Dependencies are vendored (`vendor/` directory)
- SPDX license headers on all source files (MIT)
- TUI uses bubbletea + bubbles (text input) + lipgloss (styling)
- Releases via GoReleaser (`.goreleaser.yaml`)
- Go 1.25.2, module: `github.com/boyter/cs`
