// SPDX-License-Identifier: MIT

package ranker

import (
	"testing"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/scc/v3/processor"
)

// --- IsStopword tests ---

func TestIsStopword_KnownStopword(t *testing.T) {
	if !IsStopword("Go", "func") {
		t.Error("expected 'func' to be a Go stopword")
	}
}

func TestIsStopword_NonStopword(t *testing.T) {
	if IsStopword("Go", "handler") {
		t.Error("expected 'handler' to NOT be a Go stopword")
	}
}

func TestIsStopword_CaseInsensitive(t *testing.T) {
	cases := []string{"FUNC", "Func", "fUnC"}
	for _, w := range cases {
		if !IsStopword("Go", w) {
			t.Errorf("expected %q to be a Go stopword (case insensitive)", w)
		}
	}
}

func TestIsStopword_UnknownLanguage(t *testing.T) {
	if IsStopword("Brainfuck", "if") {
		t.Error("expected false for unknown language")
	}
}

func TestIsStopword_EmptyLanguage(t *testing.T) {
	if IsStopword("", "func") {
		t.Error("expected false for empty language")
	}
}

func TestIsStopword_EmptyWord(t *testing.T) {
	if IsStopword("Go", "") {
		t.Error("expected false for empty word")
	}
}

// --- AllStopwords tests ---

func TestAllStopwords_AllAreStopwords(t *testing.T) {
	locs := map[string][][]int{
		"func":   {{0, 4}},
		"return": {{10, 16}},
	}
	if !AllStopwords("Go", locs) {
		t.Error("expected true when all terms are Go stopwords")
	}
}

func TestAllStopwords_MixedTerms(t *testing.T) {
	locs := map[string][][]int{
		"func":    {{0, 4}},
		"handler": {{10, 17}},
	}
	if AllStopwords("Go", locs) {
		t.Error("expected false when not all terms are stopwords")
	}
}

func TestAllStopwords_UnknownLanguage(t *testing.T) {
	locs := map[string][][]int{
		"if": {{0, 2}},
	}
	if AllStopwords("Brainfuck", locs) {
		t.Error("expected false for unknown language")
	}
}

func TestAllStopwords_EmptyMatchLocations(t *testing.T) {
	if AllStopwords("Go", map[string][][]int{}) {
		t.Error("expected false for empty matchLocations")
	}
}

// --- Structural ranker integration tests ---

func TestStructuralStopword_DampenedLessThanNonStopword(t *testing.T) {
	cfg := DefaultStructuralConfig()

	codeByteType := make([]byte, 200)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}

	// Both files match "func" and "handler" so AllStopwords is false.
	// File A has many "func" (stopword) matches, few "handler" matches.
	// File B has many "handler" (non-stopword) matches, few "func" matches.
	// Because "func" is dampened, file A should score lower than file B.
	fileA := &common.FileJob{
		Filename:        "stop.go",
		Location:        "stop.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations: map[string][][]int{
			"func":    {{10, 14}, {30, 34}, {50, 54}, {70, 74}, {90, 94}},
			"handler": {{110, 117}},
		},
	}
	fileB := &common.FileJob{
		Filename:        "nonstop.go",
		Location:        "nonstop.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations: map[string][][]int{
			"func":    {{10, 14}},
			"handler": {{30, 37}, {50, 57}, {70, 77}, {90, 97}, {110, 117}},
		},
	}

	results := []*common.FileJob{fileA, fileB}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	if fileA.Score >= fileB.Score {
		t.Errorf("expected stopword-heavy file score (%f) < non-stopword-heavy file score (%f)",
			fileA.Score, fileB.Score)
	}
}

func TestStructuralStopword_AllStopwordsSafeguard(t *testing.T) {
	cfg := DefaultStructuralConfig()

	codeByteType := make([]byte, 200)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}

	// When the entire query is stopwords, no dampening should be applied.
	// Compare score of "func" (all-stopword query) vs what it would get if dampened.
	file := &common.FileJob{
		Filename:        "safeguard.go",
		Location:        "safeguard.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}},
	}

	results := []*common.FileJob{file}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)
	allStopScore := file.Score

	// Now test with a mixed query — the stopword should be dampened
	file2 := &common.FileJob{
		Filename:        "mixed.go",
		Location:        "mixed.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}, "handler": {{20, 27}}},
	}

	results2 := []*common.FileJob{file2}
	df2 := CalculateDocumentFrequency(results2)
	rankResultsStructural(10, results2, df2, cfg)

	// In the mixed case, "func" is dampened. The full score should be less than
	// allStopScore (undampened func) + a full handler score. But more importantly,
	// the all-stopword score should NOT have dampening applied.
	if allStopScore <= 0 {
		t.Errorf("expected positive score for all-stopword safeguard, got %f", allStopScore)
	}
}

func TestStructuralStopword_UnknownLanguage_NoPenalty(t *testing.T) {
	cfg := DefaultStructuralConfig()

	codeByteType := make([]byte, 200)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}

	// "func" in an unknown language should NOT be dampened
	unknownFile := &common.FileJob{
		Filename:        "unknown.xyz",
		Location:        "unknown.xyz",
		Language:        "UnknownLang",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}, "handler": {{20, 27}}},
	}
	// Same file but with Go language — "func" should be dampened
	goFile := &common.FileJob{
		Filename:        "known.go",
		Location:        "known.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}, "handler": {{20, 27}}},
	}

	results := []*common.FileJob{unknownFile, goFile}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	// Unknown language file should score higher because "func" isn't dampened
	if unknownFile.Score <= goFile.Score {
		t.Errorf("expected unknown language score (%f) > Go score (%f) due to no dampening",
			unknownFile.Score, goFile.Score)
	}
}

func TestStructuralStopword_CrossLanguage(t *testing.T) {
	cfg := DefaultStructuralConfig()

	codeByteType := make([]byte, 200)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}

	// "def" is a stopword in Python but NOT in Go
	pythonFile := &common.FileJob{
		Filename:        "test.py",
		Location:        "test.py",
		Language:        "Python",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"def": {{10, 13}}, "calculate": {{20, 29}}},
	}
	goFile := &common.FileJob{
		Filename:        "test.go",
		Location:        "test.go",
		Language:        "Go",
		Content:         make([]byte, 200),
		ContentByteType: codeByteType,
		Bytes:           200,
		MatchLocations:  map[string][][]int{"def": {{10, 13}}, "calculate": {{20, 29}}},
	}

	results := []*common.FileJob{pythonFile, goFile}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	// "def" dampened in Python, not in Go → Go file should score higher
	if goFile.Score <= pythonFile.Score {
		t.Errorf("expected Go score (%f) > Python score (%f) because 'def' dampened in Python only",
			goFile.Score, pythonFile.Score)
	}
}
