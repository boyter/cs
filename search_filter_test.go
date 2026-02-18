// SPDX-License-Identifier: MIT

package main

import (
	"testing"

	"github.com/boyter/scc/v3/processor"
)

func TestFilterMatchLocations_OnlyCode_KeepsCodeMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeCode
	}
	locs := map[string][][]int{"func": {{0, 4}, {10, 14}}}
	cfg := &Config{OnlyCode: true}

	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["func"]) != 2 {
		t.Errorf("expected 2 locations, got %d", len(filtered["func"]))
	}
}

func TestFilterMatchLocations_OnlyCode_RemovesCommentMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeComment
	}
	locs := map[string][][]int{"TODO": {{0, 4}}}
	cfg := &Config{OnlyCode: true}

	_, ok := filterMatchLocations(locs, bt, cfg)
	if ok {
		t.Fatal("expected no matches to survive")
	}
}

func TestFilterMatchLocations_OnlyComments_KeepsCommentMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeComment
	}
	locs := map[string][][]int{"TODO": {{0, 4}}}
	cfg := &Config{OnlyComments: true}

	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["TODO"]) != 1 {
		t.Errorf("expected 1 location, got %d", len(filtered["TODO"]))
	}
}

func TestFilterMatchLocations_OnlyComments_RemovesCodeMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeCode
	}
	locs := map[string][][]int{"func": {{0, 4}}}
	cfg := &Config{OnlyComments: true}

	_, ok := filterMatchLocations(locs, bt, cfg)
	if ok {
		t.Fatal("expected no matches to survive")
	}
}

func TestFilterMatchLocations_OnlyStrings_KeepsStringMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeString
	}
	locs := map[string][][]int{"hello": {{0, 5}}}
	cfg := &Config{OnlyStrings: true}

	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["hello"]) != 1 {
		t.Errorf("expected 1 location, got %d", len(filtered["hello"]))
	}
}

func TestFilterMatchLocations_OnlyStrings_RemovesCodeMatches(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeCode
	}
	locs := map[string][][]int{"func": {{0, 4}}}
	cfg := &Config{OnlyStrings: true}

	_, ok := filterMatchLocations(locs, bt, cfg)
	if ok {
		t.Fatal("expected no matches to survive")
	}
}

func TestFilterMatchLocations_MixedContent_FiltersCorrectly(t *testing.T) {
	bt := make([]byte, 20)
	// Bytes 0-4: code, 5-9: comment, 10-14: string, 15-19: code
	for i := 0; i < 5; i++ {
		bt[i] = processor.ByteTypeCode
	}
	for i := 5; i < 10; i++ {
		bt[i] = processor.ByteTypeComment
	}
	for i := 10; i < 15; i++ {
		bt[i] = processor.ByteTypeString
	}
	for i := 15; i < 20; i++ {
		bt[i] = processor.ByteTypeCode
	}

	locs := map[string][][]int{
		"match": {{0, 4}, {5, 9}, {10, 14}, {15, 19}},
	}

	// OnlyCode should keep code matches (at 0 and 15)
	cfg := &Config{OnlyCode: true}
	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["match"]) != 2 {
		t.Errorf("expected 2 code locations, got %d", len(filtered["match"]))
	}

	// OnlyComments should keep comment match (at 5)
	cfg = &Config{OnlyComments: true}
	filtered, ok = filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["match"]) != 1 {
		t.Errorf("expected 1 comment location, got %d", len(filtered["match"]))
	}

	// OnlyStrings should keep string match (at 10)
	cfg = &Config{OnlyStrings: true}
	filtered, ok = filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected matches to survive")
	}
	if len(filtered["match"]) != 1 {
		t.Errorf("expected 1 string location, got %d", len(filtered["match"]))
	}
}

func TestFilterMatchLocations_NilByteType_PassesThrough(t *testing.T) {
	locs := map[string][][]int{"func": {{0, 4}}}
	cfg := &Config{OnlyCode: true}

	filtered, ok := filterMatchLocations(locs, nil, cfg)
	if !ok {
		t.Fatal("expected matches to pass through with nil byte type")
	}
	if len(filtered["func"]) != 1 {
		t.Errorf("expected 1 location, got %d", len(filtered["func"]))
	}
}

func TestFilterMatchLocations_AllFilteredOut_ReturnsFalse(t *testing.T) {
	bt := make([]byte, 20)
	for i := range bt {
		bt[i] = processor.ByteTypeString
	}
	locs := map[string][][]int{"hello": {{0, 5}}}
	cfg := &Config{OnlyCode: true}

	_, ok := filterMatchLocations(locs, bt, cfg)
	if ok {
		t.Fatal("expected no matches to survive when all in wrong type")
	}
}

func TestFilterMatchLocations_OnlyCode_KeepsBlankAsCode(t *testing.T) {
	bt := make([]byte, 10)
	for i := range bt {
		bt[i] = processor.ByteTypeBlank
	}
	locs := map[string][][]int{"x": {{0, 1}}}
	cfg := &Config{OnlyCode: true}

	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected blank-region matches to pass through with OnlyCode")
	}
	if len(filtered["x"]) != 1 {
		t.Errorf("expected 1 location, got %d", len(filtered["x"]))
	}
}

func TestFilterMatchLocations_MultipleTerms_PartialFilter(t *testing.T) {
	bt := make([]byte, 20)
	for i := 0; i < 10; i++ {
		bt[i] = processor.ByteTypeCode
	}
	for i := 10; i < 20; i++ {
		bt[i] = processor.ByteTypeComment
	}

	locs := map[string][][]int{
		"code_term":    {{0, 4}},
		"comment_term": {{10, 14}},
	}
	cfg := &Config{OnlyCode: true}

	filtered, ok := filterMatchLocations(locs, bt, cfg)
	if !ok {
		t.Fatal("expected at least one term to survive")
	}
	if _, exists := filtered["code_term"]; !exists {
		t.Error("expected code_term to survive")
	}
	if _, exists := filtered["comment_term"]; exists {
		t.Error("expected comment_term to be filtered out")
	}
}
