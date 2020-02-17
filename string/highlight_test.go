// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package string

import "testing"

func TestHighlightStringSimple(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})

	got := HighlightString("this", loc, "[in]", "[out]")

	expected := "[in]this[out]"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringCheckInOut(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})

	got := HighlightString("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringCheck2(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})

	got := HighlightString("bing", loc, "__", "__")

	expected := "__bing__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringCheckTwoWords(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})
	loc = append(loc, []int{5,4})

	got := HighlightString("this this", loc, "__", "__")

	expected := "__this__ __this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringCheckMixedWords(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})
	loc = append(loc, []int{5,4})
	loc = append(loc, []int{10,9})

	got := HighlightString("this this something", loc, "__", "__")

	expected := "__this__ __this__ __something__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringOverlapStart(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,1})
	loc = append(loc, []int{0,4})

	got := HighlightString("THIS", loc, "__", "__")

	expected := "__THIS__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringOverlapMiddle(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,4})
	loc = append(loc, []int{1,1})

	got := HighlightString("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringNoOverlapMiddleNextSame(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,1})
	loc = append(loc, []int{1,1})

	got := HighlightString("this", loc, "__", "__")

	expected := "__t____h__is"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHighlightStringOverlapMiddleLonger(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{0,2})
	loc = append(loc, []int{1,3})

	got := HighlightString("this", loc, "__", "__")

	expected := "__this__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}

func TestBugOne(t *testing.T) {
	loc := [][]int{}
	loc = append(loc, []int{10,8})

	got := HighlightString("this is unexpected", loc, "__", "__")

	expected := "this is un__expected__"
	if got != expected {
		t.Error("Expected", expected, "got", got)
	}
}
