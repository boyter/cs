package snippet

import (
	"testing"
)

func TestExtractLocationZero(t *testing.T) {
	loc := determineSnipLocations([]int{90, 110, 140}, 50)

	t.Error("Something", loc)
}
