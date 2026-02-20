// SPDX-License-Identifier: MIT

package snippet

import (
	"math"
	"testing"

	"github.com/boyter/cs/pkg/common"
)

func TestExtractRelevant_ZeroDocumentFrequency_NoInf(t *testing.T) {
	res := &common.FileJob{
		Content: []byte("hello world foo bar baz"),
		MatchLocations: map[string][][]int{
			"hello": {{0, 5}},
		},
	}

	// Empty documentFrequencies map — word not present → defaults to 0
	df := map[string]int{}
	snippets := ExtractRelevant(res, df, 200)

	// Should not panic and should produce results
	if len(snippets) == 0 {
		t.Fatal("expected at least one snippet")
	}

	for _, s := range snippets {
		if math.IsInf(s.Score, 0) || math.IsNaN(s.Score) {
			t.Errorf("expected finite score, got %f", s.Score)
		}
	}
}

func TestExtractRelevant_WithDocumentFrequency(t *testing.T) {
	res := &common.FileJob{
		Content: []byte("hello world hello again"),
		MatchLocations: map[string][][]int{
			"hello": {{0, 5}, {12, 17}},
		},
	}

	df := map[string]int{"hello": 3}
	snippets := ExtractRelevant(res, df, 200)

	if len(snippets) == 0 {
		t.Fatal("expected at least one snippet")
	}

	for _, s := range snippets {
		if math.IsInf(s.Score, 0) || math.IsNaN(s.Score) {
			t.Errorf("expected finite score, got %f", s.Score)
		}
	}
}

func TestExtractRelevant_Empty(t *testing.T) {
	res := &common.FileJob{
		Content:        []byte{},
		MatchLocations: map[string][][]int{},
	}

	df := map[string]int{}
	snippets := ExtractRelevant(res, df, 200)

	if len(snippets) != 0 {
		t.Errorf("expected 0 snippets for empty input, got %d", len(snippets))
	}
}
