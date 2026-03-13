// SPDX-License-Identifier: MIT

package ranker

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	str "github.com/boyter/go-string"

	"github.com/boyter/cs/v3/pkg/common"
	"github.com/boyter/scc/v3/processor"
)

// RankingProfile bundles BM25 parameters and post-ranking knobs into a single
// named configuration. Consumers can use DefaultRankingProfile(), PreciseProfile,
// or BroadProfile, or construct custom values.
type RankingProfile struct {
	K1               float64
	B                float64
	GravityStrength  float64
	NoiseSensitivity float64
	TestPenalty      float64
	LengthBias       float64
}

// DefaultRankingProfile returns the balanced profile matching the previously
// hardcoded values.
func DefaultRankingProfile() RankingProfile {
	return RankingProfile{K1: 1.2, B: 0.75, GravityStrength: 1.0, NoiseSensitivity: 1.0, TestPenalty: 0.4, LengthBias: 0.0}
}

// PreciseProfile favours short, focused files with exact matches.
var PreciseProfile = RankingProfile{K1: 0.3, B: 0.95, GravityStrength: 2.5, NoiseSensitivity: 2.5, TestPenalty: 0.1, LengthBias: 0.3}

// BroadProfile is exploratory: tolerates noise and long files, rewards repeated matches.
var BroadProfile = RankingProfile{K1: 2.5, B: 0.15, GravityStrength: 0.3, NoiseSensitivity: 0.3, TestPenalty: 1.0, LengthBias: -0.1}

// ResolveProfileByName maps a profile name to a RankingProfile.
// Unrecognised names (including "") return the default balanced profile.
func ResolveProfileByName(name string) *RankingProfile {
	switch strings.ToLower(name) {
	case "precise":
		p := PreciseProfile
		return &p
	case "broad":
		p := BroadProfile
		return &p
	default:
		p := DefaultRankingProfile()
		return &p
	}
}

// Base value used to determine how much location matches
// should be boosted by
const (
	LocationBoostValue = 0.05
	DefaultScoreValue  = 0.01
	BytesWordDivisor   = 2
)

// StructuralConfig holds weights and filters for the structural ranker.
type StructuralConfig struct {
	WeightCode    float64
	WeightComment float64
	WeightString  float64
	OnlyCode      bool
	OnlyComments  bool
	OnlyStrings   bool
}

// DefaultStructuralConfig returns a StructuralConfig with sensible defaults.
func DefaultStructuralConfig() StructuralConfig {
	return StructuralConfig{
		WeightCode:    1.0,
		WeightComment: 0.2,
		WeightString:  0.5,
	}
}

// RankResults takes in the search results and applies chained
// ranking over them to produce a score and then sort those results
// and return them.
// The rankerName parameter selects the algorithm: "simple", "bm25", "tfidf",
// "structural", or anything else for classic TF-IDF.
// structuralCfg is only used when rankerName is "structural" and may be nil otherwise.
func RankResults(rankerName string, corpusCount int, results []*common.FileJob, structuralCfg *StructuralConfig, profile *RankingProfile, testIntent bool) []*common.FileJob {
	if profile == nil {
		d := DefaultRankingProfile()
		profile = &d
	}

	// needs to come first because it resets the scores
	switch rankerName {
	case "simple":
		// in this case the results are already ranked by the number of matches
	case "structural":
		cfg := DefaultStructuralConfig()
		if structuralCfg != nil {
			cfg = *structuralCfg
		}
		results = rankResultsStructural(corpusCount, results, CalculateDocumentFrequency(results), cfg, *profile)
		results = rankResultsLocation(results)
	case "bm25":
		results = rankResultsBM25(corpusCount, results, CalculateDocumentFrequency(results), *profile)
		results = rankResultsLocation(results)
	case "tfidf":
		results = rankResultsTFIDF(corpusCount, results, CalculateDocumentFrequency(results), false)
		results = rankResultsLocation(results)
	default:
		results = rankResultsTFIDF(corpusCount, results, CalculateDocumentFrequency(results), true)
		results = rankResultsLocation(results)
	}

	results = rankResultsComplexityGravity(results, profile.GravityStrength)

	// Noise penalty applies to all rankers (not just structural) because it is
	// a file-level property (size vs complexity), not a term-frequency calculation.
	// The structural ranker will become the default in future, but until then this
	// ensures the penalty is active regardless of ranker choice.
	results = rankResultsNoisePenalty(results, profile.NoiseSensitivity)
	results = rankResultsLengthBias(results, profile.LengthBias)
	results = rankResultsTestDampening(results, profile.TestPenalty, testIntent)

	sortResults(results)
	return results
}

