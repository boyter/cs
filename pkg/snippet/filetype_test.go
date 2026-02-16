// SPDX-License-Identifier: MIT OR Unlicense

package snippet

import "testing"

func TestSnippetModeForExtensionCode(t *testing.T) {
	codeExts := []string{"go", "py", "js", "ts", "rs", "c", "cpp", "java", "rb", "sh"}
	for _, ext := range codeExts {
		if got := SnippetModeForExtension(ext); got != "lines" {
			t.Errorf("SnippetModeForExtension(%q) = %q, want \"lines\"", ext, got)
		}
	}
}

func TestSnippetModeForExtensionProse(t *testing.T) {
	proseExts := []string{"md", "markdown", "txt", "text", "rst", "adoc", "asciidoc",
		"html", "htm", "xml", "svg", "csv", "tsv", "log", "org",
		"tex", "latex", "wiki", "rdoc", "textile", "pod", "man", "roff"}
	for _, ext := range proseExts {
		if got := SnippetModeForExtension(ext); got != "snippet" {
			t.Errorf("SnippetModeForExtension(%q) = %q, want \"snippet\"", ext, got)
		}
	}
}

func TestSnippetModeForExtensionUnknown(t *testing.T) {
	unknownExts := []string{"xyz", "foo", "zzz"}
	for _, ext := range unknownExts {
		if got := SnippetModeForExtension(ext); got != "lines" {
			t.Errorf("SnippetModeForExtension(%q) = %q, want \"lines\"", ext, got)
		}
	}
}

func TestSnippetModeForExtensionEmpty(t *testing.T) {
	if got := SnippetModeForExtension(""); got != "lines" {
		t.Errorf("SnippetModeForExtension(\"\") = %q, want \"lines\"", got)
	}
}
