// SPDX-License-Identifier: MIT

package ranker

import (
	"testing"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/scc/v3/processor"
)

// --- matchWeight tests ---

func TestMatchWeight_NilByteType_FallbackToCode(t *testing.T) {
	cfg := DefaultStructuralConfig()
	w := matchWeight(nil, 0, cfg)
	if w != cfg.WeightCode {
		t.Errorf("expected %f, got %f", cfg.WeightCode, w)
	}
}

func TestMatchWeight_CodeByte(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeCode}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightCode {
		t.Errorf("expected %f, got %f", cfg.WeightCode, w)
	}
}

func TestMatchWeight_CommentByte(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeComment}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightComment {
		t.Errorf("expected %f, got %f", cfg.WeightComment, w)
	}
}

func TestMatchWeight_StringByte(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeString}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightString {
		t.Errorf("expected %f, got %f", cfg.WeightString, w)
	}
}

func TestMatchWeight_BlankByte(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeBlank}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightCode {
		t.Errorf("expected %f (blank treated as code), got %f", cfg.WeightCode, w)
	}
}

func TestMatchWeight_OnlyCode_ZerosComments(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyCode = true
	bt := []byte{processor.ByteTypeComment}
	w := matchWeight(bt, 0, cfg)
	if w != 0 {
		t.Errorf("expected 0 for comment with OnlyCode, got %f", w)
	}
}

func TestMatchWeight_OnlyCode_ZerosStrings(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyCode = true
	bt := []byte{processor.ByteTypeString}
	w := matchWeight(bt, 0, cfg)
	if w != 0 {
		t.Errorf("expected 0 for string with OnlyCode, got %f", w)
	}
}

func TestMatchWeight_OnlyCode_KeepsCode(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyCode = true
	bt := []byte{processor.ByteTypeCode}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightCode {
		t.Errorf("expected %f for code with OnlyCode, got %f", cfg.WeightCode, w)
	}
}

func TestMatchWeight_OnlyComments_ZerosCode(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyComments = true
	bt := []byte{processor.ByteTypeCode}
	w := matchWeight(bt, 0, cfg)
	if w != 0 {
		t.Errorf("expected 0 for code with OnlyComments, got %f", w)
	}
}

func TestMatchWeight_OnlyComments_KeepsComments(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyComments = true
	bt := []byte{processor.ByteTypeComment}
	w := matchWeight(bt, 0, cfg)
	if w != cfg.WeightComment {
		t.Errorf("expected %f for comment with OnlyComments, got %f", cfg.WeightComment, w)
	}
}

func TestMatchWeight_OutOfBounds_FallbackToCode(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeComment}
	w := matchWeight(bt, 5, cfg) // offset 5 out of bounds for length 1
	if w != cfg.WeightCode {
		t.Errorf("expected %f (out of bounds fallback), got %f", cfg.WeightCode, w)
	}
}

func TestMatchWeight_NegativeOffset_FallbackToCode(t *testing.T) {
	cfg := DefaultStructuralConfig()
	bt := []byte{processor.ByteTypeComment}
	w := matchWeight(bt, -1, cfg)
	if w != cfg.WeightCode {
		t.Errorf("expected %f (negative offset fallback), got %f", cfg.WeightCode, w)
	}
}

// --- rankResultsStructural tests ---

func TestRankResultsStructural_EmptyResults(t *testing.T) {
	cfg := DefaultStructuralConfig()
	results := rankResultsStructural(100, nil, map[string]int{}, cfg)
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestRankResultsStructural_CodeMatchScoresHigherThanComment(t *testing.T) {
	cfg := DefaultStructuralConfig()

	// File with match in code
	codeByteType := make([]byte, 100)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}
	codeFile := &common.FileJob{
		Filename:        "code.go",
		Location:        "code.go",
		Content:         make([]byte, 100),
		ContentByteType: codeByteType,
		Bytes:           100,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}},
	}

	// File with match in comment
	commentByteType := make([]byte, 100)
	for i := range commentByteType {
		commentByteType[i] = processor.ByteTypeComment
	}
	commentFile := &common.FileJob{
		Filename:        "comment.go",
		Location:        "comment.go",
		Content:         make([]byte, 100),
		ContentByteType: commentByteType,
		Bytes:           100,
		MatchLocations:  map[string][][]int{"func": {{10, 14}}},
	}

	results := []*common.FileJob{codeFile, commentFile}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	if codeFile.Score <= commentFile.Score {
		t.Errorf("expected code match score (%f) > comment match score (%f)", codeFile.Score, commentFile.Score)
	}
}

