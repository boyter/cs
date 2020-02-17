// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package string

import "testing"

func TestWriteColoredSimple(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	got := WriteHighlights("this", loc, "[red]", "[white]")

	expected := "[red]this[white]"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredTermSimple(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	got := WriteHighlights("this", loc, "\033[1;31m", "\033[0m")

	expected := "\033[1;31mthis\033[0m"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheckInOut(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}

	got := WriteHighlights("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheck2(t *testing.T) {
	loc := map[string][]int{}
	loc["bing"] = []int{0}

	got := WriteHighlights("bing", loc, "__", "__")

	expected := "__bing__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheckTwoWords(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}

	got := WriteHighlights("this this", loc, "__", "__")

	expected := "__this__ __this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCheckMixedWords(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0, 5}
	loc["something"] = []int{10}

	got := WriteHighlights("this this something", loc, "__", "__")

	expected := "__this__ __this__ __something__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredCaseCheck(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["t"] = []int{0}

	got := WriteHighlights("THIS", loc, "__", "__")

	expected := "__THIS__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredOverlapStart(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["t"] = []int{0}

	got := WriteHighlights("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredOverlapMiddle(t *testing.T) {
	loc := map[string][]int{}
	loc["this"] = []int{0}
	loc["h"] = []int{1}

	got := WriteHighlights("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestWriteColoredOverlapMiddleLonger(t *testing.T) {
	loc := map[string][]int{}
	loc["th"] = []int{0}
	loc["his"] = []int{1}

	got := WriteHighlights("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestBugOne(t *testing.T) {
	loc := map[string][]int{}
	loc["expected"] = []int{10}

	got := WriteHighlights("this is unexpected", loc, "__", "__")

	expected := "this is un__expected__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestBugTwo(t *testing.T) {
	loc := map[string][]int{}
	loc["got"] = []int{22, 71, 77}
	loc["expected"] = []int{0, 29}

	got := WriteHighlights(`expected := "this" if got != expected { t.Error("Expected", expected, "got", got)}`, loc, "[red]", "[white]")

	expected := `[red]expected[white] := "this" if [red]got[white] != [red]expected[white] { t.Error("Expected", expected, "[red]got[white]", [red]got[white])}`
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestBugThree(t *testing.T) {
	loc := map[string][]int{}
	loc[`"`] = []int{5, 8}
	loc[`",`] = []int{8}

	got := WriteHighlights(`Use: "cs",`, loc, "[red]", "[white]")

	expected := `Use: [red]"[white]cs[red]",[white]`
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

