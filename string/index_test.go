// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package string

import (
	"math"
	"regexp"
	"strings"
	"testing"
)

func TestExtractLocations(t *testing.T) {
	locations := IndexAll("test that this returns a match", "test", math.MaxInt64)

	if locations[0][0] != 0 {
		t.Error("Expected to find location 0")
	}
}

func TestExtractLocationsLarge(t *testing.T) {
	locations := IndexAll("1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890 test that this returns a match", "test", math.MaxInt64)

	if locations[0][0] != 101 {
		t.Error("Expected to find location 101")
	}
}

func TestExtractLocationsLimit(t *testing.T) {
	locations := IndexAll("test test", "test", 1)

	if len(locations) != 1 {
		t.Error("Expected to find a single location")
	}
}

func TestExtractLocationsLimitTwo(t *testing.T) {
	locations := IndexAll("test test test", "test", 2)

	if len(locations) != 2 {
		t.Error("Expected to find two locations")
	}
}

func TestExtractLocationsLimitThree(t *testing.T) {
	locations := IndexAll("test test test", "test", 3)

	if len(locations) != 3 {
		t.Error("Expected to find three locations")
	}
}

func TestExtractLocationsNegativeLimit(t *testing.T) {
	locations := IndexAll("test test test", "test", -1)

	if len(locations) != 3 {
		t.Error("Expected to find three locations")
	}
}

func TestDropInReplacement(t *testing.T) {
	r := regexp.MustCompile(`test`)

	matches1 := r.FindAllIndex([]byte(test_MatchEndCase), -1)
	matches2 := IndexAll(test_MatchEndCase, "test", -1)

	for i := 0; i < len(matches1); i++ {
		if matches1[i][0] != matches2[i][0] || matches1[i][1] != matches2[i][1] {
			t.Error("Expect results to match", i)
		}
	}
}

func TestDropInReplacementNil(t *testing.T) {
	r := regexp.MustCompile(`test`)

	matches1 := r.FindAllIndex([]byte(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`), -1)
	matches2 := IndexAll(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, "test", -1)

	if matches1 != nil || matches2 != nil {
		t.Error("Expect results to be nil")
	}
}

func TestDropInReplacementMultiple(t *testing.T) {
	r := regexp.MustCompile(`1`)

	matches1 := r.FindAllIndex([]byte(`111`), -1)
	matches2 := IndexAll(`111`, "1", -1)

	for i := 0; i < len(matches1); i++ {
		if matches1[i][0] != matches2[i][0] || matches1[i][1] != matches2[i][1] {
			t.Error("Expect results to match", i)
		}
	}
}

func TestIndexAllIgnoreCaseUnicodeLongNeedleNoMatch(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("aaaaabbbbb", "aaaaaa", -1)

	if matches != nil {
		t.Error("Expected no matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeLongNeedleSingleMatch(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("aaaaaabbbbb", "aaaaaa", -1)

	if len(matches) != 1 {
		t.Error("Expected single matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeSingleMatch(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("aaaa", "a", 1)

	if len(matches) != 1 {
		t.Error("Expected single match")
	}
}

func TestIndexAllIgnoreCaseUnicodeTwoMatch(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("aaaa", "a", 2)

	if len(matches) != 2 {
		t.Error("Expected two matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeNegativeLimit(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("aaaa", "a", -1)

	if len(matches) != 4 {
		t.Error("Expected four matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeOutOfRange(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("veryuni", "unique", -1)

	if len(matches) != 0 {
		t.Error("Expected zero matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeOutOfRange2(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("veryuni", "uniq", -1)

	if len(matches) != 0 {
		t.Error("Expected zero matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeOutOfRange3(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("ve", "ee", -1)

	if len(matches) != 0 {
		t.Error("Expected zero matches")
	}
}

func TestIndexAllIgnoreCaseUnicodeCheck(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("a secret a", "ſecret", -1)

	if matches[0][0] != 2 || matches[0][1] != 8 {
		t.Error("Expected 2 and 8 got", matches[0][0], "and", matches[0][1])
	}

	if "a secret a"[matches[0][0]:matches[0][1]] != "secret" {
		t.Error("Expected secret")
	}
}

func TestIndexAllIgnoreCaseUnicodeCheckEnd(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode("a ſecret a", "secret", -1)

	if matches[0][0] != 2 || matches[0][1] != 9 {
		t.Error("Expected 2 and 7 got", matches[0][0], "and", matches[0][1])
	}

	if "a ſecret a"[matches[0][0]:matches[0][1]] != "ſecret" {
		t.Error("Expected ſecret")
	}
}

func TestDropInReplacementMultipleIndexAllIgnoreCaseUnicode(t *testing.T) {
	r := regexp.MustCompile(`1`)

	matches1 := r.FindAllIndex([]byte(`111`), -1)
	matches2 := IndexAllIgnoreCaseUnicode(`111`, "1", -1)

	for i := 0; i < len(matches1); i++ {
		if matches1[i][0] != matches2[i][0] || matches1[i][1] != matches2[i][1] {
			t.Error("Expect results to match", i)
		}
	}
}

func TestIndexAllIgnoreCaseUnicodeSpace(t *testing.T) {
	matches := IndexAllIgnoreCaseUnicode(prideAndPrejudice, "ten thousand a year", -1)
	m := IndexAll(strings.ToLower(prideAndPrejudice), "ten thousand a year", -1)

	r := regexp.MustCompile(`(?i)ten thousand a year`)
	index := r.FindAllIndex([]byte(prideAndPrejudice), -1)

	if len(matches) != len(m) || len(matches) != len(index) {
		t.Error("Expected 2 got", len(matches))
	}
}
