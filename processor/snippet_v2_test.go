package processor

import "testing"

func TestExtractRelevantV2(t *testing.T) {
	locations := [][]int{}
	locations = append(locations, []int{0, 4})
	locations = append(locations, []int{5, 7})
	locations = append(locations, []int{53, 55})

	fulltext := "this is some text (╯°□°）╯︵ ┻━┻) the thing we want is here"
	snippet := extractRelevantV2(fulltext, locations, 30, "...")

	if len(snippet.Content) == 0 || snippet.StartPos == 0 || snippet.EndPos == len(fulltext) {
		t.Error("Expected some value")
	}
}
