// SPDX-License-Identifier: MIT

package ranker

import "strings"

// StopwordDampenFactor is the multiplier applied to stopword term scores.
// A value of 0.1 means stopwords contribute 10% of their original weight.
const StopwordDampenFactor = 0.1

// languageStopwords maps scc language names to sets of common syntactic keywords
// that are ubiquitous in that language. These are dampened in the structural ranker
// so that meaningful identifiers rank higher than language noise.
var languageStopwords = map[string]map[string]struct{}{
	"Go": {
		"break": {}, "case": {}, "chan": {}, "const": {}, "continue": {},
		"default": {}, "defer": {}, "else": {}, "for": {}, "func": {},
		"go": {}, "if": {}, "import": {}, "interface": {}, "map": {},
		"package": {}, "range": {}, "return": {}, "struct": {}, "switch": {},
		"type": {}, "var": {},
	},
	"Rust": {
		"as": {}, "break": {}, "const": {}, "continue": {}, "else": {},
		"enum": {}, "fn": {}, "for": {}, "if": {}, "impl": {},
		"in": {}, "let": {}, "loop": {}, "match": {}, "mod": {},
		"mut": {}, "pub": {}, "return": {}, "self": {}, "struct": {},
		"type": {}, "use": {},
	},
	"Java": {
		"abstract": {}, "break": {}, "case": {}, "catch": {}, "class": {},
		"continue": {}, "default": {}, "do": {}, "else": {}, "extends": {},
		"final": {}, "for": {}, "if": {}, "implements": {}, "import": {},
		"new": {}, "package": {}, "private": {}, "public": {}, "return": {},
		"static": {}, "throw": {}, "try": {}, "void": {},
	},
	"Python": {
		"and": {}, "as": {}, "assert": {}, "break": {}, "class": {},
		"continue": {}, "def": {}, "del": {}, "elif": {}, "else": {},
		"except": {}, "for": {}, "from": {}, "if": {}, "import": {},
		"in": {}, "is": {}, "not": {}, "or": {}, "pass": {},
		"raise": {}, "return": {}, "try": {}, "while": {}, "with": {},
	},
	"C": {
		"auto": {}, "break": {}, "case": {}, "char": {}, "const": {},
		"continue": {}, "default": {}, "do": {}, "double": {}, "else": {},
		"enum": {}, "extern": {}, "for": {}, "if": {}, "int": {},
		"long": {}, "return": {}, "static": {}, "struct": {}, "switch": {},
		"typedef": {}, "void": {},
	},
	"C++": {
		"auto": {}, "break": {}, "case": {}, "catch": {}, "class": {},
		"const": {}, "continue": {}, "default": {}, "delete": {}, "do": {},
		"else": {}, "for": {}, "if": {}, "namespace": {}, "new": {},
		"private": {}, "public": {}, "return": {}, "static": {}, "struct": {},
		"template": {}, "throw": {}, "try": {}, "virtual": {}, "void": {},
	},
	"C Header": {
		"auto": {}, "break": {}, "case": {}, "char": {}, "const": {},
		"continue": {}, "default": {}, "do": {}, "double": {}, "else": {},
		"enum": {}, "extern": {}, "for": {}, "if": {}, "int": {},
		"long": {}, "return": {}, "static": {}, "struct": {}, "switch": {},
		"typedef": {}, "void": {},
	},
	"C#": {
		"abstract": {}, "break": {}, "case": {}, "catch": {}, "class": {},
		"const": {}, "continue": {}, "default": {}, "do": {}, "else": {},
		"for": {}, "foreach": {}, "if": {}, "internal": {}, "namespace": {},
		"new": {}, "private": {}, "public": {}, "return": {}, "static": {},
		"throw": {}, "try": {}, "using": {}, "void": {},
	},
	"JavaScript": {
		"break": {}, "case": {}, "catch": {}, "class": {}, "const": {},
		"continue": {}, "default": {}, "do": {}, "else": {}, "export": {},
		"for": {}, "function": {}, "if": {}, "import": {}, "let": {},
		"new": {}, "return": {}, "switch": {}, "throw": {}, "try": {},
		"var": {}, "while": {},
	},
	"TypeScript": {
		"break": {}, "case": {}, "catch": {}, "class": {}, "const": {},
		"continue": {}, "default": {}, "do": {}, "else": {}, "export": {},
		"for": {}, "function": {}, "if": {}, "import": {}, "interface": {},
		"let": {}, "new": {}, "return": {}, "switch": {}, "throw": {},
		"try": {}, "type": {}, "var": {}, "while": {},
	},
	"Ruby": {
		"begin": {}, "break": {}, "case": {}, "class": {}, "def": {},
		"do": {}, "else": {}, "elsif": {}, "end": {}, "ensure": {},
		"for": {}, "if": {}, "include": {}, "module": {}, "next": {},
		"nil": {}, "require": {}, "rescue": {}, "return": {}, "self": {},
		"then": {}, "unless": {}, "while": {},
	},
	"PHP": {
		"abstract": {}, "break": {}, "case": {}, "catch": {}, "class": {},
		"const": {}, "continue": {}, "default": {}, "do": {}, "echo": {},
		"else": {}, "extends": {}, "for": {}, "foreach": {}, "function": {},
		"if": {}, "namespace": {}, "new": {}, "private": {}, "public": {},
		"return": {}, "static": {}, "throw": {}, "try": {}, "use": {},
	},
	"Swift": {
		"break": {}, "case": {}, "catch": {}, "class": {}, "continue": {},
		"default": {}, "do": {}, "else": {}, "enum": {}, "for": {},
		"func": {}, "guard": {}, "if": {}, "import": {}, "in": {},
		"let": {}, "private": {}, "public": {}, "return": {}, "self": {},
		"struct": {}, "switch": {}, "throw": {}, "try": {}, "var": {},
	},
	"Kotlin": {
		"abstract": {}, "break": {}, "class": {}, "continue": {}, "do": {},
		"else": {}, "for": {}, "fun": {}, "if": {}, "import": {},
		"in": {}, "interface": {}, "is": {}, "object": {}, "override": {},
		"package": {}, "private": {}, "public": {}, "return": {}, "val": {},
		"var": {}, "when": {}, "while": {},
	},
	"Scala": {
		"abstract": {}, "case": {}, "catch": {}, "class": {}, "def": {},
		"do": {}, "else": {}, "extends": {}, "final": {}, "for": {},
		"if": {}, "implicit": {}, "import": {}, "match": {}, "new": {},
		"object": {}, "override": {}, "package": {}, "private": {}, "return": {},
		"trait": {}, "try": {}, "type": {}, "val": {}, "var": {},
	},
	"Perl": {
		"chomp": {}, "die": {}, "do": {}, "else": {}, "elsif": {},
		"for": {}, "foreach": {}, "if": {}, "last": {}, "local": {},
		"my": {}, "next": {}, "our": {}, "package": {}, "print": {},
		"return": {}, "sub": {}, "unless": {}, "use": {}, "while": {},
	},
	"Lua": {
		"and": {}, "break": {}, "do": {}, "else": {}, "elseif": {},
		"end": {}, "for": {}, "function": {}, "if": {}, "in": {},
		"local": {}, "nil": {}, "not": {}, "or": {}, "repeat": {},
		"return": {}, "then": {}, "until": {}, "while": {},
	},
	"Dart": {
		"abstract": {}, "as": {}, "break": {}, "case": {}, "catch": {},
		"class": {}, "const": {}, "continue": {}, "default": {}, "do": {},
		"else": {}, "extends": {}, "final": {}, "for": {}, "if": {},
		"import": {}, "in": {}, "new": {}, "return": {}, "static": {},
		"switch": {}, "throw": {}, "try": {}, "var": {}, "void": {},
	},
	"Haskell": {
		"case": {}, "class": {}, "data": {}, "default": {}, "deriving": {},
		"do": {}, "else": {}, "if": {}, "import": {}, "in": {},
		"instance": {}, "let": {}, "module": {}, "newtype": {}, "of": {},
		"then": {}, "type": {}, "where": {},
	},
	"Shell": {
		"case": {}, "do": {}, "done": {}, "echo": {}, "elif": {},
		"else": {}, "esac": {}, "exit": {}, "fi": {}, "for": {},
		"function": {}, "if": {}, "in": {}, "local": {}, "return": {},
		"then": {}, "while": {},
	},
}

// IsStopword returns true if word is a common syntactic keyword for the given language.
// The check is case-insensitive. Returns false for unknown languages or empty inputs.
func IsStopword(language, word string) bool {
	if language == "" || word == "" {
		return false
	}
	words, ok := languageStopwords[language]
	if !ok {
		return false
	}
	_, found := words[strings.ToLower(word)]
	return found
}

// AllStopwords returns true only if every key in matchLocations is a stopword
// for the given language. Returns false for unknown languages or empty maps.
// This is used as a safeguard: when the entire query consists of stopwords,
// no dampening is applied (enabling keyword pattern searches).
func AllStopwords(language string, matchLocations map[string][][]int) bool {
	if len(matchLocations) == 0 {
		return false
	}
	words, ok := languageStopwords[language]
	if !ok {
		return false
	}
	for term := range matchLocations {
		if _, found := words[strings.ToLower(term)]; !found {
			return false
		}
	}
	return true
}
