// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestBuildJSONResultsEmpty(t *testing.T) {
	cfg := DefaultConfig()
	results := buildJSONResults(&cfg, nil)
	if results != nil {
		t.Errorf("expected nil for empty input, got %v", results)
	}
}

func TestBuildJSONResultsSnippetMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SnippetLength = 300

	fj := &common.FileJob{
		Filename: "hello.go",
		Location: "/tmp/hello.go",
		Content:  []byte("package main\n\nfunc hello() {\n\tprintln(\"hello world\")\n}\n"),
		Bytes:    55,
		MatchLocations: map[string][][]int{
			"hello": {{14, 19}, {35, 40}},
		},
		Language:   "Go",
		Lines:      5,
		Code:       3,
		Comment:    0,
		Blank:      1,
		Complexity: 0,
		Score:      1.5,
	}

	results := buildJSONResults(&cfg, []*common.FileJob{fj})
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	r := results[0]
	if r.Filename != "hello.go" {
		t.Errorf("expected filename hello.go, got %s", r.Filename)
	}
	if r.Location != "/tmp/hello.go" {
		t.Errorf("expected location /tmp/hello.go, got %s", r.Location)
	}
	if r.Language != "Go" {
		t.Errorf("expected language Go, got %s", r.Language)
	}
	if r.TotalLines != 5 {
		t.Errorf("expected total_lines 5, got %d", r.TotalLines)
	}
	if r.Code != 3 {
		t.Errorf("expected code 3, got %d", r.Code)
	}
}

func TestBuildJSONResultsLineMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SnippetMode = "lines"

	content := []byte("line one\nline two hello\nline three\n")
	fj := &common.FileJob{
		Filename: "test.txt",
		Location: "/tmp/test.txt",
		Content:  content,
		Bytes:    len(content),
		MatchLocations: map[string][][]int{
			"hello": {{18, 23}},
		},
		Score: 2.0,
	}

	results := buildJSONResults(&cfg, []*common.FileJob{fj})
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Lines == nil {
		t.Fatal("expected line results for lines mode")
	}
}

func TestMCPSearchHandlerMissingQuery(t *testing.T) {
	cfg := DefaultConfig()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	// No query argument
	req := mcp.CallToolRequest{}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing query")
	}
}

func TestMCPSearchHandlerEmptyQuery(t *testing.T) {
	cfg := DefaultConfig()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query": "   ",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for empty query")
	}
}

func TestMCPSearchHandlerReturnsJSON(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query":       "nonexistent_term_xyz",
		"max_results": float64(5),
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	// Should return valid JSON (empty array for no matches)
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if parsed.TotalMatches != 0 {
		t.Errorf("expected 0 total_matches for no results, got %d", parsed.TotalMatches)
	}
	if parsed.Truncated {
		t.Error("expected truncated=false for no results")
	}
}

