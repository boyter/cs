// SPDX-License-Identifier: MIT

package ranker

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/boyter/cs/pkg/common"
)

// ComputeMatchHash returns a SHA-256 hex digest of the concatenated matched
// byte regions in fj, sorted by position. Returns "" if there are no match
// locations or no content.
func ComputeMatchHash(fj *common.FileJob) string {
	if len(fj.MatchLocations) == 0 || len(fj.Content) == 0 {
		return ""
	}

	// Collect all [start, end] spans
	var spans [][]int
	for _, locs := range fj.MatchLocations {
		spans = append(spans, locs...)
	}

	// Sort by start position, then by end position
	sort.Slice(spans, func(i, j int) bool {
		if spans[i][0] == spans[j][0] {
			return spans[i][1] < spans[j][1]
		}
		return spans[i][0] < spans[j][0]
	})

	h := sha256.New()
	for _, span := range spans {
		start, end := span[0], span[1]
		if start < 0 {
			start = 0
		}
		if end > len(fj.Content) {
			end = len(fj.Content)
		}
		if start >= end {
			continue
		}
		h.Write(fj.Content[start:end])
	}

	return hex.EncodeToString(h.Sum(nil))
}

// DeduplicateResults groups results by their MatchHash (computed if not already
// set), keeps the first result in each group (highest-scored, assuming the
// input is already sorted by score descending), and populates DuplicateCount
// and DuplicateLocations on the representative.
func DeduplicateResults(results []*common.FileJob) []*common.FileJob {
	if len(results) == 0 {
		return results
	}

	// Compute hashes for any results that don't have one yet
	for _, fj := range results {
		if fj.MatchHash == "" {
			fj.MatchHash = ComputeMatchHash(fj)
		}
	}

	type group struct {
		representative *common.FileJob
		locations      []string
	}

	seen := map[string]*group{}
	var order []string // preserve insertion order

	for _, fj := range results {
		hash := fj.MatchHash
		if hash == "" {
			// No hash means no match content — keep as-is (unique)
			hash = "[[no-hash]]:" + fj.Location
		}

		if g, ok := seen[hash]; ok {
			g.locations = append(g.locations, fj.Location)
		} else {
			seen[hash] = &group{
				representative: fj,
			}
			order = append(order, hash)
		}
	}

	out := make([]*common.FileJob, 0, len(order))
	for _, hash := range order {
		g := seen[hash]
		g.representative.DuplicateCount = len(g.locations)
		g.representative.DuplicateLocations = g.locations
		out = append(out, g.representative)
	}

	return out
}
