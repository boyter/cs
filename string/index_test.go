package string

import (
	"math"
	"regexp"
	"testing"
)

func TestExtractLocations(t *testing.T) {
	locations := IndexAll("test that this returns a match", "test", math.MaxInt64)

	if locations[0][0] != 0 {
		t.Error("Expected to find location 0")
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
