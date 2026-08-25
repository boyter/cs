// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/cs/v3/pkg/ranker"
	"github.com/boyter/cs/v3/pkg/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// mcpSearchResponse wraps search results with metadata so callers always know
// which directory was searched, whether results were truncated, and what the
// full match count was.
type mcpSearchResponse struct {
	SearchedDirectory string       `json:"searched_directory"`
	TotalMatches      int          `json:"total_matches"`
	ResultsReturned   int          `json:"results_returned"`
	Offset            int          `json:"offset"`
	NextOffset        int          `json:"next_offset"`
	Truncated         bool         `json:"truncated"`
	Message           string       `json:"message,omitempty"`
	GitSyncNote       string       `json:"git_sync_note,omitempty"`
	Results           []jsonResult `json:"results"`
}

// mcpFileResult is the JSON response for the get_file tool.
type mcpFileResult struct {
	Language   string `json:"language,omitempty"`
	Lines      int64  `json:"lines,omitempty"`
	Code       int64  `json:"code,omitempty"`
	Comment    int64  `json:"comment,omitempty"`
	Blank      int64  `json:"blank,omitempty"`
	Complexity int64  `json:"complexity,omitempty"`
	Content    string `json:"content"`
}

// defaultSearchRoot returns the root a call uses when the caller passes no
// "path": the configured --dir, or the process working directory. This is also
// the tree --git-sync keeps up to date.
func defaultSearchRoot(cfg *Config) (string, error) {
	dir := strings.TrimSpace(cfg.Directory)
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot determine working directory: %w", err)
		}
		dir = wd
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve directory %s: %w", dir, err)
	}
	return abs, nil
}

// expandHome expands a leading ~ or ~/ to the user's home directory. Agents
// routinely pass shell-style paths that were never expanded by a shell.
func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// withinRoot reports whether target is root itself or lives beneath it.
func withinRoot(root, target string) bool {
	return target == root || strings.HasPrefix(target, root+string(filepath.Separator))
}

// resolveSearchRoot turns the caller-supplied "path" argument into the absolute
// directory to walk. It reports whether the caller named a root explicitly, so
// the handler can honour it literally rather than letting --find-root climb out
// of it. Relative paths resolve against the default root, which is the agent's
// notion of "here".
func resolveSearchRoot(cfg *Config, raw string) (root string, explicit bool, err error) {
	base, err := defaultSearchRoot(cfg)
	if err != nil {
		return "", false, err
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return base, false, nil
	}

	// A glob here is almost always an agent reaching for the old filter
	// semantics. Fail loudly and point at the parameter it meant.
	if strings.ContainsAny(raw, "*?[") {
		return "", false, fmt.Errorf(
			"path %q looks like a pattern, but path is the directory to search in. "+
				"To filter results by path use path_filter=%q instead", raw, raw)
	}

	resolved := expandHome(raw)
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(base, resolved)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", false, fmt.Errorf("cannot resolve path %q: %v", raw, err)
	}

	info, statErr := os.Stat(resolved)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", false, fmt.Errorf("path does not exist: %s", resolved)
		}
		return "", false, fmt.Errorf("cannot access path %s: %v", resolved, statErr)
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf(
			"path must be a directory to search in, but %s is a file. Use get_file to read a single file",
			resolved)
	}

	if cfg.MCPLockDir && !withinRoot(base, resolved) {
		return "", false, fmt.Errorf(
			"path %s is outside %s and this server was started with --mcp-lock-dir",
			resolved, base)
	}

	return resolved, true, nil
}

