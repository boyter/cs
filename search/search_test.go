package search

import (
	"math"
	"testing"
)

func TestExtractLocations(t *testing.T) {
	locations := ExtractLocations("test", "test that this returns a match", math.MaxInt64)

	if locations[0] != 0 {
		t.Error("Expected to find location 0")
	}
}

func TestExtractLocationsLimit(t *testing.T) {
	locations := ExtractLocations("test", "test test", 1)

	if len(locations) != 1 {
		t.Error("Expected to find a single location")
	}
}

func TestExtractLocationsLimitTwo(t *testing.T) {
	locations := ExtractLocations("test", "test test test", 2)

	if len(locations) != 2 {
		t.Error("Expected to find two locations")
	}
}

func TestExtractLocationsLimitThree(t *testing.T) {
	locations := ExtractLocations("test", "test test test", 3)

	if len(locations) != 3 {
		t.Error("Expected to find three locations")
	}
}
