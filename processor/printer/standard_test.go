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

func TestWriteColoredCheck2(t *testing.T) {
	loc := map[string][]int{}
	loc["bing"] = []int{0}

	got := WriteColored([]byte("bing"), loc, "__", "__")

	expected := "__bing__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheckTwoWords(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}

	got := WriteColored([]byte("this this"), loc, "__", "__")

	expected := "__this__ __this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}