func TestRankResultsStructural_OnlyCode_ZerosCommentOnlyFile(t *testing.T) {
	cfg := DefaultStructuralConfig()
	cfg.OnlyCode = true

	commentByteType := make([]byte, 100)
	for i := range commentByteType {
		commentByteType[i] = processor.ByteTypeComment
	}
	commentFile := &common.FileJob{
		Filename:        "comment.go",
		Location:        "comment.go",
		Content:         make([]byte, 100),
		ContentByteType: commentByteType,
		Bytes:           100,
		MatchLocations:  map[string][][]int{"TODO": {{5, 9}}},
	}

	results := []*common.FileJob{commentFile}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	if commentFile.Score != 0 {
		t.Errorf("expected score 0 for comment-only file with OnlyCode, got %f", commentFile.Score)
	}
}

func TestRankResultsStructural_NilByteType_StillScored(t *testing.T) {
	cfg := DefaultStructuralConfig()

	file := &common.FileJob{
		Filename:        "unknown.txt",
		Location:        "unknown.txt",
		Content:         make([]byte, 100),
		ContentByteType: nil, // unrecognised language
		Bytes:           100,
		MatchLocations:  map[string][][]int{"hello": {{0, 5}}},
	}

	results := []*common.FileJob{file}
	df := CalculateDocumentFrequency(results)
	rankResultsStructural(10, results, df, cfg)

	if file.Score <= 0 {
		t.Errorf("expected positive score for nil ContentByteType fallback, got %f", file.Score)
	}
}

// --- RankResults integration tests ---

func TestRankResults_StructuralCase(t *testing.T) {
	codeByteType := make([]byte, 50)
	for i := range codeByteType {
		codeByteType[i] = processor.ByteTypeCode
	}
	file := &common.FileJob{
		Filename:        "test.go",
		Location:        "test.go",
		Content:         make([]byte, 50),
		ContentByteType: codeByteType,
		Bytes:           50,
		MatchLocations:  map[string][][]int{"test": {{0, 4}}},
	}

	results := RankResults("structural", 10, []*common.FileJob{file}, nil, 0.0, 100.0, 1.0, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", results[0].Score)
	}
}

func TestRankResults_BM25WithNilStructuralCfg(t *testing.T) {
	file := &common.FileJob{
		Filename:       "test.go",
		Location:       "test.go",
		Content:        make([]byte, 50),
		Bytes:          50,
		MatchLocations: map[string][][]int{"test": {{0, 4}}},
	}

	results := RankResults("bm25", 10, []*common.FileJob{file}, nil, 0.0, 100.0, 1.0, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Score <= 0 {
		t.Errorf("expected positive score for bm25, got %f", results[0].Score)
	}
}

// --- rankResultsComplexityGravity tests ---

func TestComplexityGravity_ZeroStrength_NoOp(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Complexity: 10,
	}
	results := rankResultsComplexityGravity([]*common.FileJob{file}, 0.0)
	if results[0].Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged with zero strength, got %f", results[0].Score)
	}
}

func TestComplexityGravity_ZeroComplexity_NoChange(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Complexity: 0,
	}
	results := rankResultsComplexityGravity([]*common.FileJob{file}, 1.0)
	// ln(1+0) = 0, so boost = 0, score unchanged
	if results[0].Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged with zero complexity, got %f", results[0].Score)
	}
}

func TestComplexityGravity_PositiveComplexity_Boosts(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Complexity: 10,
	}
	original := file.Score
	rankResultsComplexityGravity([]*common.FileJob{file}, 1.0)
	if file.Score <= original {
		t.Errorf("expected score > %f with positive complexity, got %f", original, file.Score)
	}
}

func TestComplexityGravity_HigherComplexity_RanksHigher(t *testing.T) {
	low := &common.FileJob{
		Score:      5.0,
		Complexity: 2,
	}
	high := &common.FileJob{
		Score:      5.0,
		Complexity: 50,
	}
	rankResultsComplexityGravity([]*common.FileJob{low, high}, 1.5)
	if high.Score <= low.Score {
		t.Errorf("expected high complexity score (%f) > low complexity score (%f)", high.Score, low.Score)
	}
}

func TestComplexityGravity_ZeroScore_Skipped(t *testing.T) {
	file := &common.FileJob{
		Score:      0,
		Complexity: 100,
	}
	rankResultsComplexityGravity([]*common.FileJob{file}, 2.5)
	if file.Score != 0 {
		t.Errorf("expected score 0 to remain unchanged, got %f", file.Score)
	}
}

