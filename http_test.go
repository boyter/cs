// SPDX-License-Identifier: MIT

package main

import (
	"testing"

	"github.com/boyter/cs/pkg/common"
)

func TestHttpCalculatePages_Empty(t *testing.T) {
	pages := httpCalculatePages(nil, 10, "test", 300, "", "bm25", "", "", "")
	if len(pages) != 0 {
		t.Errorf("expected 0 pages for nil results, got %d", len(pages))
	}
}

func TestHttpCalculatePages_SinglePage(t *testing.T) {
	results := make([]*common.FileJob, 5)
	for i := range results {
		results[i] = &common.FileJob{}
	}
	pages := httpCalculatePages(results, 10, "test", 300, "", "bm25", "", "", "")
	if len(pages) != 1 {
		t.Errorf("expected 1 page for 5 results with pageSize 10, got %d", len(pages))
	}
}

func TestHttpCalculatePages_MultiplePages(t *testing.T) {
	results := make([]*common.FileJob, 25)
	for i := range results {
		results[i] = &common.FileJob{}
	}
	pages := httpCalculatePages(results, 10, "test", 300, "", "bm25", "", "", "")
	// 25 results / 10 per page = 3 pages (0, 1, 2)
	if len(pages) != 3 {
		t.Errorf("expected 3 pages for 25 results with pageSize 10, got %d", len(pages))
	}
}

func TestHttpCalculatePages_ExactMultiple(t *testing.T) {
	results := make([]*common.FileJob, 20)
	for i := range results {
		results[i] = &common.FileJob{}
	}
	pages := httpCalculatePages(results, 10, "test", 300, "", "bm25", "", "", "")
	// 20 results / 10 per page = exactly 2 pages
	if len(pages) != 2 {
		t.Errorf("expected 2 pages for 20 results with pageSize 10, got %d", len(pages))
	}
}

func TestHttpPaginationBoundary_NoOutOfBounds(t *testing.T) {
	// This test simulates the pagination logic in the search handler
	// to verify that page == len(pages) does NOT cause an out-of-bounds access.
	pageSize := 10
	results := make([]*common.FileJob, 25)
	for i := range results {
		results[i] = &common.FileJob{Filename: "test.go", Location: "test.go"}
	}

	pages := httpCalculatePages(results, pageSize, "test", 300, "", "bm25", "", "", "")
	// pages has 3 entries (indices 0, 1, 2)

	// Test that page == len(pages) (which is 3) does NOT enter the pagination block.
	// Before the fix, this would panic with page*pageSize (30) > len(results) (25).
	page := len(pages) // This is the off-by-one case
	displayResults := results
	if displayResults != nil && len(displayResults) > pageSize {
		displayResults = displayResults[:pageSize]
	}
	// With the fix: page < len(pages) prevents entering this block when page == len(pages)
	if page != 0 && page < len(pages) {
		end := page*pageSize + pageSize
		if end > len(results) {
			end = len(results)
		}
		displayResults = results[page*pageSize : end]
	}

	// Should still have the first page of results (not panic)
	if len(displayResults) != pageSize {
		t.Errorf("expected %d display results, got %d", pageSize, len(displayResults))
	}
}

func TestHttpPaginationLastValidPage(t *testing.T) {
	// Verify that the last valid page (len(pages)-1) works correctly.
	pageSize := 10
	results := make([]*common.FileJob, 25)
	for i := range results {
		results[i] = &common.FileJob{Filename: "test.go", Location: "test.go"}
	}

	pages := httpCalculatePages(results, pageSize, "test", 300, "", "bm25", "", "", "")

	page := len(pages) - 1 // Last valid page (index 2)
	displayResults := results
	if displayResults != nil && len(displayResults) > pageSize {
		displayResults = displayResults[:pageSize]
	}
	if page != 0 && page < len(pages) {
		end := page*pageSize + pageSize
		if end > len(results) {
			end = len(results)
		}
		displayResults = results[page*pageSize : end]
	}

	// Last page should have 5 results (25 - 20)
	if len(displayResults) != 5 {
		t.Errorf("expected 5 display results on last page, got %d", len(displayResults))
	}
}
