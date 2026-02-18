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

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/ranker"
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
		mcp.WithDescription("Search code files recursively using boolean queries, regex, and fuzzy matching with relevance ranking. "+
			"Supports: exact match with quotes, fuzzy match (term~1, term~2), NOT operator, regex (/pattern/), "+
			"file filtering (file:test, filename:.go), path filtering (path:pkg/search)."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query",
			mcp.Description("The search query. Supports boolean logic (AND/OR/NOT), quoted phrases, regex (/pattern/), fuzzy matching (term~1), file filtering (file:name), and path filtering (path:dir)."),
			mcp.Required(),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return. Defaults to 20."),
		),
		mcp.WithNumber("snippet_length",
			mcp.Description("Size of the code snippet to display in characters."),
		),
		mcp.WithBoolean("case_sensitive",
			mcp.Description("Make the search case sensitive."),
		),
		mcp.WithString("include_ext",
			mcp.Description("Comma-separated list of file extensions to search (e.g. \"go,js,py\")."),
		),
		mcp.WithString("language",
			mcp.Description("Comma-separated list of language types to search (e.g. \"Go,Python,JavaScript\")."),
		),
		mcp.WithString("gravity",
			mcp.Description("Complexity gravity intent: brain (2.5), logic (1.5), default (1.0), low (0.2), off (0.0). Controls how much cyclomatic complexity boosts ranking."),
		),
	)

	mcpServer.AddTool(searchTool, mcpSearchHandler(cfg, cache))

	getFileTool := mcp.NewTool("get_file",
		mcp.WithDescription("Read the contents of a file within the project directory. Returns JSON with 'content' (line-numbered file text) and, for recognised source files, 'language', 'lines', 'code', 'comment', 'blank', 'complexity' fields."),
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

		// Run search
		ch, stats := DoSearch(ctx, &searchCfg, query, cache)

		var results []*common.FileJob
		for fj := range ch {
			results = append(results, fj)
		}

		// Rank results
		textFileCount := int(stats.TextFileCount.Load())
		results = ranker.RankResults(searchCfg.Ranker, textFileCount, results, searchCfg.StructuralRankerConfig(), searchCfg.ResolveGravityStrength())

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
