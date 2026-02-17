// SPDX-License-Identifier: MIT

package main

import (
	"sort"
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	initLanguageDatabase()

	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "Go"},
		{"App.java", "Java"},
		{"script.py", "Python"},
		{"index.js", "JavaScript"},
		{"style.css", "CSS"},
		{"page.html", "HTML"},
		{"Makefile", "Makefile"},
		{"unknown.zzzzz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectLanguage(tt.filename, nil)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFileCodeStats(t *testing.T) {
	initLanguageDatabase()

	content := []byte("package main\n\n// main entry point\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	lang, lines, code, comment, blank, complexity := fileCodeStats("main.go", content)
	if lang != "Go" {
		t.Errorf("expected language Go, got %q", lang)
	}
	if lines == 0 {
		t.Error("expected non-zero lines")
	}
	if code == 0 {
		t.Error("expected non-zero code")
	}
	if comment == 0 {
		t.Error("expected non-zero comment")
	}
	if blank == 0 {
		t.Error("expected non-zero blank")
	}
	_ = complexity // complexity may be zero for simple code
}

func TestFileCodeStatsUnknown(t *testing.T) {
	initLanguageDatabase()

	content := []byte("some random text\n")
	lang, lines, code, comment, blank, complexity := fileCodeStats("unknown.zzzzz", content)
	if lang != "" {
		t.Errorf("expected empty language, got %q", lang)
	}
	if lines != 0 || code != 0 || comment != 0 || blank != 0 || complexity != 0 {
		t.Errorf("expected all zero stats, got lines=%d code=%d comment=%d blank=%d complexity=%d",
			lines, code, comment, blank, complexity)
	}
}

func TestLanguageExtensions(t *testing.T) {
	initLanguageDatabase()

	t.Run("single language", func(t *testing.T) {
		exts := languageExtensions([]string{"Go"})
		if len(exts) == 0 {
			t.Fatal("expected at least one extension for Go")
		}
		found := false
		for _, e := range exts {
			if e == "go" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'go' in extensions, got %v", exts)
		}
	})

	t.Run("multiple languages", func(t *testing.T) {
		exts := languageExtensions([]string{"Go", "Java"})
		hasGo, hasJava := false, false
		for _, e := range exts {
			if e == "go" {
				hasGo = true
			}
			if e == "java" {
				hasJava = true
			}
		}
		if !hasGo || !hasJava {
			t.Errorf("expected both 'go' and 'java' in extensions, got %v", exts)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		lower := languageExtensions([]string{"go"})
		upper := languageExtensions([]string{"Go"})
		sort.Strings(lower)
		sort.Strings(upper)
		if strings.Join(lower, ",") != strings.Join(upper, ",") {
			t.Errorf("case sensitivity mismatch: %v vs %v", lower, upper)
		}
	})

	t.Run("unknown language", func(t *testing.T) {
		exts := languageExtensions([]string{"NotARealLanguage12345"})
		if len(exts) != 0 {
			t.Errorf("expected no extensions for unknown language, got %v", exts)
		}
	})
}