func TestComplexityGravity_IntegrationWithBM25(t *testing.T) {
	lowComplexity := &common.FileJob{
		Filename:       "simple.go",
		Location:       "simple.go",
		Content:        make([]byte, 100),
		Bytes:          100,
		Complexity:     1,
		MatchLocations: map[string][][]int{"test": {{0, 4}}},
	}
	highComplexity := &common.FileJob{
		Filename:       "complex.go",
		Location:       "complex.go",
		Content:        make([]byte, 100),
		Bytes:          100,
		Complexity:     50,
		MatchLocations: map[string][][]int{"test": {{0, 4}}},
	}

	results := RankResults("bm25", 10, []*common.FileJob{lowComplexity, highComplexity}, nil, 1.5, 100.0, 1.0, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// With gravity, the high complexity file should rank first
	if results[0].Complexity != 50 {
		t.Errorf("expected high complexity file ranked first, got complexity=%d", results[0].Complexity)
	}
}

// --- rankResultsNoisePenalty tests ---

func TestNoisePenalty_RawSensitivity_NoOp(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Bytes:      1_000_000,
		Complexity: 0,
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 100.0)
	if file.Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged with raw sensitivity, got %f", file.Score)
	}
}

func TestNoisePenalty_LargeFileZeroComplexity_Penalised(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Bytes:      1_000_000, // log10 = 6
		Complexity: 0,         // signal = 1/6 ≈ 0.167
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 1.0)
	if file.Score >= 5.0 {
		t.Errorf("expected score < 5.0 for large zero-complexity file, got %f", file.Score)
	}
	if file.Score <= 0 {
		t.Errorf("expected positive score, got %f", file.Score)
	}
}

func TestNoisePenalty_HighComplexitySmallFile_Untouched(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Bytes:      5000, // log10 ≈ 3.7
		Complexity: 20,   // signal = 21/3.7 ≈ 5.67, clamped to 1.0
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 1.0)
	if file.Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged for high-complexity file, got %f", file.Score)
	}
}

func TestNoisePenalty_ZeroScore_Skipped(t *testing.T) {
	file := &common.FileJob{
		Score:      0,
		Bytes:      1_000_000,
		Complexity: 0,
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 1.0)
	if file.Score != 0 {
		t.Errorf("expected score 0 unchanged, got %f", file.Score)
	}
}

func TestNoisePenalty_SilenceSeverePenalty(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Bytes:      100_000,
		Complexity: 0,
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 0.1) // silence
	// signal = 1/5 = 0.2, penalty = min(1.0, 0.2*0.1) = 0.02
	if file.Score >= 1.0 {
		t.Errorf("expected severe penalty with silence sensitivity, got %f", file.Score)
	}
}

func TestNoisePenalty_SmallBytesFloor(t *testing.T) {
	file := &common.FileJob{
		Score:      5.0,
		Bytes:      0, // should be floored to 10
		Complexity: 5,
	}
	rankResultsNoisePenalty([]*common.FileJob{file}, 1.0)
	// safeBytes = 10, log10(10) = 1, signal = 6/1 = 6, clamped to 1.0
	if file.Score != 5.0 {
		t.Errorf("expected score unchanged when bytes floored to 10, got %f", file.Score)
	}
}

func TestNoisePenalty_DataFileRankedBelowCodeFile(t *testing.T) {
	jsonFile := &common.FileJob{
		Score:      5.0,
		Bytes:      500_000,
		Complexity: 0,
	}
	goFile := &common.FileJob{
		Score:      5.0,
		Bytes:      5000,
		Complexity: 20,
	}
	rankResultsNoisePenalty([]*common.FileJob{jsonFile, goFile}, 1.0)
	if jsonFile.Score >= goFile.Score {
		t.Errorf("expected JSON file score (%f) < Go file score (%f)", jsonFile.Score, goFile.Score)
	}
}

// --- CalculateDocumentFrequency sanity check ---

func TestCalculateDocumentFrequency(t *testing.T) {
	results := []*common.FileJob{
		{MatchLocations: map[string][][]int{"a": {{0, 1}}, "b": {{2, 3}}}},
		{MatchLocations: map[string][][]int{"a": {{0, 1}}}},
	}

	df := CalculateDocumentFrequency(results)
	if df["a"] != 2 {
		t.Errorf("expected df[a]=2, got %d", df["a"])
	}
	if df["b"] != 1 {
		t.Errorf("expected df[b]=1, got %d", df["b"])
	}
}

// --- IsTestFile tests ---

func TestIsTestFile_GoTestFile(t *testing.T) {
	if !IsTestFile("foo_test.go") {
		t.Error("expected foo_test.go to be a test file")
	}
}

func TestIsTestFile_JsTestFile(t *testing.T) {
	if !IsTestFile("foo.test.js") {
		t.Error("expected foo.test.js to be a test file")
	}
}

func TestIsTestFile_SpecFile(t *testing.T) {
	if !IsTestFile("foo.spec.ts") {
		t.Error("expected foo.spec.ts to be a test file")
	}
}

func TestIsTestFile_TestPrefix(t *testing.T) {
	if !IsTestFile("test_helper.py") {
		t.Error("expected test_helper.py to be a test file")
	}
}

func TestIsTestFile_JavaSuffix(t *testing.T) {
	if !IsTestFile("UserServiceTest.java") {
		t.Error("expected UserServiceTest.java to be a test file")
	}
}