func TestMCPSearchHandlerTruncation(t *testing.T) {
	dir := t.TempDir()
	// Create 30 files that all match the search term
	for i := 0; i < 30; i++ {
		content := fmt.Sprintf("package main\n\nfunc handler%d() {\n\t// unicorntoken\n}\n", i)
		fname := fmt.Sprintf("file%d.go", i)
		if err := os.WriteFile(filepath.Join(dir, fname), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := DefaultConfig()
	cfg.Directory = dir
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query":       "unicorntoken",
		"max_results": float64(5),
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if !parsed.Truncated {
		t.Error("expected truncated=true when results exceed max_results")
	}
	if parsed.TotalMatches != 30 {
		t.Errorf("expected total_matches=30, got %d", parsed.TotalMatches)
	}
	if parsed.ResultsReturned != 5 {
		t.Errorf("expected results_returned=5, got %d", parsed.ResultsReturned)
	}
	if len(parsed.Results) != 5 {
		t.Errorf("expected 5 results, got %d", len(parsed.Results))
	}
	if parsed.Message == "" {
		t.Error("expected non-empty message when truncated")
	}
	if !strings.Contains(parsed.Message, "30") {
		t.Errorf("expected message to contain total count '30', got: %s", parsed.Message)
	}
}

func TestMCPSearchHandlerNoTruncation(t *testing.T) {
	dir := t.TempDir()
	// Create 3 files — fewer than default max_results of 20
	for i := 0; i < 3; i++ {
		content := fmt.Sprintf("package main\n\nfunc handler%d() {\n\t// zebratoken\n}\n", i)
		fname := fmt.Sprintf("file%d.go", i)
		if err := os.WriteFile(filepath.Join(dir, fname), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := DefaultConfig()
	cfg.Directory = dir
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query": "zebratoken",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if parsed.Truncated {
		t.Error("expected truncated=false when all results fit")
	}
	if parsed.TotalMatches != 3 {
		t.Errorf("expected total_matches=3, got %d", parsed.TotalMatches)
	}
	if parsed.ResultsReturned != 3 {
		t.Errorf("expected results_returned=3, got %d", parsed.ResultsReturned)
	}
	if parsed.Message != "" {
		t.Errorf("expected empty message when not truncated, got: %s", parsed.Message)
	}
}

func TestMCPSearchHandlerRejectsUnknownParam(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	// Reproduces the reported bug: caller passes non-existent top-level
	// "path" and "ext" keys. These must be rejected, not silently dropped.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query": "classic migration backup bootstrap",
		"path":  "projects/desktop-raycast/packages/node-backend/src",
		"ext":   "ts",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for unknown parameter")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "ext") {
		t.Errorf("expected error to name the unknown 'ext' parameter, got: %s", text)
	}
	if !strings.Contains(text, "include_ext") {
		t.Errorf("expected error to point to 'include_ext', got: %s", text)
	}
}

func TestMCPSearchHandlerAcceptsKnownParams(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	// "path_filter" is a valid top-level param — this must not be rejected.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query":       "anything",
		"path_filter": "src",
		"file":        "*.go",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result for valid params: %v", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestMCPSearchHandlerTopLevelPathScopesGlobally(t *testing.T) {
	dir := t.TempDir()
	// Two dirs. Only "wanted" should survive a path filter, even though the
	// query uses OR (which would leak "other" matches via in-query precedence).
	for _, sub := range []string{"wanted", "other"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
		content := "package main\n\nfunc f() {\n\t// alphatoken betatoken\n}\n"
		if err := os.WriteFile(filepath.Join(dir, sub, "f.go"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := DefaultConfig()
	cfg.Directory = dir
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query":       "alphatoken OR betatoken",
		"path_filter": "wanted",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result.Content[0].(mcp.TextContent).Text)
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if parsed.TotalMatches != 1 {
		t.Fatalf("expected exactly 1 match under 'wanted', got %d", parsed.TotalMatches)
	}
	if !strings.Contains(parsed.Results[0].Location, filepath.Join("wanted", "f.go")) {
		t.Errorf("expected match in wanted/f.go, got %s", parsed.Results[0].Location)
	}
}

func TestComposeSearchQuery(t *testing.T) {
	cases := []struct {
		name             string
		query, path, fil string
		want             string
	}{
		{"none", "foo OR bar", "", "", "foo OR bar"},
		{"path", "foo OR bar", "src", "", "(foo OR bar) path:src"},
		{"file", "foo", "", "*.go", "(foo) file:*.go"},
		{"both", "foo", "src", "*.go", "(foo) path:src file:*.go"},
		{"multi-path ORed", "foo", "src,internal", "", "(foo) (path:src OR path:internal)"},
		{"trims and skips empties", "foo", " src , ", "", "(foo) path:src"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := composeSearchQuery(tc.query, tc.path, tc.fil); got != tc.want {
				t.Errorf("composeSearchQuery(%q, %q, %q) = %q, want %q", tc.query, tc.path, tc.fil, got, tc.want)
			}
		})
	}
}

func TestAndedKeywordsForHint(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"multi keyword AND", "database init startup sequence", []string{"database", "init", "startup", "sequence"}},
		{"single keyword", "database", nil},
		{"has OR", "database OR init OR startup", nil},
		{"grouped OR", "(database OR init) startup", nil},
		{"phrase not counted", `"database init"`, nil},
		{"regex not counted", `/func\s+init/`, nil},
		{"negated excluded", "database NOT init", nil},
		{"keywords plus filter still hints", "database init path:src", []string{"database", "init"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := andedKeywordsForHint(tc.query)
			if len(got) != len(tc.want) {
				t.Fatalf("andedKeywordsForHint(%q) = %v, want %v", tc.query, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("andedKeywordsForHint(%q) = %v, want %v", tc.query, got, tc.want)
				}
			}
		})
	}
}

func TestMCPSearchHandlerEmptyResultAndHint(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir() // empty dir → guaranteed zero matches
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query": "performStartupStep database init startup sequence",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result.Content[0].(mcp.TextContent).Text)
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if parsed.TotalMatches != 0 {
		t.Fatalf("expected 0 matches, got %d", parsed.TotalMatches)
	}
	if !strings.Contains(parsed.Message, "ANDed") {
		t.Errorf("expected an AND-explanation hint, got: %q", parsed.Message)
	}
	if !strings.Contains(parsed.Message, "OR") {
		t.Errorf("expected the hint to suggest OR, got: %q", parsed.Message)
	}
}

func TestMCPSearchHandlerEmptyResultNoHintForSingleTerm(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	cache := NewSearchCache()
	handler := mcpSearchHandler(&cfg, cache)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query": "loneterm",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if parsed.Message != "" {
		t.Errorf("expected no hint for a single-term query, got: %q", parsed.Message)
	}
}

