// SPDX-License-Identifier: MIT

package ranker

import (
	"testing"

	"github.com/boyter/cs/v3/pkg/common"
)

func TestComputeMatchHash_Identical(t *testing.T) {
	a := &common.FileJob{
		Content:        []byte("hello world"),
		MatchLocations: map[string][][]int{"hello": {{0, 5}}},
	}
	b := &common.FileJob{
		Content:        []byte("hello world"),
		MatchLocations: map[string][][]int{"hello": {{0, 5}}},
	}

	ha := ComputeMatchHash(a)
	hb := ComputeMatchHash(b)

	if ha == "" {
		t.Fatal("expected non-empty hash")
	}
	if ha != hb {
		t.Fatalf("expected identical hashes, got %s and %s", ha, hb)
	}
}

func TestComputeMatchHash_Different(t *testing.T) {
	a := &common.FileJob{
		Content:        []byte("hello world"),
		MatchLocations: map[string][][]int{"hello": {{0, 5}}},
	}
	b := &common.FileJob{
		Content:        []byte("goodbye world"),
		MatchLocations: map[string][][]int{"goodbye": {{0, 7}}},
	}

	ha := ComputeMatchHash(a)
	hb := ComputeMatchHash(b)

	if ha == hb {
		t.Fatal("expected different hashes for different content")
	}
}

func TestComputeMatchHash_Empty(t *testing.T) {
	a := &common.FileJob{
		Content:        nil,
		MatchLocations: nil,
	}
	h := ComputeMatchHash(a)
	if h != "" {
		t.Fatalf("expected empty hash for empty FileJob, got %s", h)
	}

	b := &common.FileJob{
		Content:        []byte("hello"),
		MatchLocations: map[string][][]int{},
	}
	h = ComputeMatchHash(b)
	if h != "" {
		t.Fatalf("expected empty hash for no match locations, got %s", h)
	}
}

func TestDeduplicateResults_NoDuplicates(t *testing.T) {
	results := []*common.FileJob{
		{
			Location:       "a.go",
			Score:          10,
			Content:        []byte("aaa"),
			MatchLocations: map[string][][]int{"a": {{0, 1}}},
		},
		{
			Location:       "b.go",
			Score:          8,
			Content:        []byte("bbb"),
			MatchLocations: map[string][][]int{"b": {{0, 1}}},
		},
	}

	out := DeduplicateResults(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if out[0].DuplicateCount != 0 {
		t.Fatalf("expected 0 duplicates, got %d", out[0].DuplicateCount)
	}
}

func TestDeduplicateResults_WithDuplicates(t *testing.T) {
	// 3 identical + 2 unique = 3 results
	mkDup := func(loc string, score float64) *common.FileJob {
		return &common.FileJob{
			Location:       loc,
			Score:          score,
			Content:        []byte("same content here"),
			MatchLocations: map[string][][]int{"same": {{0, 4}}},
		}
	}

	results := []*common.FileJob{
		mkDup("dup1.go", 10),
		mkDup("dup2.go", 8),
		{
			Location:       "unique1.go",
			Score:          6,
			Content:        []byte("unique stuff"),
			MatchLocations: map[string][][]int{"unique": {{0, 6}}},
		},
		mkDup("dup3.go", 4),
		{
			Location:       "unique2.go",
			Score:          2,
			Content:        []byte("other content"),
			MatchLocations: map[string][][]int{"other": {{0, 5}}},
		},
	}

	out := DeduplicateResults(results)
	if len(out) != 3 {
		t.Fatalf("expected 3 results, got %d", len(out))
	}

	// First should be the dup representative
	if out[0].Location != "dup1.go" {
		t.Fatalf("expected dup1.go as representative, got %s", out[0].Location)
	}
	if out[0].DuplicateCount != 2 {
		t.Fatalf("expected 2 duplicates, got %d", out[0].DuplicateCount)
	}
}

func TestDeduplicateResults_HighestScoreKept(t *testing.T) {
	mk := func(loc string, score float64) *common.FileJob {
		return &common.FileJob{
			Location:       loc,
			Score:          score,
			Content:        []byte("identical"),
			MatchLocations: map[string][][]int{"identical": {{0, 9}}},
		}
	}

	results := []*common.FileJob{
		mk("high.go", 100),
		mk("low.go", 1),
	}

	out := DeduplicateResults(results)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Location != "high.go" {
		t.Fatalf("expected high.go kept, got %s", out[0].Location)
	}
	if out[0].Score != 100 {
		t.Fatalf("expected score 100, got %f", out[0].Score)
	}
}

func TestDeduplicateResults_Empty(t *testing.T) {
	out := DeduplicateResults(nil)
	if len(out) != 0 {
		t.Fatalf("expected 0 results, got %d", len(out))
	}

	out = DeduplicateResults([]*common.FileJob{})
	if len(out) != 0 {
		t.Fatalf("expected 0 results, got %d", len(out))
	}
}

func TestDeduplicateResults_AllIdentical(t *testing.T) {
	mk := func(loc string) *common.FileJob {
		return &common.FileJob{
			Location:       loc,
			Score:          5,
			Content:        []byte("same"),
			MatchLocations: map[string][][]int{"same": {{0, 4}}},
		}
	}

	results := []*common.FileJob{
		mk("a.go"),
		mk("b.go"),
		mk("c.go"),
		mk("d.go"),
		mk("e.go"),
	}

	out := DeduplicateResults(results)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Location != "a.go" {
		t.Fatalf("expected a.go as representative, got %s", out[0].Location)
	}
	if out[0].DuplicateCount != 4 {
		t.Fatalf("expected 4 duplicates, got %d", out[0].DuplicateCount)
	}
	if len(out[0].DuplicateLocations) != 4 {
		t.Fatalf("expected 4 duplicate locations, got %d", len(out[0].DuplicateLocations))
	}
}