// StartMCPServer starts an MCP server over stdio, exposing a "search" tool
// that uses the same DoSearch pipeline as console and HTTP modes.
func StartMCPServer(cfg *Config) {
	cache := NewSearchCache()

	mcpServer := server.NewMCPServer(
		"codespelunker",
		Version,
		server.WithToolCapabilities(false),
	)

	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search code files recursively using boolean queries, regex, and fuzzy matching with relevance ranking.\n\n"+
			"Two modes:\n"+
			"- Default: finds and ranks the most relevant FILES. Use when discovering where something lives.\n"+
			"- Grep (snippet_mode='grep'): returns every matching LINE with context (like grep -C). Use when reading/tracing code: understanding implementations, following call chains, seeing all usages. Pair with 'context' (e.g. 5-15) and optionally 'line_limit'.\n\n"+
			"Query syntax:\n"+
			"- Keywords: terms are ANDed by default (e.g. 'jwt middleware' finds files with both terms)\n"+
			"- OR: 'error OR exception' matches either term\n"+
			"- NOT: 'NOT path:vendor' excludes matches\n"+
			"- Grouping: '(auth OR login) AND handler'\n"+
			"- Phrases: '\"exact phrase\"' for exact match\n"+
			"- Regex: wrapping in forward slashes `/pattern/` activates regex mode. Without slashes, terms are treated as keywords, NOT regex. Example: `/func\\s+Test/`, `/TODO|FIXME/`\n"+
			"  Keywords: `magic number` → files with both words. Regex: `/magic.*number/` → pattern match. Phrase: `\"magic number\"` → exact sequence.\n"+
			"- Fuzzy: 'term~1' or 'term~2' for typo-tolerant matching (Levenshtein distance 1 or 2)\n\n"+
			"Filters (in-query):\n"+
			"- file:pattern — match filename (substring or glob: file:*.go, file:*_test.go)\n"+
			"- path:pattern — match full path (substring or glob: path:*/pkg/*, NOT path:vendor/*/*)\n"+
			"- lang:value — filter by language: lang:go, lang=go,python (multi-value with commas)\n"+
			"- ext:value — filter by extension: ext:go, ext=ts,tsx\n\n"+
			"Filter operators: = != (e.g. lang!=python, file!=test)\n"+
			"Negation: NOT file:test, file!=test, NOT path:vendor, path!=vendor\n\n"+
			"IMPORTANT — filter precedence: an in-query filter binds to the ADJACENT term, not the whole query. "+
			"So 'a OR b path:src' parses as 'a OR (b AND path:src)' and matches for 'a' leak in from ANY path. "+
			"To scope the whole query, either group it — '(a OR b) path:src' — or use the top-level 'path_filter'/'file' "+
			"parameters, which are ANDed with the entire query regardless of OR grouping. Same rule applies to lang:/ext:.\n\n"+
			"Top-level filter parameters (path_filter, file, include_ext, language) are the ROBUST way to scope a search: "+
			"they apply to the whole query, so they never hit the precedence trap above. Prefer them over in-query filters "+
			"when you just want to restrict a search to a directory, filename, extension, or language. "+
			"NOTE: there is no top-level 'ext' parameter — the extension parameter is named 'include_ext'. "+
			"Unknown parameters are rejected with an error rather than silently ignored.\n\n"+
			"WHERE TO SEARCH ('path' parameter): 'path' is the DIRECTORY to search in, exactly like grep/rg. "+
			"Omit it to search the server's default directory. Pass it to search anywhere else on the machine — "+
			"another repository, a subdirectory, an absolute path. It is NOT a filter: to narrow results by path use "+
			"'path_filter'. Passing a glob to 'path' is an error.\n\n"+
			"Content type filter (code_filter parameter):\n"+
			"- 'only-code': matches in code only, skipping comments and strings — e.g. find where a function is called, not just mentioned\n"+
			"- 'only-strings': matches in string literals only — find SQL queries, error messages, config values, connection strings\n"+
			"- 'only-comments': matches in comments only — find TODOs, developer explanations, annotations\n"+
			"- 'only-declarations': matches only on declaration lines (func, type, class, def, struct, etc.) — find where something is defined\n"+
			"- 'only-usages': matches only on non-declaration lines — find where something is called/referenced (impact analysis)\n\n"+
			"Combined examples:\n"+
			"- 'jwt middleware lang:go NOT path:vendor' — find Go JWT middleware outside vendor\n"+
			"- query='dense_rank' code_filter='only-strings' — find the actual SQL string, not code references\n"+
			"- query='middleware' code_filter='only-code' path_filter='src' — find middleware implementations scoped to src via the top-level path_filter param\n"+
			"- query='parser' path='/home/me/other-repo' — search a different checkout entirely\n"+
			"- query='authentication' code_filter='only-comments' — find where devs explain auth flow\n"+
			"- query='ConnectDB' code_filter='only-declarations' language='Go' — find where ConnectDB is defined (func/type/var declaration)\n"+
			"- query='ConnectDB' code_filter='only-usages' language='Go' — find all call sites of ConnectDB, excluding its definition\n\n"+
			"Tips and common mistakes:\n"+
			"- Terms are ANDed: 'sql.Open pgx.Connect mongo.Connect' requires ALL terms in one file. Use OR for alternatives: 'sql.Open OR pgx.Connect OR mongo.Connect'\n"+
			"- Too many AND terms = no results. Start with 1-2 specific terms, then narrow with filters.\n"+
			"- Dot-separated names (sql.Open, fmt.Println) work as literal substrings. Quoting is optional: sql.Open and \"sql.Open\" behave identically.\n"+
			"- Exclude dependency dirs: add 'NOT path:vendor NOT path:node_modules' to avoid vendored/dependency results.\n"+
			"- File exclusion with many AND terms: 'process calculate transform aggregate NOT file:*_test.go' fails because no file contains all four keywords. Reduce terms: 'process aggregate NOT file:*_test.go lang:go'\n"+
			"- For structural patterns use regex: '/type\\s+\\w+Error\\s+struct/' not 'type Error struct'. Keywords match anywhere in the file, not adjacently.\n"+
			"- Common regex mistake: `magic.*number` without slashes is treated as the keyword 'magic.*number', not as regex. Always wrap in slashes: `/magic.*number/`.\n"+
			"- NOT binds to the next term only, not the whole query. 'a OR b NOT path:vendor' means 'a OR (b AND NOT path:vendor)'. To exclude globally, use grouping: '(a OR b) NOT path:vendor'. Precedence: NOT (tightest) > AND > OR (loosest).\n"+
			"- The same trap applies to POSITIVE filters: 'a OR b path:src' means 'a OR (b AND path:src)', so 'a' matches leak in from any path. Group the query — '(a OR b) path:src' — or use the top-level 'path_filter'/'file'/'include_ext'/'language' parameters, which scope the whole query.\n"+
			"- max_results defaults to 20. Set higher (e.g. 100) for broad discovery or exploring unfamiliar code.\n\n"+
			"Workflow tips:\n"+
			"- Searching for a specific term, identifier, or function name → use snippet_mode='grep' with context=5-10. This gives every occurrence with surrounding code in one call.\n"+
			"- Conceptual or discovery queries ('how does auth work', 'what handles errors') → use the default auto mode. The ranker surfaces the most relevant files.\n"+
			"- Once a specific file is identified, switch to get_file to read it — don't keep searching the same file.\n"+
			"- Searching a repository other than the default one → pass its directory as 'path'. Note that only the default directory is kept up to date by --git-sync; other directories are searched exactly as they are on disk."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query",
			mcp.Description("The search query. Terms are ANDed by default. Supports: OR ('error OR exception'), NOT ('NOT vendor'), "+
				"grouping ('(auth OR login) AND handler'), quoted phrases ('\"exact match\"'), regex (/pattern/), fuzzy (term~1, term~2). "+
				"In-query filters: file:name, path:dir, lang:go, ext:ts. Operators: = != (lang!=python, file!=test). "+
				"Multi-value: lang=go,python, ext=ts,tsx. 'file:' matches filename only; 'path:' matches the full directory path. "+
				"NOTE: in-query filters bind to the adjacent term — 'a OR b path:src' means 'a OR (b AND path:src)'. Group with parens "+
				"or use the top-level 'path_filter'/'file'/'include_ext'/'language' parameters to scope the whole query. "+
				"Query limits: max 250 characters and 12 unique search terms."),
			mcp.Required(),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return. Defaults to 20. No upper limit enforced. Use higher values (50-100) for broad discovery queries or when exploring unfamiliar codebases."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of results to skip before returning. Use for pagination: pass 'next_offset' from a previous response as 'offset' to get the next page. Defaults to 0."),
		),
		mcp.WithNumber("snippet_length",
			mcp.Description("Size of the code snippet to display in characters."),
		),
		mcp.WithBoolean("case_sensitive",
			mcp.Description("Make the search case sensitive."),
		),
		mcp.WithString("include_ext",
			mcp.Description("Comma-separated list of file extensions to search (e.g. \"go,js,py\"). Convenience parameter equivalent to in-query 'ext:go,js,py' filter."),
		),
		mcp.WithString("language",
			mcp.Description("Comma-separated list of language types to search (e.g. \"Go,Python,JavaScript\"). Convenience parameter equivalent to in-query 'lang:Go,Python' filter."),
		),
		mcp.WithString("path",
			mcp.Description("The DIRECTORY to search in, like grep/rg. Defaults to the directory the server was started with "+
				"(--dir, or its working directory). Absolute paths are used as-is; relative paths resolve against that default "+
				"directory; a leading ~ is expanded. Use this to search a different repository or a subdirectory. "+
				"Must be an existing directory — a glob or a file path is rejected. "+
				"NOTE: this is not a filter. To narrow results by path use 'path_filter'. "+
				"Only the default directory is refreshed by --git-sync; other directories are searched as they are on disk. "+
				"The resolved directory is echoed back as 'searched_directory' in the response."),
		),
		mcp.WithString("path_filter",
			mcp.Description("Restrict results to files whose full path matches. Substring match (e.g. \"src/handlers\"), or glob if it "+
				"contains */?/[ (e.g. \"*/pkg/*\"). Comma-separated values are ORed (e.g. \"src,internal\" = under src OR internal). "+
				"Applied as an AND against the ENTIRE query, so — unlike an in-query 'path:' filter — it is never subject to OR "+
				"precedence and cannot be leaked past. This is the robust way to scope a search to a subdirectory of whatever "+
				"'path' is being searched. Equivalent to wrapping your query: '(<query>) path:<value>'."),
		),
		mcp.WithString("file",
			mcp.Description("Restrict results to files whose FILENAME matches (not the directory path — use 'path' for that). Substring match "+
				"(e.g. \"handler\"), or glob if it contains */?/[ (e.g. \"*_test.go\"). Comma-separated values are ORed. "+
				"Applied as an AND against the entire query, so it is never subject to OR precedence. Equivalent to "+
				"wrapping your query: '(<query>) file:<value>'."),
		),
		mcp.WithString("gravity",
			mcp.Description("Complexity gravity intent controlling how much cyclomatic complexity boosts ranking. "+
				"Values: brain (2.5) — find complex algorithmic code, logic (1.5) — prefer branching/control flow, "+
				"default (1.0) — balanced, low (0.2) — mostly ignore complexity, off (0.0) — pure text relevance only."),
		),
		mcp.WithString("profile",
			mcp.Description("Ranking profile — a preset that tunes multiple ranking parameters at once. "+
				"Values: balanced (default) — general-purpose ranking. "+
				"precise — favours short focused source files, penalises long files and test files, "+
				"best for 'find the one file that matters'. "+
				"broad — rewards repeated matches, includes test files at full weight, "+
				"best for 'show me everything relevant'. "+
				"Overrides gravity, noise, and test-penalty settings when set."),
		),
		mcp.WithBoolean("dedup",
			mcp.Description("Collapse byte-identical search matches, keeping the highest-scored representative. Useful in monorepos with duplicated code."),
		),
		mcp.WithString("code_filter",
			mcp.Description("Content type filter — narrows matches to a specific part of the source file.\n"+
				"Values:\n"+
				"- 'only-code': match only in executable code lines (skip comments and string literals). "+
				"Use when searching for function calls, variable usage, or control flow.\n"+
				"- 'only-strings': match only in string literals. "+
				"Use when searching for SQL queries (e.g. 'dense_rank'), error messages, log messages, config keys, dependency names, or connection strings.\n"+
				"- 'only-comments': match only in comments. "+
				"Use when searching for TODOs, FIXMEs, developer explanations of complex logic, or doc annotations.\n"+
				"- 'only-declarations': match only on declaration lines (func, type, class, def, struct, const, var, interface, enum, trait, impl, etc.). "+
				"Use to find where a function, type, class, or variable is DEFINED — answers 'where is this declared?'. "+
				"Works by matching line-start heuristics after trimming whitespace, so indented methods/functions inside classes are detected. "+
				"Supported languages: Go, Python, JavaScript, TypeScript, TSX, Rust, Java, C, C++, C#, Ruby, PHP, Kotlin, Swift. "+
				"Files in unsupported languages are excluded (conservative: can't identify declarations without patterns).\n"+
				"- 'only-usages': match only on non-declaration lines (inverse of only-declarations). "+
				"Use for impact analysis — answers 'where is this called/referenced?'. "+
				"Returns every match that is NOT on a declaration line. "+
				"For unsupported languages, all matches are returned (conservative: if we can't identify declarations, everything is a usage).\n"+
				"Default: no filter (searches all content types).\n"+
				"IMPORTANT: When using code_filter, always also set the 'language' parameter. Example: code_filter='only-comments' language='Go'. Without language, results from all languages (including node_modules, vendor, site-packages) will dominate.\n"+
				"NOTE: only-declarations/only-usages are mutually exclusive with only-code/only-comments/only-strings. Only one code_filter value can be active at a time."),
		),
		mcp.WithString("snippet_mode",
			mcp.Description("Snippet extraction mode. Valid values: 'auto' (default), 'snippet', 'lines', 'grep'.\n"+
				"DEFAULT TO GREP for any query containing a specific known term, identifier, function name, or keyword. Only use 'auto' for broad conceptual or discovery queries where you do not know the exact term.\n\n"+
				"WHEN TO USE GREP:\n"+
				"- You are searching for a specific term, identifier, or function name\n"+
				"- You need exhaustive results (every occurrence, not a ranked subset)\n"+
				"- You are tracing a function through call sites or following a value through code\n"+
				"- The query intent is 'where is X', 'find all X', 'how is X used', 'show me every X'\n"+
				"Returns every matching line with context (like grep -C). You see ALL matches.\n\n"+
				"WHEN NOT TO USE GREP:\n"+
				"- Conceptual or discovery queries ('how does auth work', 'what handles errors')\n"+
				"- You want the ranker to surface the most relevant files, not every mention\n"+
				"- The query is broad and would produce hundreds of matches\n"+
				"For these, use 'auto' — it returns ranked, relevance-focused snippets.\n\n"+
				"GREP SETTINGS:\n"+
				"- Always pair with 'context': 5 for quick lookups, 10-15 for understanding logic flow\n"+
				"- Use 'line_limit' to cap output for high-frequency terms (e.g. line_limit=5)\n\n"+
				"Example: query='BM25' snippet_mode='grep' context=10 — find every occurrence of BM25 with surrounding code."),
		),
		mcp.WithNumber("line_limit",
			mcp.Description("Max matching lines per file in grep mode. Defaults to -1 (unlimited). Only applies when snippet_mode is 'grep'."),
		),
		mcp.WithNumber("context_before",
			mcp.Description("Lines of context to show before each matching line in grep mode (like grep -B). Only applies when snippet_mode is 'grep'."),
		),
		mcp.WithNumber("context_after",
			mcp.Description("Lines of context to show after each matching line in grep mode (like grep -A). Only applies when snippet_mode is 'grep'."),
		),
		mcp.WithNumber("context",
			mcp.Description("Lines of context to show before and after each matching line in grep mode (like grep -C). "+
				"Sets both context_before and context_after. Individual context_before/context_after override this value. "+
				"ALWAYS set this when using grep mode — omitting it gives bare matching lines with no surrounding code, which is rarely useful. "+
				"Start with context=5 for quick identifier lookups. Use context=10-15 when you need to understand surrounding logic flow."),
		),
	)

	mcpServer.AddTool(searchTool, mcpSearchHandler(cfg, cache))

	getFileTool := mcp.NewTool("get_file",
		mcp.WithDescription("Read a file's full contents by path. Prefer this over repeated searches once a file is identified — search snippets are truncated and miss logic between matches. Use start_line/end_line for large files. Returns JSON with line-numbered 'content' and, for source files, language/complexity stats."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("path",
			mcp.Description("File path relative to the project directory, or absolute path within the project."),
			mcp.Required(),
		),
		mcp.WithNumber("start_line",
			mcp.Description("1-based start line number. If omitted, reads from the beginning."),
		),
		mcp.WithNumber("end_line",
			mcp.Description("1-based end line number (inclusive). If omitted, reads to the end."),
		),
	)

	mcpServer.AddTool(getFileTool, mcpGetFileHandler(cfg))

	// stdout is reserved for MCP JSON-RPC; log to stderr
	errLogger := log.New(os.Stderr, "cs-mcp: ", log.LstdFlags)
	if err := server.ServeStdio(mcpServer, server.WithErrorLogger(errLogger)); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

// mcpGetFileHandler returns an MCP tool handler that reads a file's contents.
func mcpGetFileHandler(cfg *Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: path"), nil
		}
		if strings.TrimSpace(path) == "" {
			return mcp.NewToolResultError("path must not be empty"), nil
		}

		// Resolve relative paths against the default search root, so a bare
		// filename means the same thing here as it does in a default search.
		absProject, err := defaultSearchRoot(cfg)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve project directory: %v", err)), nil
		}
		resolved := expandHome(path)
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(absProject, resolved)
		}
		absResolved, err := filepath.Abs(resolved)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve file path: %v", err)), nil
		}

		// Search can name any directory, so get_file must be able to read what
		// search returned. --mcp-lock-dir restores the single-tree restriction.
		if cfg.MCPLockDir && !withinRoot(absProject, absResolved) {
			return mcp.NewToolResultError(fmt.Sprintf(
				"path is outside %s and this server was started with --mcp-lock-dir", absProject)), nil
		}

		// Read the file
		content, err := readFileContent(absResolved, cfg.MaxReadSizeBytes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
		}

		// Binary detection: check first 10KB for NUL bytes
		check := content
		if len(check) > 10_000 {
			check = content[:10_000]
		}
		if bytes.IndexByte(check, 0) != -1 {
			return mcp.NewToolResultError("file appears to be binary"), nil
		}

		// Detect language and compute code stats
		lang, sccLines, sccCode, sccComment, sccBlank, sccComplexity, _ := fileCodeStats(filepath.Base(absResolved), content)

		lines := strings.Split(string(content), "\n")

		// Apply optional line range
		startLine := 1
		endLine := len(lines)
		if v, ok := request.GetArguments()["start_line"]; ok {
			if n, ok := v.(float64); ok && n >= 1 {
				startLine = int(n)
			}
		}
		if v, ok := request.GetArguments()["end_line"]; ok {
			if n, ok := v.(float64); ok && n >= 1 {
				endLine = int(n)
			}
		}

		if startLine > len(lines) {
			return mcp.NewToolResultError(fmt.Sprintf("start_line %d exceeds file length of %d lines", startLine, len(lines))), nil
		}
		if endLine > len(lines) {
			endLine = len(lines)
		}
		if startLine > endLine {
			return mcp.NewToolResultError(fmt.Sprintf("start_line %d is greater than end_line %d", startLine, endLine)), nil
		}

		// Format line-numbered content
		var sb strings.Builder
		for i := startLine; i <= endLine; i++ {
			fmt.Fprintf(&sb, "%d\t%s\n", i, lines[i-1])
		}

		result := mcpFileResult{
			Content: sb.String(),
		}
		if lang != "" {
			result.Language = lang
			result.Lines = sccLines
			result.Code = sccCode
			result.Comment = sccComment
			result.Blank = sccBlank
			result.Complexity = sccComplexity
		}
		jsonResult, err := mcp.NewToolResultJSON(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		return jsonResult, nil
	}
}