func TestMCPGetFileHandlerMissingPath(t *testing.T) {
	cfg := DefaultConfig()
	handler := mcpGetFileHandler(&cfg)

	req := mcp.CallToolRequest{}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing path")
	}
}

func TestMCPGetFileHandlerEmptyPath(t *testing.T) {
	cfg := DefaultConfig()
	handler := mcpGetFileHandler(&cfg)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "   ",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for empty path")
	}
}

func TestMCPGetFileHandlerFileNotFound(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	handler := mcpGetFileHandler(&cfg)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "nonexistent.txt",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for nonexistent file")
	}
}

func TestMCPGetFileHandlerReadsFile(t *testing.T) {
	initLanguageDatabase()
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	content := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "main.go",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var fr mcpFileResult
	if err := json.Unmarshal([]byte(text.Text), &fr); err != nil {
		t.Fatalf("expected JSON response, got error: %v\ntext: %s", err, text.Text)
	}
	if fr.Language != "Go" {
		t.Errorf("expected language Go, got %s", fr.Language)
	}
	if fr.Lines <= 0 {
		t.Errorf("expected lines > 0, got %d", fr.Lines)
	}
	if fr.Code <= 0 {
		t.Errorf("expected code > 0, got %d", fr.Code)
	}
	if !strings.Contains(fr.Content, "1\tpackage main") {
		t.Errorf("expected line-numbered output in content, got: %s", fr.Content)
	}
}