func TestIsTestFile_TestsDirectory(t *testing.T) {
	if !IsTestFile("src/tests/foo.go") {
		t.Error("expected src/tests/foo.go to be a test file")
	}
}

func TestIsTestFile_JestDirectory(t *testing.T) {
	if !IsTestFile("src/__tests__/foo.js") {
		t.Error("expected src/__tests__/foo.js to be a test file")
	}
}

func TestIsTestFile_RegularFile(t *testing.T) {
	if IsTestFile("src/auth.go") {
		t.Error("expected src/auth.go to NOT be a test file")
	}
}

func TestIsTestFile_CaseInsensitive(t *testing.T) {
	if !IsTestFile("Foo_Test.GO") {
		t.Error("expected Foo_Test.GO to be a test file")
	}
}

// --- HasTestIntent tests ---

func TestHasTestIntent_WithTestKeyword(t *testing.T) {
	if !HasTestIntent([]string{"auth", "test"}) {
		t.Error("expected test intent with 'test' keyword")
	}
}

func TestHasTestIntent_WithMockKeyword(t *testing.T) {
	if !HasTestIntent([]string{"service", "mock"}) {
		t.Error("expected test intent with 'mock' keyword")
	}
}

func TestHasTestIntent_CaseInsensitive(t *testing.T) {
	if !HasTestIntent([]string{"TEST"}) {
		t.Error("expected test intent with 'TEST' (case insensitive)")
	}
}

func TestHasTestIntent_NoIntentTerms(t *testing.T) {
	if HasTestIntent([]string{"auth", "login"}) {
		t.Error("expected no test intent for 'auth login'")
	}
}

func TestHasTestIntent_EmptyInput(t *testing.T) {
	if HasTestIntent([]string{}) {
		t.Error("expected no test intent for empty input")
	}
}

// --- rankResultsTestDampening tests ---

func TestTestDampening_NonTestFileUnchanged(t *testing.T) {
	file := &common.FileJob{
		Location: "src/auth.go",
		Score:    5.0,
	}
	rankResultsTestDampening([]*common.FileJob{file}, 0.4, false)
	if file.Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged for non-test file, got %f", file.Score)
	}
}

func TestTestDampening_TestFilePenalized(t *testing.T) {
	file := &common.FileJob{
		Location: "auth_test.go",
		Score:    5.0,
	}
	rankResultsTestDampening([]*common.FileJob{file}, 0.4, false)
	expected := 5.0 * 0.4
	if file.Score != expected {
		t.Errorf("expected score %f, got %f", expected, file.Score)
	}
}

func TestTestDampening_TestFileBoosted(t *testing.T) {
	file := &common.FileJob{
		Location: "auth_test.go",
		Score:    5.0,
	}
	rankResultsTestDampening([]*common.FileJob{file}, 0.4, true)
	expected := 5.0 * 1.5
	if file.Score != expected {
		t.Errorf("expected score %f, got %f", expected, file.Score)
	}
}

func TestTestDampening_ZeroScoreSkipped(t *testing.T) {
	file := &common.FileJob{
		Location: "auth_test.go",
		Score:    0,
	}
	rankResultsTestDampening([]*common.FileJob{file}, 0.4, false)
	if file.Score != 0 {
		t.Errorf("expected score 0 unchanged, got %f", file.Score)
	}
}

func TestTestDampening_NeutralPenaltyNoOp(t *testing.T) {
	file := &common.FileJob{
		Location: "auth_test.go",
		Score:    5.0,
	}
	rankResultsTestDampening([]*common.FileJob{file}, 1.0, false)
	if file.Score != 5.0 {
		t.Errorf("expected score 5.0 unchanged with neutral penalty, got %f", file.Score)
	}
}

func TestTestDampening_IntegrationWithBM25(t *testing.T) {
	implFile := &common.FileJob{
		Filename:       "auth.go",
		Location:       "auth.go",
		Content:        make([]byte, 100),
		Bytes:          100,
		Complexity:     10,
		MatchLocations: map[string][][]int{"auth": {{0, 4}}},
	}
	testFile := &common.FileJob{
		Filename:       "auth_test.go",
		Location:       "auth_test.go",
		Content:        make([]byte, 100),
		Bytes:          100,
		Complexity:     10,
		MatchLocations: map[string][][]int{"auth": {{0, 4}, {10, 14}, {20, 24}}},
	}

	results := RankResults("bm25", 10, []*common.FileJob{implFile, testFile}, nil, 0.0, 100.0, 0.4, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// With test dampening, the impl file should rank above the test file
	if results[0].Location != "auth.go" {
		t.Errorf("expected auth.go ranked first, got %s (scores: auth.go=%f, auth_test.go=%f)",
			results[0].Location, implFile.Score, testFile.Score)
	}
}