// Given the results will boost the rank of them based on matches in the
// file location field.
func rankResultsLocation(results []*common.FileJob) []*common.FileJob {
	for i := 0; i < len(results); i++ {
		foundTerms := 0
		for key := range results[i].MatchLocations {
			l := str.IndexAllIgnoreCase(results[i].Location, key, -1)

			if len(l) != 0 {
				foundTerms++

				if results[i].Score == 0 || math.IsNaN(results[i].Score) {
					results[i].Score = DefaultScoreValue
				}

				results[i].Score = results[i].Score * (1.0 +
					(LocationBoostValue * float64(len(l)) * float64(len(key))))

				low := math.MaxInt32
				for _, l := range l {
					if l[0] < low {
						low = l[0]
					}
				}

				results[i].Score = results[i].Score * (1.0 / (1.0 + float64(low)*0.02))
			}
		}

		if foundTerms > 1 {
			results[i].Score = results[i].Score * (1 + LocationBoostValue*float64(foundTerms))
		}
	}

	return results
}

// TF-IDF implementation which ranks the results
func rankResultsTFIDF(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int, classic bool) []*common.FileJob {
	var weight float64
	for i := 0; i < len(results); i++ {
		weight = 0

		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		for word, wordCount := range results[i].MatchLocations {
			tf := float64(len(wordCount)) / words
			idf := math.Log10(float64(corpusCount) / float64(maxInt(1, documentFrequencies[word])))

			if classic {
				weight += tf * idf
			} else {
				weight += math.Sqrt(tf) * idf * (1 / math.Sqrt(words))
			}
		}

		results[i].Score = weight
	}

	return results
}

// BM25 implementation which ranks the results
//
//	IDF * TF * (k1 + 1)
//
// BM25 = sum ----------------------------
//
//	TF + k1 * (1 - b + b * D / L)
func rankResultsBM25(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int, profile RankingProfile) []*common.FileJob {
	if len(results) == 0 {
		return results
	}

	var weight float64

	var averageDocumentWords float64
	for i := 0; i < len(results); i++ {
		averageDocumentWords += float64(maxInt(1, results[i].Bytes/BytesWordDivisor))
	}
	averageDocumentWords = averageDocumentWords / float64(len(results))

	k1 := profile.K1
	b := profile.B

	for i := 0; i < len(results); i++ {
		weight = 0

		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))

		for word, wordCount := range results[i].MatchLocations {
			rawCount := float64(len(wordCount))
			idf := math.Log10(1 + float64(corpusCount)/float64(maxInt(1, documentFrequencies[word])))

			step1 := idf * rawCount * (k1 + 1)
			step2 := rawCount + k1*(1-b+(b*words/averageDocumentWords))

			weight += step1 / step2
		}

		results[i].Score = weight
	}

	return results
}

// CalculateDocumentTermFrequency calculates the document term frequency for all words
// across all documents, letting us know how many times a term appears across the corpus.
// This is mostly used for snippet extraction.
func CalculateDocumentTermFrequency(results []*common.FileJob) map[string]int {
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + len(results[i].MatchLocations[k])
		}
	}

	return documentFrequencies
}

// CalculateDocumentFrequency calculates the document frequency for all words
// across all documents, allowing us to know the number of documents for which a term appears.
// This is mostly used for TF-IDF calculation.
func CalculateDocumentFrequency(results []*common.FileJob) map[string]int {
	documentFrequencies := map[string]int{}
	for i := 0; i < len(results); i++ {
		for k := range results[i].MatchLocations {
			documentFrequencies[k] = documentFrequencies[k] + 1
		}
	}

	return documentFrequencies
}

// Sort a slice of filejob results based on their score for displaying
// and then sort based on location to stop any undeterministic ordering happening
func sortResults(results []*common.FileJob) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return strings.Compare(results[i].Location, results[j].Location) < 0
		}

		return results[i].Score > results[j].Score
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// matchWeight returns the weight for a match at startByte based on the byte
// type classification. Falls back to WeightCode when contentByteType is nil
// (unrecognised language) or the offset is out of bounds.
func matchWeight(contentByteType []byte, startByte int, cfg StructuralConfig) float64 {
	if contentByteType == nil || startByte < 0 || startByte >= len(contentByteType) {
		if cfg.OnlyComments || cfg.OnlyStrings {
			return 0
		}
		return cfg.WeightCode
	}

	switch contentByteType[startByte] {
	case processor.ByteTypeComment:
		if cfg.OnlyCode || cfg.OnlyStrings {
			return 0
		}
		return cfg.WeightComment
	case processor.ByteTypeString:
		if cfg.OnlyCode || cfg.OnlyComments {
			return 0
		}
		return cfg.WeightString
	case processor.ByteTypeCode:
		if cfg.OnlyComments || cfg.OnlyStrings {
			return 0
		}
		return cfg.WeightCode
	default:
		// ByteTypeBlank or unknown — treat as code
		if cfg.OnlyComments || cfg.OnlyStrings {
			return 0
		}
		return cfg.WeightCode
	}
}

// rankResultsComplexityGravity applies a post-ranking boost based on each file's
// cyclomatic complexity. Complex code files get boosted; prose/markdown (complexity=0)
// are unaffected. Formula: Score_final = Score_base * (1 + (ln(1 + Complexity) * Strength))
func rankResultsComplexityGravity(results []*common.FileJob, gravityStrength float64) []*common.FileJob {
	if gravityStrength == 0 {
		return results
	}
	for i := 0; i < len(results); i++ {
		if results[i].Score == 0 {
			continue
		}
		boost := math.Log1p(float64(results[i].Complexity))
		results[i].Score *= (1.0 + (boost * gravityStrength))
	}
	return results
}

