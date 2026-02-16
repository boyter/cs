// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSearchTemplate_AllStyles(t *testing.T) {
	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			cfg := &Config{TemplateStyle: style}
			tmpl, err := resolveSearchTemplate(cfg)
			if err != nil {
				t.Fatalf("resolveSearchTemplate(%q) error: %v", style, err)
			}
			if tmpl == nil {
				t.Fatalf("resolveSearchTemplate(%q) returned nil", style)
			}
		})
	}
}

func TestResolveDisplayTemplate_AllStyles(t *testing.T) {
	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			cfg := &Config{TemplateStyle: style}
			tmpl, err := resolveDisplayTemplate(cfg)
			if err != nil {
				t.Fatalf("resolveDisplayTemplate(%q) error: %v", style, err)
			}
			if tmpl == nil {
				t.Fatalf("resolveDisplayTemplate(%q) returned nil", style)
			}
		})
	}
}

func TestResolveSearchTemplate_InvalidStyle(t *testing.T) {
	cfg := &Config{TemplateStyle: "neon"}
	_, err := resolveSearchTemplate(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style, got nil")
	}
}

func TestResolveDisplayTemplate_InvalidStyle(t *testing.T) {
	cfg := &Config{TemplateStyle: "neon"}
	_, err := resolveDisplayTemplate(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style, got nil")
	}
}

func TestSearchTemplateExecution(t *testing.T) {
	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			cfg := &Config{TemplateStyle: style}
			tmpl, err := resolveSearchTemplate(cfg)
			if err != nil {
				t.Fatalf("resolve error: %v", err)
			}

			data := httpSearch{
				SearchTerm:          "test query",
				SnippetSize:         300,
				ResultsCount:        1,
				RuntimeMilliseconds: 42,
				ProcessedFileCount:  100,
				Results: []httpSearchResult{
					{
						Title:    "main.go",
						Location: "main.go",
						Content:  []template.HTML{"func <strong>main</strong>() {}"},
						StartPos: 0,
						EndPos:   10,
						Score:    1.5,
					},
				},
				ExtensionFacet: []httpFacetResult{
					{Title: "go", Count: 5, SearchTerm: "test query", SnippetSize: 300},
				},
				Pages: []httpPageResult{
					{SearchTerm: "test query", SnippetSize: 300, Value: 0, Name: "1"},
				},
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("execute error: %v", err)
			}
			if buf.Len() == 0 {
				t.Fatal("template produced empty output")
			}
		})
	}
}

func TestDisplayTemplateExecution(t *testing.T) {
	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			cfg := &Config{TemplateStyle: style}
			tmpl, err := resolveDisplayTemplate(cfg)
			if err != nil {
				t.Fatalf("resolve error: %v", err)
			}

			data := httpFileDisplay{
				Location:            "main.go",
				Content:             template.HTML("package main\n\nfunc <strong>main</strong>() {}"),
				RuntimeMilliseconds: 7,
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("execute error: %v", err)
			}
			if buf.Len() == 0 {
				t.Fatal("template produced empty output")
			}
		})
	}
}

func TestCustomTemplateOverride(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal custom search template
	searchPath := filepath.Join(dir, "custom_search.html")
	if err := os.WriteFile(searchPath, []byte(`<p>{{ .SearchTerm }}</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a minimal custom display template
	displayPath := filepath.Join(dir, "custom_display.html")
	if err := os.WriteFile(displayPath, []byte(`<p>{{ .Location }}</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("search override", func(t *testing.T) {
		cfg := &Config{
			TemplateStyle:  "dark",
			SearchTemplate: searchPath,
		}
		tmpl, err := resolveSearchTemplate(cfg)
		if err != nil {
			t.Fatalf("resolve error: %v", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, httpSearch{SearchTerm: "hello"}); err != nil {
			t.Fatalf("execute error: %v", err)
		}
		if got := buf.String(); got != "<p>hello</p>" {
			t.Fatalf("expected custom output, got %q", got)
		}
	})

	t.Run("display override", func(t *testing.T) {
		cfg := &Config{
			TemplateStyle:   "dark",
			DisplayTemplate: displayPath,
		}
		tmpl, err := resolveDisplayTemplate(cfg)
		if err != nil {
			t.Fatalf("resolve error: %v", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, httpFileDisplay{Location: "/tmp/test.go"}); err != nil {
			t.Fatalf("execute error: %v", err)
		}
		if got := buf.String(); got != "<p>/tmp/test.go</p>" {
			t.Fatalf("expected custom output, got %q", got)
		}
	})
}

func TestSearchTemplateExecution_LineMode(t *testing.T) {
	for _, style := range validStyles {
		t.Run(style, func(t *testing.T) {
			cfg := &Config{TemplateStyle: style}
			tmpl, err := resolveSearchTemplate(cfg)
			if err != nil {
				t.Fatalf("resolve error: %v", err)
			}

			data := httpSearch{
				SearchTerm:          "test query",
				SnippetSize:         300,
				ResultsCount:        1,
				RuntimeMilliseconds: 42,
				ProcessedFileCount:  100,
				Results: []httpSearchResult{
					{
						Title:      "main.go",
						Location:   "main.go",
						Score:      1.5,
						IsLineMode: true,
						LineResults: []httpLineResult{
							{LineNumber: 10, Content: "func <strong>main</strong>() {"},
							{LineNumber: 11, Content: "    fmt.Println()"},
							{LineNumber: 12, Content: "}"},
						},
					},
				},
				ExtensionFacet: []httpFacetResult{
					{Title: "go", Count: 5, SearchTerm: "test query", SnippetSize: 300},
				},
				Pages: []httpPageResult{
					{SearchTerm: "test query", SnippetSize: 300, Value: 0, Name: "1"},
				},
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("execute error: %v", err)
			}
			if buf.Len() == 0 {
				t.Fatal("template produced empty output")
			}
		})
	}
}

func TestIsValidStyle(t *testing.T) {
	tests := []struct {
		style string
		want  bool
	}{
		{"dark", true},
		{"light", true},
		{"bare", true},
		{"neon", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isValidStyle(tt.style); got != tt.want {
			t.Errorf("isValidStyle(%q) = %v, want %v", tt.style, got, tt.want)
		}
	}
}
