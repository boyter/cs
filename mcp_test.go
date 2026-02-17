// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boyter/cs/pkg/common"
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
	var parsed []jsonResult
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
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
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Directory = dir

	content := "line one\nline two\nline three\n"
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	handler := mcpGetFileHandler(&cfg)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"path": "test.txt",
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
	if !strings.Contains(text.Text, "1\tline one") {
		t.Errorf("expected line-numbered output, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "2\tline two") {
		t.Errorf("expected line 2 in output, got: %s", text.Text)
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
	if !strings.Contains(text, "2\tbeta") {
		t.Errorf("expected line 2 beta, got: %s", text)
	}
	if !strings.Contains(text, "4\tdelta") {
		t.Errorf("expected line 4 delta, got: %s", text)
	}
	if strings.Contains(text, "1\talpha") {
		t.Errorf("should not contain line 1, got: %s", text)
	}
	if strings.Contains(text, "5\tepsilon") {
		t.Errorf("should not contain line 5, got: %s", text)
	}
}

func TestMCPGetFileHandlerPathTraversal(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Directory = t.TempDir()
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
		t.Error("expected error result for path traversal")
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
