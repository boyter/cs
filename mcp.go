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
	"strings"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/cs/v3/pkg/ranker"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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
			"- Regex: '/pattern/' (e.g. '/func\\s+Test/')\n"+
			"- Fuzzy: 'term~1' or 'term~2' for typo-tolerant matching (Levenshtein distance 1 or 2)\n\n"+
			"Filters (in-query):\n"+
			"- file:pattern — match filename (substring or glob: file:*.go, file:*_test.go)\n"+
			"- path:pattern — match full path (substring or glob: path:*/pkg/*, NOT path:vendor/*/*)\n"+
			"- lang:value — filter by language: lang:go, lang=go,python (multi-value with commas)\n"+
			"- ext:value — filter by extension: ext:go, ext=ts,tsx\n\n"+
			"Filter operators: = != (e.g. lang!=python, file!=test)\n"+
			"Negation: NOT file:test, file!=test, NOT path:vendor, path!=vendor\n\n"+
			"Content type filter (code_filter parameter):\n"+
			"- 'only-code': matches in code only, skipping comments and strings — e.g. find where a function is called, not just mentioned\n"+
			"- 'only-strings': matches in string literals only — find SQL queries, error messages, config values, connection strings\n"+
			"- 'only-comments': matches in comments only — find TODOs, developer explanations, annotations\n"+
			"- 'only-declarations': matches only on declaration lines (func, type, class, def, struct, etc.) — find where something is defined\n"+
			"- 'only-usages': matches only on non-declaration lines — find where something is called/referenced (impact analysis)\n\n"+
			"Combined examples:\n"+
			"- 'jwt middleware lang:go NOT path:vendor' — find Go JWT middleware outside vendor\n"+
			"- query='dense_rank' code_filter='only-strings' — find the actual SQL string, not code references\n"+
			"- query='middleware' code_filter='only-code' path filter='NOT path:vendor' — find middleware implementations\n"+
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
			"- NOT binds to the next term only, not the whole query. 'a OR b NOT path:vendor' means 'a OR (b AND NOT path:vendor)'. To exclude globally, use grouping: '(a OR b) NOT path:vendor'. Precedence: NOT (tightest) > AND > OR (loosest).\n"+
			"- max_results defaults to 20. Set higher (e.g. 100) for broad discovery or exploring unfamiliar code.\n\n"+
			"Workflow tips:\n"+
			"- Searching for a specific term, identifier, or function name → use snippet_mode='grep' with context=5-10. This gives every occurrence with surrounding code in one call.\n"+
			"- Conceptual or discovery queries ('how does auth work', 'what handles errors') → use the default auto mode. The ranker surfaces the most relevant files.\n"+
			"- Once a specific file is identified, switch to get_file to read it — don't keep searching the same file."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query",
			mcp.Description("The search query. Terms are ANDed by default. Supports: OR ('error OR exception'), NOT ('NOT vendor'), "+
				"grouping ('(auth OR login) AND handler'), quoted phrases ('\"exact match\"'), regex (/pattern/), fuzzy (term~1, term~2). "+
				"In-query filters: file:name, path:dir, lang:go, ext:ts. Operators: = != (lang!=python, file!=test). "+
				"Multi-value: lang=go,python, ext=ts,tsx. 'file:' matches filename only; 'path:' matches the full directory path. "+
				"Query limits: max 250 characters and 12 unique search terms."),
			mcp.Required(),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return. Defaults to 20. No upper limit enforced. Use higher values (50-100) for broad discovery queries or when exploring unfamiliar codebases."),
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
		mcp.WithString("gravity",
			mcp.Description("Complexity gravity intent controlling how much cyclomatic complexity boosts ranking. "+
				"Values: brain (2.5) — find complex algorithmic code, logic (1.5) — prefer branching/control flow, "+
				"default (1.0) — balanced, low (0.2) — mostly ignore complexity, off (0.0) — pure text relevance only."),
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
				"IMPORTANT: When using code_filter, always also set the 'language' parameter to scope results to the relevant language(s). Without it, results from all languages in the project (including dependency directories like node_modules, vendor, site-packages) will dominate.\n"+
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

		// Resolve path relative to project directory
		resolved := path
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(cfg.Directory, resolved)
		}

		// Security: ensure resolved path is within the project directory
		absProject, err := filepath.Abs(cfg.Directory)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve project directory: %v", err)), nil
		}
		absResolved, err := filepath.Abs(resolved)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve file path: %v", err)), nil
		}
		if !strings.HasPrefix(absResolved, absProject+string(filepath.Separator)) && absResolved != absProject {
			return mcp.NewToolResultError("path is outside the project directory"), nil
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

		// Copy config so we can override per-request without mutating the shared config
		searchCfg := *cfg
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
		results = ranker.RankResults(searchCfg.Ranker, textFileCount, results, searchCfg.StructuralRankerConfig(), searchCfg.ResolveGravityStrength(), searchCfg.ResolveNoiseSensitivity(), searchCfg.TestPenalty, testIntent)

		// Dedup (before limit, so freed slots get backfilled)
		if v, ok := request.GetArguments()["dedup"]; ok {
			if b, ok := v.(bool); ok && b {
				results = ranker.DeduplicateResults(results)
			}
		}

		// Apply max_results limit
		if maxResults > 0 && len(results) > maxResults {
			results = results[:maxResults]
		}

		// Build JSON using the shared helper
		jsonResults := buildJSONResults(&searchCfg, results)
		jsonBytes, err := json.Marshal(jsonResults)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
