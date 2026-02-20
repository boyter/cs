// SPDX-License-Identifier: MIT

package ranker

import (
	"bytes"
)

// DeclarationPattern represents a line-start pattern that indicates a declaration.
type DeclarationPattern struct {
	Prefix []byte // bytes that the trimmed line must start with
}

// languageDeclarationPatterns maps scc language names to their declaration patterns.
// Each pattern matches the beginning of a trimmed (leading-whitespace-removed) line.
var languageDeclarationPatterns = map[string][]DeclarationPattern{
	"Go": {
		{Prefix: []byte("func ")},
		{Prefix: []byte("func(")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("var(")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("const(")},
	},
	"Python": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("async def ")},
	},
	"JavaScript": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("function(")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("let ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("export function ")},
		{Prefix: []byte("export default function")},
		{Prefix: []byte("export class ")},
		{Prefix: []byte("export const ")},
		{Prefix: []byte("export let ")},
		{Prefix: []byte("export default class")},
	},
	"TypeScript": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("function(")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("let ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("export function ")},
		{Prefix: []byte("export default function")},
		{Prefix: []byte("export class ")},
		{Prefix: []byte("export const ")},
		{Prefix: []byte("export let ")},
		{Prefix: []byte("export default class")},
		{Prefix: []byte("export interface ")},
		{Prefix: []byte("export type ")},
		{Prefix: []byte("export enum ")},
	},
	"TSX": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("function(")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("let ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("export function ")},
		{Prefix: []byte("export default function")},
		{Prefix: []byte("export class ")},
		{Prefix: []byte("export const ")},
		{Prefix: []byte("export let ")},
		{Prefix: []byte("export default class")},
		{Prefix: []byte("export interface ")},
		{Prefix: []byte("export type ")},
		{Prefix: []byte("export enum ")},
	},
	"Rust": {
		{Prefix: []byte("fn ")},
		{Prefix: []byte("pub fn ")},
		{Prefix: []byte("pub(crate) fn ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("pub struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("pub enum ")},
		{Prefix: []byte("trait ")},
		{Prefix: []byte("pub trait ")},
		{Prefix: []byte("impl ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("pub type ")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("pub const ")},
		{Prefix: []byte("static ")},
		{Prefix: []byte("pub static ")},
	},
	"Java": {
		{Prefix: []byte("public class ")},
		{Prefix: []byte("public interface ")},
		{Prefix: []byte("public enum ")},
		{Prefix: []byte("public abstract class ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("abstract class ")},
		{Prefix: []byte("private class ")},
		{Prefix: []byte("protected class ")},
	},
	"C": {
		{Prefix: []byte("#define ")},
		{Prefix: []byte("typedef ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("union ")},
	},
	"C++": {
		{Prefix: []byte("#define ")},
		{Prefix: []byte("typedef ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("union ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("namespace ")},
		{Prefix: []byte("template")},
	},
	"C#": {
		{Prefix: []byte("public class ")},
		{Prefix: []byte("public interface ")},
		{Prefix: []byte("public enum ")},
		{Prefix: []byte("public struct ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("private class ")},
		{Prefix: []byte("protected class ")},
		{Prefix: []byte("internal class ")},
	},
	"Ruby": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("module ")},
	},
	"PHP": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("trait ")},
		{Prefix: []byte("abstract class ")},
		{Prefix: []byte("public function ")},
		{Prefix: []byte("private function ")},
		{Prefix: []byte("protected function ")},
	},
	"Kotlin": {
		{Prefix: []byte("fun ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("data class ")},
		{Prefix: []byte("sealed class ")},
		{Prefix: []byte("object ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("enum class ")},
		{Prefix: []byte("typealias ")},
		{Prefix: []byte("val ")},
		{Prefix: []byte("var ")},
	},
	"Swift": {
		{Prefix: []byte("func ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("protocol ")},
		{Prefix: []byte("typealias ")},
		{Prefix: []byte("let ")},
		{Prefix: []byte("var ")},
	},
	"Shell": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("function(")},
	},
	"Lua": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("local function ")},
	},
	"Scala": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("val ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("trait ")},
		{Prefix: []byte("object ")},
		{Prefix: []byte("case class ")},
		{Prefix: []byte("case object ")},
		{Prefix: []byte("sealed trait ")},
		{Prefix: []byte("sealed class ")},
		{Prefix: []byte("abstract class ")},
		{Prefix: []byte("type ")},
	},
	"Elixir": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("defp ")},
		{Prefix: []byte("defmodule ")},
		{Prefix: []byte("defmacro ")},
		{Prefix: []byte("defmacrop ")},
		{Prefix: []byte("defstruct")},
		{Prefix: []byte("defprotocol ")},
		{Prefix: []byte("defimpl ")},
	},
	"Haskell": {
		{Prefix: []byte("data ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("newtype ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("instance ")},
		{Prefix: []byte("module ")},
	},
	"Perl": {
		{Prefix: []byte("sub ")},
		{Prefix: []byte("package ")},
		{Prefix: []byte("use constant ")},
	},
	"Zig": {
		{Prefix: []byte("fn ")},
		{Prefix: []byte("pub fn ")},
		{Prefix: []byte("const ")},
		{Prefix: []byte("pub const ")},
		{Prefix: []byte("var ")},
		{Prefix: []byte("pub var ")},
	},
	"Dart": {
		{Prefix: []byte("class ")},
		{Prefix: []byte("abstract class ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("mixin ")},
		{Prefix: []byte("extension ")},
		{Prefix: []byte("typedef ")},
	},
	"Julia": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("mutable struct ")},
		{Prefix: []byte("abstract type ")},
		{Prefix: []byte("macro ")},
		{Prefix: []byte("module ")},
	},
	"Clojure": {
		{Prefix: []byte("(defn ")},
		{Prefix: []byte("(def ")},
		{Prefix: []byte("(defmacro ")},
		{Prefix: []byte("(defprotocol ")},
		{Prefix: []byte("(defrecord ")},
		{Prefix: []byte("(deftype ")},
		{Prefix: []byte("(ns ")},
	},
	"Erlang": {
		{Prefix: []byte("-module(")},
		{Prefix: []byte("-export(")},
		{Prefix: []byte("-define(")},
		{Prefix: []byte("-record(")},
		{Prefix: []byte("-type(")},
		{Prefix: []byte("-type ")},
		{Prefix: []byte("-spec(")},
		{Prefix: []byte("-spec ")},
	},
	"Groovy": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("interface ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("trait ")},
	},
	"OCaml": {
		{Prefix: []byte("let ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("module ")},
		{Prefix: []byte("val ")},
		{Prefix: []byte("external ")},
	},
	"MATLAB": {
		{Prefix: []byte("function ")},
	},
	"Powershell": {
		{Prefix: []byte("function ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("enum ")},
	},
	"Nim": {
		{Prefix: []byte("proc ")},
		{Prefix: []byte("func ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("template ")},
		{Prefix: []byte("macro ")},
		{Prefix: []byte("method ")},
	},
	"Crystal": {
		{Prefix: []byte("def ")},
		{Prefix: []byte("class ")},
		{Prefix: []byte("module ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("lib ")},
		{Prefix: []byte("macro ")},
	},
	"V": {
		{Prefix: []byte("fn ")},
		{Prefix: []byte("pub fn ")},
		{Prefix: []byte("struct ")},
		{Prefix: []byte("pub struct ")},
		{Prefix: []byte("enum ")},
		{Prefix: []byte("type ")},
		{Prefix: []byte("const ")},
	},
}

// IsDeclarationLine checks if a line of code is a declaration based on
// language-specific heuristics. The line should have leading whitespace
// already trimmed.
func IsDeclarationLine(trimmedLine []byte, language string) bool {
	patterns, ok := languageDeclarationPatterns[language]
	if !ok {
		return false
	}

	for _, pat := range patterns {
		if bytes.HasPrefix(trimmedLine, pat.Prefix) {
			return true
		}
	}
	return false
}

// HasDeclarationPatterns returns true if the language has declaration
// patterns defined. Used to determine if filtering is possible.
func HasDeclarationPatterns(language string) bool {
	_, ok := languageDeclarationPatterns[language]
	return ok
}

// ClassifyMatchLocations classifies each match location as a declaration
// or usage based on the line it appears on. Returns two maps:
// declarations and usages, each containing the match locations that
// fall on declaration/usage lines respectively.
//
// If the language has no declaration patterns, all matches are returned
// as usages (conservative: we can't identify declarations without patterns).
func ClassifyMatchLocations(
	content []byte,
	matchLocations map[string][][]int,
	language string,
) (declarations, usages map[string][][]int) {
	declarations = make(map[string][][]int)
	usages = make(map[string][][]int)

	if len(content) == 0 || len(matchLocations) == 0 {
		return declarations, usages
	}

	// Pre-compute line boundaries for O(1) line lookup
	// lineStarts[i] = byte offset where line i begins (0-indexed)
	lineStarts := []int{0}
	for i, b := range content {
		if b == '\n' && i+1 < len(content) {
			lineStarts = append(lineStarts, i+1)
		}
	}

	// For each match location, find its line and classify
	for term, locs := range matchLocations {
		for _, loc := range locs {
			if len(loc) < 2 {
				continue
			}
			startByte := loc[0]
			if startByte < 0 || startByte >= len(content) {
				usages[term] = append(usages[term], loc)
				continue
			}

			// Binary search for the line containing startByte
			lineIdx := findLine(lineStarts, startByte)

			// Extract the line
			lineStart := lineStarts[lineIdx]
			lineEnd := len(content)
			if lineIdx+1 < len(lineStarts) {
				lineEnd = lineStarts[lineIdx+1] - 1 // -1 to exclude \n
			}
			if lineEnd > len(content) {
				lineEnd = len(content)
			}

			line := content[lineStart:lineEnd]
			trimmedLine := bytes.TrimLeft(line, " \t")

			if IsDeclarationLine(trimmedLine, language) {
				declarations[term] = append(declarations[term], loc)
			} else {
				usages[term] = append(usages[term], loc)
			}
		}
	}

	return declarations, usages
}

// findLine returns the 0-indexed line number for a byte offset using binary search.
func findLine(lineStarts []int, offset int) int {
	lo, hi := 0, len(lineStarts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if lineStarts[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// SupportedDeclarationLanguages returns the list of languages that have
// declaration patterns defined.
func SupportedDeclarationLanguages() []string {
	langs := make([]string, 0, len(languageDeclarationPatterns))
	for lang := range languageDeclarationPatterns {
		langs = append(langs, lang)
	}
	return langs
}