func TestMCPGetFileHandlerNoLanguageHeader(t *testing.T) {
	initLanguageDatabase()
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	content := "line one\nline two\nline three\n"
	if err := os.WriteFile(filepath.Join(dir, "test.zzzzz"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "test.zzzzz",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var fr mcpFileResult
	if err := json.Unmarshal([]byte(text.Text), &fr); err != nil {
		t.Fatalf("expected JSON response, got error: %v\ntext: %s", err, text.Text)
	}
	if fr.Language != "" {
		t.Errorf("expected empty language for unknown extension, got %s", fr.Language)
	}
	if !strings.Contains(fr.Content, "1\tline one") {
		t.Errorf("expected line-numbered output in content, got: %s", fr.Content)
	}
}

func TestMCPGetFileHandlerLineRange(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	content := "alpha\nbeta\ngamma\ndelta\nepsilon\n"
	if err := os.WriteFile(filepath.Join(dir, "range.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path":       "range.txt",
		"start_line": float64(2),
		"end_line":   float64(4),
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %v", result)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var fr mcpFileResult
	if err := json.Unmarshal([]byte(text), &fr); err != nil {
		t.Fatalf("expected JSON response, got error: %v\ntext: %s", err, text)
	}
	if !strings.Contains(fr.Content, "2\tbeta") {
		t.Errorf("expected line 2 beta, got: %s", fr.Content)
	}
	if !strings.Contains(fr.Content, "4\tdelta") {
		t.Errorf("expected line 4 delta, got: %s", fr.Content)
	}
	if strings.Contains(fr.Content, "1\talpha") {
		t.Errorf("should not contain line 1, got: %s", fr.Content)
	}
	if strings.Contains(fr.Content, "5\tepsilon") {
		t.Errorf("should not contain line 5, got: %s", fr.Content)
	}
}

func TestMCPGetFileHandlerPathTraversalLocked(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	cfg.MCPLockDir = true
	handler := mcpGetFileHandler(&cfg)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "../../../etc/passwd",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for path traversal under --mcp-lock-dir")
	}
	if text := result.Content[0].(mcp.TextContent).Text; !strings.Contains(text, "mcp-lock-dir") {
		t.Errorf("expected error to mention --mcp-lock-dir, got: %s", text)
	}
}

// Without --mcp-lock-dir, get_file must be able to read files outside the
// default directory — search can now return locations from anywhere.
func TestMCPGetFileHandlerOutsideDefaultRootAllowed(t *testing.T) {
	outside := t.TempDir()
	target := filepath.Join(outside, "elsewhere.go")
	if err := os.WriteFile(target, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
	handler := mcpGetFileHandler(&cfg)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"path": target}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected outside path to be readable, got: %s", result.Content[0].(mcp.TextContent).Text)
	}
	var fr mcpFileResult
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &fr); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if !strings.Contains(fr.Content, "package main") {
		t.Errorf("expected file content, got: %s", fr.Content)
	}
}

func TestMCPGetFileHandlerBinaryFile(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	// Write a file with NUL bytes
	binaryContent := []byte("hello\x00world")
	if err := os.WriteFile(filepath.Join(dir, "binary.bin"), binaryContent, 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "binary.bin",
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for binary file")
	}
}

func TestMCPGetFileHandlerStartLineExceedsLength(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	if err := os.WriteFile(filepath.Join(dir, "short.txt"), []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path":       "short.txt",
		"start_line": float64(100),
	}
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for start_line exceeding file length")
	}
}

// writeGoFile creates dir/name containing a unique token, for root-scoping tests.
func writeGoFile(t *testing.T, dir, name, token string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	content := fmt.Sprintf("package main\n\nfunc f() {\n\tprintln(%q)\n}\n", token)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// searchWith runs the search handler and returns the decoded response, failing
// the test if the handler reported an error.
func searchWith(t *testing.T, cfg *Config, args map[string]any) mcpSearchResponse {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := mcpSearchHandler(cfg, NewSearchCache())(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Content[0].(mcp.TextContent).Text)
	}
	var parsed mcpSearchResponse
	if err := json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	return parsed
}

// searchErrorText runs the search handler expecting an error result.
func searchErrorText(t *testing.T, cfg *Config, args map[string]any) string {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := mcpSearchHandler(cfg, NewSearchCache())(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result, got: %s", result.Content[0].(mcp.TextContent).Text)
	}
	return result.Content[0].(mcp.TextContent).Text
}

