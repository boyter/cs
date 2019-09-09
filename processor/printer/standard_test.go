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

func TestWriteColoredCheckMixedWords(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}
	loc["something"] = []int{10}

	got := WriteColored([]byte("this this something"), loc, "__", "__")

	expected := "__this__ __this__ __something__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCaseCheck(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["t"] = []int{0}

	got := WriteColored([]byte("THIS"), loc, "__", "__")

	expected := "__THIS__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredOverlapStart(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["t"] = []int{0}

	got := WriteColored([]byte("this"), loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredOverlapMiddle(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["h"] = []int{1}

	got := WriteColored([]byte("this"), loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}