// rankResultsNoisePenalty dampens scores for files with low information density
// (large size but little logical complexity). Files like minified JS, JSON blobs,
// SQL dumps, and logs get penalised while clean code files are left untouched.
// Formula: SignalRatio = (Complexity + 1) / log10(Bytes), Penalty = min(1.0, SignalRatio * Sensitivity)
func rankResultsNoisePenalty(results []*common.FileJob, sensitivity float64) []*common.FileJob {
	if sensitivity >= 100.0 {
		return results
	}
	for i := 0; i < len(results); i++ {
		if results[i].Score == 0 {
			continue
		}

		safeBytes := math.Max(10, float64(results[i].Bytes))
		fileSizeLog := math.Log10(safeBytes)

		signalRatio := (float64(results[i].Complexity) + 1.0) / fileSizeLog
		penalty := math.Min(1.0, signalRatio*sensitivity)

		results[i].Score *= penalty
	}
	return results
}

// rankResultsLengthBias adjusts scores based on file size (in bytes).
// Positive bias penalises longer files; negative bias gives them a slight boost.
// Formula: Score *= 1.0 / (1.0 + LengthBias * log10(max(10, Bytes)))
func rankResultsLengthBias(results []*common.FileJob, lengthBias float64) []*common.FileJob {
	if lengthBias == 0 {
		return results
	}
	for i := 0; i < len(results); i++ {
		if results[i].Score == 0 {
			continue
		}
		safeBytes := math.Max(10, float64(results[i].Bytes))
		results[i].Score *= 1.0 / (1.0 + lengthBias*math.Log10(safeBytes))
	}
	return results
}

// IsTestFile returns true if the file path looks like a test file based on
// common naming conventions across languages.
func IsTestFile(path string) bool {
	lower := strings.ToLower(path)
	base := strings.ToLower(filepath.Base(path))

	// Directory patterns
	if strings.Contains(lower, "/tests/") || strings.Contains(lower, "/test/") || strings.Contains(lower, "/__tests__/") {
		return true
	}

	// Filename patterns: _test., .test., .spec., test_
	if strings.Contains(base, "_test.") || strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.HasPrefix(base, "test_") {
		return true
	}

	// Base name (before extension) ends with "test" or "tests"
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if strings.HasSuffix(name, "test") || strings.HasSuffix(name, "tests") {
		return true
	}

	return false
}

// HasTestIntent returns true if any of the query terms indicate the user is
// searching for test-related code.
func HasTestIntent(queryTerms []string) bool {
	intentWords := map[string]struct{}{
		"test": {}, "assert": {}, "expect": {}, "mock": {},
		"stub": {}, "bench": {}, "suite": {}, "spec": {}, "should": {},
	}
	for _, term := range queryTerms {
		if _, ok := intentWords[strings.ToLower(term)]; ok {
			return true
		}
	}
	return false
}

// rankResultsTestDampening adjusts scores for test files. When the query has
// no test intent, test files are penalised by testPenalty (e.g. 0.4). When
// the query has test intent, test files are boosted (×1.5).
func rankResultsTestDampening(results []*common.FileJob, testPenalty float64, testIntent bool) []*common.FileJob {
	if testPenalty == 1.0 && !testIntent {
		return results
	}
	for i := 0; i < len(results); i++ {
		if results[i].Score == 0 {
			continue
		}
		if !IsTestFile(results[i].Location) {
			continue
		}
		if testIntent {
			results[i].Score *= 1.5
		} else {
			results[i].Score *= testPenalty
		}
	}
	return results
}

// rankResultsStructural is a BM25 variant that weights term frequency by the
// structural type (code/comment/string) of each match location.
func rankResultsStructural(corpusCount int, results []*common.FileJob, documentFrequencies map[string]int, cfg StructuralConfig, profile RankingProfile) []*common.FileJob {
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
		var weight float64
		words := float64(maxInt(1, results[i].Bytes/BytesWordDivisor))
		allStop := AllStopwords(results[i].Language, results[i].MatchLocations)

		for word, wordLocs := range results[i].MatchLocations {
			var weightedTf float64
			for _, loc := range wordLocs {
				if len(loc) < 2 {
					continue
				}
				weightedTf += matchWeight(results[i].ContentByteType, loc[0], cfg)
			}

			idf := math.Log10(1 + float64(corpusCount)/float64(maxInt(1, documentFrequencies[word])))

			step1 := idf * weightedTf * (k1 + 1)
			step2 := weightedTf + k1*(1-b+(b*words/averageDocumentWords))

			termScore := step1 / step2
			if !allStop && IsStopword(results[i].Language, word) {
				termScore *= StopwordDampenFactor
			}
			weight += termScore
		}

		results[i].Score = weight
	}

	return results
}
