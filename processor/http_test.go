// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"testing"
)

func TestCalculatePagesNone(t *testing.T) {
	var pages = calculatePages([]*FileJob{}, 20, "", 100)

	if len(pages) != 0 {
		t.Error("expected no result")
	}
}

func TestCalculatePagesSingle(t *testing.T) {
	var pages = calculatePages([]*FileJob{
		{},
	}, 20, "", 100)

	if len(pages) != 1 {
		t.Error("expected single result")
	}

	if pages[0].SnippetSize != 100 {
		t.Error("expected snippet size to be 100")
	}

	if pages[0].Name != "1" {
		t.Error("page name incorrect")
	}
}

func TestCalculatePagesEdgeStart(t *testing.T) {
	var fj []*FileJob
	for i := 0; i < 20; i++ {
		fj = append(fj, &FileJob{})
	}

	var pages = calculatePages(fj, 20, "", 100)

	if len(pages) != 1 {
		t.Error("expected single result got", len(pages))
	}
}

func TestCalculatePagesEdgeOver(t *testing.T) {
	var fj []*FileJob
	for i := 0; i < 21; i++ {
		fj = append(fj, &FileJob{})
	}

	var pages = calculatePages(fj, 20, "", 100)

	if len(pages) != 2 {
		t.Error("expected two result got", len(pages))
	}
}

func TestCalculatePagesSecondPageEdge(t *testing.T) {
	var fj []*FileJob
	for i := 0; i < 40; i++ {
		fj = append(fj, &FileJob{})
	}

	var pages = calculatePages(fj, 20, "", 100)

	if len(pages) != 2 {
		t.Error("expected two result got", len(pages))
	}
}

func TestCalculatePagesSecondPageEdgeOver(t *testing.T) {
	var fj []*FileJob
	for i := 0; i < 41; i++ {
		fj = append(fj, &FileJob{})
	}

	var pages = calculatePages(fj, 20, "", 100)

	if len(pages) != 3 {
		t.Error("expected three result got", len(pages))
	}
}