func TestMCPSearchPathSelectsRoot(t *testing.T) {
	defaultRoot := t.TempDir()
	otherRoot := t.TempDir()
	writeGoFile(t, defaultRoot, "here.go", "sharedtoken")
	writeGoFile(t, otherRoot, "there.go", "sharedtoken")

	cfg := DefaultConfig()
	cfg.Directory = defaultRoot

	got := searchWith(t, &cfg, map[string]any{"query": "sharedtoken"})
	if got.TotalMatches != 1 || !strings.HasSuffix(got.Results[0].Location, "here.go") {
		t.Fatalf("default root should match here.go only, got %d: %+v", got.TotalMatches, got.Results)
	}
	if got.SearchedDirectory != defaultRoot {
		t.Errorf("expected searched_directory %s, got %s", defaultRoot, got.SearchedDirectory)
	}

	got = searchWith(t, &cfg, map[string]any{"query": "sharedtoken", "path": otherRoot})
	if got.TotalMatches != 1 || !strings.HasSuffix(got.Results[0].Location, "there.go") {
		t.Fatalf("explicit root should match there.go only, got %d: %+v", got.TotalMatches, got.Results)
	}
	if got.SearchedDirectory != otherRoot {
		t.Errorf("expected searched_directory %s, got %s", otherRoot, got.SearchedDirectory)
	}
}

func TestMCPSearchRelativePathResolvesAgainstDefaultRoot(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, filepath.Join(root, "wanted"), "a.go", "reltoken")
	writeGoFile(t, filepath.Join(root, "other"), "b.go", "reltoken")

	cfg := DefaultConfig()
	cfg.Directory = root

	got := searchWith(t, &cfg, map[string]any{"query": "reltoken", "path": "wanted"})
	if got.TotalMatches != 1 || !strings.HasSuffix(got.Results[0].Location, filepath.Join("wanted", "a.go")) {
		t.Fatalf("expected only wanted/a.go, got %d: %+v", got.TotalMatches, got.Results)
	}
	if want := filepath.Join(root, "wanted"); got.SearchedDirectory != want {
		t.Errorf("expected searched_directory %s, got %s", want, got.SearchedDirectory)
	}
}

func TestMCPSearchPathRejectsBadRoots(t *testing.T) {
	root := t.TempDir()
	file := writeGoFile(t, root, "single.go", "token")

	cfg := DefaultConfig()
	cfg.Directory = root

	cases := []struct{ name, path, wantSubstring string }{
		{"glob", "*/pkg/*", "path_filter"},
		{"missing", "no-such-dir", "does not exist"},
		{"file", file, "get_file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := searchErrorText(t, &cfg, map[string]any{"query": "token", "path": tc.path})
			if !strings.Contains(text, tc.wantSubstring) {
				t.Errorf("expected error mentioning %q, got: %s", tc.wantSubstring, text)
			}
		})
	}
}

func TestMCPSearchLockDirRejectsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeGoFile(t, outside, "there.go", "locktoken")

	cfg := DefaultConfig()
	cfg.Directory = root
	cfg.MCPLockDir = true

	text := searchErrorText(t, &cfg, map[string]any{"query": "locktoken", "path": outside})
	if !strings.Contains(text, "mcp-lock-dir") {
		t.Errorf("expected error to mention --mcp-lock-dir, got: %s", text)
	}
}

func TestMCPSearchGitSyncNoteOnlyOutsideSyncRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeGoFile(t, filepath.Join(root, "sub"), "a.go", "notetoken")
	writeGoFile(t, outside, "b.go", "notetoken")

	cfg := DefaultConfig()
	cfg.Directory = root
	cfg.GitSync = true

	if got := searchWith(t, &cfg, map[string]any{"query": "notetoken", "path": "sub"}); got.GitSyncNote != "" {
		t.Errorf("expected no git-sync note for a subdirectory of the sync root, got: %s", got.GitSyncNote)
	}
	got := searchWith(t, &cfg, map[string]any{"query": "notetoken", "path": outside})
	if !strings.Contains(got.GitSyncNote, "may be stale") {
		t.Errorf("expected a git-sync staleness note, got: %q", got.GitSyncNote)
	}
}