// mcpSearchParams is the set of accepted parameters for the search tool, in a
// stable order for error messages. Any argument not in this set is rejected so
// that malformed calls (e.g. a non-existent "ext" or "dir" top-level key) fail
// loudly instead of being silently dropped and running an unfiltered search.
var mcpSearchParams = []string{
	"query", "path", "max_results", "offset", "snippet_length", "case_sensitive",
	"include_ext", "language", "path_filter", "file", "gravity", "profile", "dedup",
	"code_filter", "snippet_mode", "line_limit", "context", "context_before", "context_after",
}

var mcpSearchParamSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(mcpSearchParams))
	for _, p := range mcpSearchParams {
		m[p] = struct{}{}
	}
	return m
}()

// unknownSearchParams returns any argument keys that are not accepted params,
// sorted for a stable error message.
func unknownSearchParams(args map[string]any) []string {
	var unknown []string
	for k := range args {
		if _, ok := mcpSearchParamSet[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// buildFilterGroup turns a comma-separated top-level filter value into an
// in-query clause for the given field ("path" or "file"). Multiple values are
// ORed and wrapped in parens so the group evaluates as a single unit. Returns
// "" when there is nothing to add.
func buildFilterGroup(field, raw string) string {
	var clauses []string
	for _, v := range strings.Split(raw, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		clauses = append(clauses, field+":"+v)
	}
	switch len(clauses) {
	case 0:
		return ""
	case 1:
		return clauses[0]
	default:
		return "(" + strings.Join(clauses, " OR ") + ")"
	}
}

// composeSearchQuery folds the top-level path/file filters into the query so
// they AND against the WHOLE query, immune to OR precedence. The user query is
// wrapped in parens and each filter group is appended (juxtaposition = AND).
// The composed string flows through the normal parser and cache key, so it
// behaves exactly as if the caller had written the filters in-query themselves.
func composeSearchQuery(query, pathFilter, fileFilter string) string {
	var groups []string
	if g := buildFilterGroup("path", pathFilter); g != "" {
		groups = append(groups, g)
	}
	if g := buildFilterGroup("file", fileFilter); g != "" {
		groups = append(groups, g)
	}
	if len(groups) == 0 {
		return query
	}
	return "(" + query + ") " + strings.Join(groups, " ")
}

// andedKeywordsForHint returns the positive, bare keyword terms of a query when
// the query is a pure conjunction of keywords (contains no OR). It returns nil
// when the query already uses OR, or has fewer than two positive keyword terms —
// i.e. when an "all terms ANDed" hint would not apply. Phrases, regex, fuzzy and
// filter terms are ignored, and negated keywords (under NOT) are excluded.
func andedKeywordsForHint(query string) []string {
	lexer := search.NewLexer(strings.NewReader(query))
	parser := search.NewParser(lexer)
	ast, _ := parser.ParseQuery()
	if ast == nil {
		return nil
	}

	var keywords []string
	hasOr := false
	var walk func(n search.Node, negated bool)
	walk = func(n search.Node, negated bool) {
		switch t := n.(type) {
		case *search.AndNode:
			walk(t.Left, negated)
			walk(t.Right, negated)
		case *search.OrNode:
			hasOr = true
		case *search.NotNode:
			walk(t.Expr, !negated)
		case *search.KeywordNode:
			if !negated {
				keywords = append(keywords, t.Value)
			}
		}
	}
	walk(ast, false)

	if hasOr || len(keywords) < 2 {
		return nil
	}
	return keywords
}

// andHintMessage builds the empty-result explanation for a multi-term AND query,
// nudging the caller toward fewer terms or an OR query.
func andHintMessage(keywords []string) string {
	quoted := make([]string, len(keywords))
	for i, k := range keywords {
		quoted[i] = "\"" + k + "\""
	}
	return fmt.Sprintf(
		"0 results. Terms are ANDed, so this requires all %d of %s in a single file. "+
			"Try fewer/more-specific terms, or OR them to match any: '%s'.",
		len(keywords), strings.Join(quoted, ", "), strings.Join(keywords, " OR "))
}

// mcpSearchHandler returns an MCP tool handler that runs a code search.
func mcpSearchHandler(cfg *Config, cache *SearchCache) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: query"), nil
		}
		if strings.TrimSpace(query) == "" {
			return mcp.NewToolResultError("query must not be empty"), nil
		}

		// Reject unknown parameters so malformed calls fail loudly instead of
		// silently running an unfiltered search (e.g. a caller passing a
		// non-existent top-level "ext" or "dir" key and getting whole-repo results).
		if unknown := unknownSearchParams(request.GetArguments()); len(unknown) > 0 {
			return mcp.NewToolResultError(fmt.Sprintf(
				"unknown parameter(s): %s. To choose the directory to search use 'path'. To filter by path, "+
					"filename, extension, or language use the 'path_filter', 'file', 'include_ext', or 'language' "+
					"parameters (there is no top-level 'ext' parameter), or in-query filters (path:, file:, ext:, lang:). "+
					"Accepted parameters: %s.",
				strings.Join(unknown, ", "), strings.Join(mcpSearchParams, ", "))), nil
		}

		// Resolve the directory to search. "path" is the root, like grep.
		searchRoot := ""
		if v, ok := request.GetArguments()["path"]; ok {
			if s, ok := v.(string); ok {
				searchRoot = s
			}
		}
		resolvedRoot, explicitRoot, rootErr := resolveSearchRoot(cfg, searchRoot)
		if rootErr != nil {
			return mcp.NewToolResultError(rootErr.Error()), nil
		}

		// Fold top-level path_filter/file filters into the query so they AND
		// against the whole query, immune to OR precedence.
		pathFilter := ""
		if v, ok := request.GetArguments()["path_filter"]; ok {
			if s, ok := v.(string); ok {
				pathFilter = s
			}
		}
		fileFilter := ""
		if v, ok := request.GetArguments()["file"]; ok {
			if s, ok := v.(string); ok {
				fileFilter = s
			}
		}
		userQuery := query
		query = composeSearchQuery(query, pathFilter, fileFilter)

		// Copy config so we can override per-request without mutating the shared config
		searchCfg := *cfg
		searchCfg.Directory = resolvedRoot
		if explicitRoot {
			// The caller named this directory, so walk it literally rather than
			// letting --find-root climb out to the enclosing repository root.
			searchCfg.FindRoot = false
		}
		searchCfg.Format = "json"
		searchCfg.MaxQueryChars = common.MaxQueryCharsMCP
		searchCfg.MaxQueryTerms = common.MaxQueryTermsMCP

		// Apply optional parameters
		maxResults := 20
		if v, ok := request.GetArguments()["max_results"]; ok {
			if n, ok := v.(float64); ok && n > 0 {
				maxResults = int(n)
			}
		}
		offset := 0
		if v, ok := request.GetArguments()["offset"]; ok {
			if n, ok := v.(float64); ok && n >= 0 {
				offset = int(n)
			}
		}
		if v, ok := request.GetArguments()["snippet_length"]; ok {
			if n, ok := v.(float64); ok && n > 0 {
				searchCfg.SnippetLength = int(n)
			}
		}
		if v, ok := request.GetArguments()["case_sensitive"]; ok {
			if b, ok := v.(bool); ok {
				searchCfg.CaseSensitive = b
			}
		}
		if v, ok := request.GetArguments()["include_ext"]; ok {
			if s, ok := v.(string); ok && s != "" {
				searchCfg.AllowListExtensions = strings.Split(s, ",")
			}
		}
		if v, ok := request.GetArguments()["language"]; ok {
			if s, ok := v.(string); ok && s != "" {
				searchCfg.LanguageTypes = strings.Split(s, ",")
			}
		}
		if v, ok := request.GetArguments()["profile"]; ok {
			if s, ok := v.(string); ok && s != "" {
				searchCfg.Profile = s
			}
		}
		if v, ok := request.GetArguments()["gravity"]; ok {
			if s, ok := v.(string); ok && s != "" {
				searchCfg.GravityIntent = s
			}
		}
		if v, ok := request.GetArguments()["code_filter"]; ok {
			if s, ok := v.(string); ok && s != "" {
				// Clear all content filters before setting the requested one
				searchCfg.OnlyCode = false
				searchCfg.OnlyComments = false
				searchCfg.OnlyStrings = false
				searchCfg.OnlyDeclarations = false
				searchCfg.OnlyUsages = false
				switch s {
				case "only-code":
					searchCfg.OnlyCode = true
				case "only-comments":
					searchCfg.OnlyComments = true
				case "only-strings":
					searchCfg.OnlyStrings = true
				case "only-declarations":
					searchCfg.OnlyDeclarations = true
				case "only-usages":
					searchCfg.OnlyUsages = true
				}
				if searchCfg.HasContentFilter() {
					searchCfg.Ranker = "structural"
				}
			}
		}
		if v, ok := request.GetArguments()["snippet_mode"]; ok {
			if s, ok := v.(string); ok && s != "" {
				searchCfg.SnippetMode = s
			}
		}
		if v, ok := request.GetArguments()["line_limit"]; ok {
			if n, ok := v.(float64); ok {
				searchCfg.LineLimit = int(n)
			}
		}
		if v, ok := request.GetArguments()["context"]; ok {
			if n, ok := v.(float64); ok && n >= 0 {
				searchCfg.ContextAround = int(n)
			}
		}
		if v, ok := request.GetArguments()["context_before"]; ok {
			if n, ok := v.(float64); ok && n >= 0 {
				searchCfg.ContextBefore = int(n)
			}
		}
		if v, ok := request.GetArguments()["context_after"]; ok {
			if n, ok := v.(float64); ok && n >= 0 {
				searchCfg.ContextAfter = int(n)
			}
		}

		// Run search
		ch, stats, searchErr := DoSearch(ctx, &searchCfg, query, cache)
		if searchErr != nil {
			return mcp.NewToolResultError(searchErr.Error()), nil
		}

		var results []*common.FileJob
		for fj := range ch {
			results = append(results, fj)
		}

		// Rank results
		textFileCount := int(stats.TextFileCount.Load())
		testIntent := ranker.HasTestIntent(strings.Fields(query))
		results = ranker.RankResults(searchCfg.Ranker, textFileCount, results, searchCfg.StructuralRankerConfig(), searchCfg.ResolveRankingProfile(), testIntent)

		// Dedup (before limit, so freed slots get backfilled)
		if v, ok := request.GetArguments()["dedup"]; ok {
			if b, ok := v.(bool); ok && b {
				results = ranker.DeduplicateResults(results)
			}
		}

		// Track total before truncation so we can report honestly
		totalMatches := len(results)

		// Apply offset (skip first N results)
		if offset > 0 {
			if offset >= len(results) {
				results = nil
			} else {
				results = results[offset:]
			}
		}

		// Apply max_results limit to the offset slice
		truncated := false
		if maxResults > 0 && len(results) > maxResults {
			results = results[:maxResults]
			truncated = true
		}

		// Build JSON using the shared helper
		jsonResults := buildJSONResults(&searchCfg, results)

		// Calculate next_offset for pagination
		nextOffset := offset + len(jsonResults)
		if nextOffset > totalMatches {
			nextOffset = totalMatches
		}

		// Build response envelope with pagination metadata
		response := mcpSearchResponse{
			SearchedDirectory: resolvedRoot,
			TotalMatches:      totalMatches,
			ResultsReturned:   len(jsonResults),
			Offset:            offset,
			NextOffset:        nextOffset,
			Truncated:         truncated,
			Results:           jsonResults,
		}

		// git-sync only follows the default directory, so say so when the caller
		// searched somewhere else rather than letting them assume it is fresh.
		if cfg.GitSync && explicitRoot {
			if base, err := defaultSearchRoot(cfg); err == nil && !withinRoot(base, resolvedRoot) {
				response.GitSyncNote = fmt.Sprintf(
					"%s is outside the --git-sync directory (%s) and was searched as it is on disk; it may be stale.",
					resolvedRoot, base)
			}
		}
		if truncated {
			startResult := offset + 1
			endResult := offset + len(jsonResults)
			response.Message = fmt.Sprintf(
				"Showing results %d\u2013%d of %d. Pass offset=%d for the next page.",
				startResult, endResult, totalMatches, nextOffset,
			)
		} else if totalMatches == 0 {
			// Explain the most common cause of a surprising empty result: a
			// multi-word query whose terms are all ANDed. Only fires for pure
			// AND-of-keywords queries (no OR), analysed on the caller's original
			// query rather than the path/file-composed one.
			if kws := andedKeywordsForHint(userQuery); len(kws) > 0 {
				response.Message = andHintMessage(kws)
			}
		}

		jsonBytes, err := json.Marshal(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
