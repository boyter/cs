// SPDX-License-Identifier: MIT
//
// The MIN ranking algorithm in this file is inspired by the Gigablast
// open-source search engine by Matt Wells.
// Source: https://github.com/gigablast/open-source-search-engine
// Gigablast is licensed under a modified Apache License 2.0.
//
// The core ideas adapted here:
//   - Document score = MIN of all query term-pair scores (not SUM)
//   - Proximity scoring via 1/(distance+1) between term occurrences
//   - Location-type weighting (Gigablast: title/heading/body/meta;
//     adapted here to: declaration/code/string/comment)

package ranker

import (
	"bytes"
	"math"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/scc/v3/processor"
)

// MinRankingProfile extends the base profile with MIN-specific knobs.
type MinRankingProfile struct {
	RankingProfile

	// ProximityWeight controls how strongly proximity affects pair scores.
	// Higher values make distance between terms matter more.
	// At 1.0, adjacent terms score ~6× higher than terms 50 bytes apart.
	// At 2.0, that ratio doubles. 0.0 disables proximity entirely.
	ProximityWeight float64

	// BlendFactor controls how much SUM-style scoring is mixed in.
	// 0.0 = pure MIN (Gigablast-style), 1.0 = pure SUM (traditional BM25).
	BlendFactor float64

	// DeclarationBoost multiplies term scores for matches on declaration lines
	// (func, class, type, etc). Analogous to Gigablast's title tag 8× boost.
	DeclarationBoost float64

	// CodeWeight is the weight for matches in code (vs comments/strings).
	// Analogous to Gigablast's body text weight of 1.0.
	CodeWeight float64

	// CommentWeight is the weight for matches in comments.
	// Analogous to Gigablast's meta tag weight of 0.1.
	CommentWeight float64

	// StringWeight is the weight for matches in string literals.
	StringWeight float64
}

// DefaultMinRankingProfile returns a MIN profile tuned to diverge from BM25.
func DefaultMinRankingProfile() MinRankingProfile {
	return MinRankingProfile{
		RankingProfile:   DefaultRankingProfile(),
		ProximityWeight:  1.5,
		BlendFactor:      0.1,
		DeclarationBoost: 3.0,
		CodeWeight:       1.0,
		CommentWeight:    0.3,
		StringWeight:     0.5,
	}
}

// rankResultsMinBM25 implements a MIN-of-term-pairs scoring algorithm
// inspired by Gigablast's approach. The key differences from BM25:
//
//  1. Score = MIN(all term-pair scores), not SUM(per-term scores).
//     A file is only as strong as its weakest query-term pair.
//  2. Proximity is a divisor (Gigablast-style), not a small bonus.
//     Terms appearing adjacent score dramatically higher than distant terms.
//  3. Location-type weighting: matches in declarations get boosted (like
//     Gigablast's title/heading hash groups), matches in comments get dampened.
//  4. Per-term scores use actual min(a,b), not geometric mean, so one
//     strong term cannot compensate for a weak one.
func rankResultsMinBM25(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int, profile MinRankingProfile) []*common.FileJob {
	if len(results) == 0 {
		return results
	}

	var averageDocumentWords float64
	for i := 0; i < len(results); i++ {
		averageDocumentWords += float64(maxInt(1, results[i].Bytes/BytesWordDivisor))
	}
	averageDocumentWords = averageDocumentWords / float64(len(results))

	k1 := profile.K1
	b := profile.B

	for i := 0; i < len(results); i++ {
		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		terms := make([]string, 0, len(results[i].MatchLocations))
		for term := range results[i].MatchLocations {
			terms = append(terms, term)
		}

		if len(terms) == 0 {
			results[i].Score = 0
			continue
		}

		// Compute per-term scores with location-type weighting
		termScores := make(map[string]float64, len(terms))
		var sumScore float64
		for _, term := range terms {
			ts := termScoreMin(term, results[i], corpusCount, documentFrequencies, words, averageDocumentWords, k1, b, profile)
			termScores[term] = ts
			sumScore += ts
		}

		// Single-term query: no pairs to compute, just use the term score
		if len(terms) == 1 {
			results[i].Score = termScores[terms[0]]
			continue
		}

		// MIN across all term pairs — the Gigablast core
		minPairScore := math.MaxFloat64

		for ti := 0; ti < len(terms); ti++ {
			for tj := ti + 1; tj < len(terms); tj++ {
				t1 := terms[ti]
				t2 := terms[tj]

				// Use actual min(a,b), not geometric mean.
				// This is what makes MIN diverge from BM25: one strong term
				// cannot compensate for a weak one at all.
				weakerScore := math.Min(termScores[t1], termScores[t2])

				// Proximity as a multiplier using Gigablast's formula.
				// Adjacent terms get the full score; distant terms get crushed.
				proxMult := proximityMultiplier(
					results[i].MatchLocations[t1],
					results[i].MatchLocations[t2],
					profile.ProximityWeight,
				)

				pairScore := weakerScore * proxMult

				if pairScore < minPairScore {
					minPairScore = pairScore
				}
			}
		}

		// Blend: mostly MIN with a small SUM safety net
		avgScore := sumScore / float64(len(terms))
		results[i].Score = (1.0-profile.BlendFactor)*minPairScore +
			profile.BlendFactor*avgScore
	}

	return results
}

