package printer

import "testing"

func TestWriteColoredSimple(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	got := WriteColored([]byte("this"), loc, "[red]", "[white]")

	expected := "[red]this[white]"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheckInOut(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	got := WriteColored([]byte("this"), loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}