// termScoreMin computes a location-weighted BM25 score for a single term.
// Unlike plain BM25 which counts all occurrences equally, this weights each
// occurrence by where it appears in the file — analogous to Gigablast's
// hash group weights (title=8×, heading=1.5×, body=1.0, meta=0.1).
//
// In code search the equivalents are:
//   - Declaration line (func/class/type): like a title tag — high signal
//   - Code: like body text — baseline signal
//   - String literal: moderate signal
//   - Comment: like a meta tag — low signal
func termScoreMin(term string, result *common.FileJob, corpusCount int, documentFrequencies map[string]int, words, avgWords, k1, b float64, profile MinRankingProfile) float64 {
	locs := result.MatchLocations[term]
	if len(locs) == 0 {
		return 0
	}

	// Weighted term frequency: each occurrence contributes based on its location type
	var weightedTf float64
	for _, loc := range locs {
		if len(loc) < 2 {
			continue
		}

		w := locationWeight(result.ContentByteType, loc[0], profile)

		// Declaration boost: if this match is on a declaration line, multiply further.
		// This is the code-search equivalent of Gigablast's title tag 8× boost.
		if profile.DeclarationBoost > 1.0 && result.Content != nil && len(result.Content) > 0 {
			if isOnDeclarationLine(result.Content, loc[0], result.Language) {
				w *= profile.DeclarationBoost
			}
		}

		weightedTf += w
	}

	if weightedTf == 0 {
		return 0
	}

	idf := math.Log10(1 + float64(corpusCount)/float64(maxInt(1, documentFrequencies[term])))

	step1 := idf * weightedTf * (k1 + 1)
	step2 := weightedTf + k1*(1-b+(b*words/avgWords))

	return step1 / step2
}

// locationWeight returns the weight for a match based on its byte-type classification.
func locationWeight(contentByteType []byte, startByte int, profile MinRankingProfile) float64 {
	if contentByteType == nil || startByte < 0 || startByte >= len(contentByteType) {
		return profile.CodeWeight
	}

	switch contentByteType[startByte] {
	case processor.ByteTypeComment:
		return profile.CommentWeight
	case processor.ByteTypeString:
		return profile.StringWeight
	case processor.ByteTypeCode:
		return profile.CodeWeight
	default:
		return profile.CodeWeight
	}
}

// isOnDeclarationLine checks whether the byte offset falls on a line that
// starts with a declaration pattern (func, class, type, def, etc).
func isOnDeclarationLine(content []byte, byteOffset int, language string) bool {
	if !HasDeclarationPatterns(language) {
		return false
	}
	if byteOffset < 0 || byteOffset >= len(content) {
		return false
	}

	// Walk backward to find the start of this line
	lineStart := byteOffset
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Walk forward to find end of line
	lineEnd := byteOffset
	for lineEnd < len(content) && content[lineEnd] != '\n' {
		lineEnd++
	}

	line := content[lineStart:lineEnd]
	trimmed := bytes.TrimLeft(line, " \t")

	return IsDeclarationLine(trimmed, language)
}

// proximityMultiplier computes a score multiplier based on how close two terms
// appear in the document. This uses Gigablast's core formula: 1 / (distance + 1),
// but applied as a multiplier scaled by the proximity weight.
//
// With weight=1.5 (default):
//   - Adjacent terms (dist=0):     multiplier = 1.0 + 1.5 = 2.5×
//   - ~30 bytes apart (~5 words):  multiplier = 1.0 + 1.5 * 0.17 = 1.25×
//   - ~300 bytes apart (~50 words): multiplier = 1.0 + 1.5 * 0.02 = 1.03×
//   - Very far apart:               multiplier ≈ 1.0× (no bonus)
func proximityMultiplier(locs1, locs2 [][]int, weight float64) float64 {
	if weight == 0 || len(locs1) == 0 || len(locs2) == 0 {
		return 1.0
	}

	minDist := minByteDist(locs1, locs2)
	if minDist < 0 {
		return 1.0
	}

	// Convert byte distance to approximate word distance (~6 bytes per word)
	wordDist := float64(minDist) / 6.0

	// Gigablast formula: 1 / (distance + 1)
	proxFactor := 1.0 / (wordDist + 1.0)

	return 1.0 + weight*proxFactor
}

// minByteDist finds the minimum byte distance between any pair of match
// locations from two terms. Returns -1 if no valid pair exists.
// Uses a two-pointer merge for efficiency when positions are sorted.
func minByteDist(locs1, locs2 [][]int) int {
	best := -1

	ai, bi := 0, 0
	for ai < len(locs1) && bi < len(locs2) {
		if len(locs1[ai]) < 2 {
			ai++
			continue
		}
		if len(locs2[bi]) < 2 {
			bi++
			continue
		}

		start1 := locs1[ai][0]
		end1 := locs1[ai][1]
		start2 := locs2[bi][0]
		end2 := locs2[bi][1]

		var dist int
		if start1 > end2 {
			dist = start1 - end2
		} else if start2 > end1 {
			dist = start2 - end1
		} else {
			// Overlapping
			return 0
		}

		if best < 0 || dist < best {
			best = dist
		}

		if start1 <= start2 {
			ai++
		} else {
			bi++
		}
	}

	return best
}
