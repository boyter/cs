// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	"math/rand"
	"testing"
)

func TestFindSpaceRight(t *testing.T) {
	var cases = []struct {
		s        string
		startpos int
		distance int
		want     int
		found    bool
	}{
		{"yo", 0, 10, 0, false},
		{"boyterwasheredoingstuff", 0, 10, 0, false},
		{"", 0, 10, 0, false},
		{"", -16, 10, -16, false},
		{"", 50, 10, 50, false},
		{"a", 1, 10, 1, false},
		{"a", 2, 10, 2, false},
		{"aa", 0, 10, 0, false},
		{"a ", 0, 10, 1, true},
		{"aa ", 0, 10, 2, true},
		{"🍺 ", 0, 10, 4, true},
		{"🍺🍺 ", 0, 10, 8, true},
		{"aaaaaaaaaaa ", 0, 10, 0, false},
		{"aaaa ", 0, 3, 0, false},
		{"，", 0, 10, 0, false},
		{"“啊，公爵，热那", 0, 10, 0, false},
		{"aaaaa aaaaa", 5, 10, 5, true},
		{" aaaa aaaaa", 5, 10, 5, true},
		{"    a aaaaa", 5, 10, 5, true},
		{"     aaaaaa", 5, 10, 5, false},
		{"aaaaaaaaaa", 9, 10, 9, false},
	}

	for i, c := range cases {
		pos, found := findSpaceRight(&fileJob{Content: []byte(c.s)}, c.startpos, c.distance)

		if pos != c.want {
			t.Error("  pos for", i, "wanted", c.want, "got", pos)
		}

		if found != c.found {
			t.Error("found for", i, "wanted", c.found, "got", found)
		}
	}
}

func TestFindSpaceLeft(t *testing.T) {
	var cases = []struct {
		s        string
		startpos int
		distance int
		want     int
		found    bool
	}{
		{"yo", 1, 10, 1, false},
		{"boyterwasheredoingstuff", 10, 10, 10, false},
		{"", 10, 10, 10, false},
		{"aaaa", 3, 10, 3, false},
		{" aaaa", 4, 10, 0, true},
		{"aaaabaaaa", 4, 10, 4, false},
		{" 🍺", 4, 10, 0, true},
		{" 🍺🍺", 6, 10, 0, true},
		{" “啊，公爵，热那", 24, 100, 0, true},
		{" “啊，公爵，热那", 50, 100, 50, false},
		{"     aaaaaa", 5, 10, 4, true},
		{"     aaaaaa", 10, 10, 4, true},
	}

	for i, c := range cases {
		pos, found := findSpaceLeft(&fileJob{Content: []byte(c.s)}, c.startpos, c.distance)

		if pos != c.want {
			t.Error("  pos for", i, "wanted", c.want, "got", pos)
		}

		if found != c.found {
			t.Error("found for", i, "wanted", c.found, "got", found)
		}
	}
}

type spaceFinderCase = struct {
	text           string
	startPos, want int
	found          bool
}

var leftCases = []spaceFinderCase{
	{" aaaa", 4, 0, true},
	{" aaaa", 24, 0, true}, // large position should reset to len(string)
	{" aaaa", -1, 0, true}, // small position should reset to len(string)
	{"a aaa", 4, 1, true},
	{"aaaa ", 4, 4, true},
	{" 12345678901", 11, 11 - SnipSideMax, false}, // Space after SNIP_SIDE_MAX
	// 24 bytes. Searches from far right. N.B those 'spaces' are actually
	// code-points that include the comma
	{"“啊，公爵，热那", -1, 15, false},
	// Start just far enough away to not hit byte 0, make sure it counts back
	// to nearest whole rune
	{"“啊，公爵，热那", 11, 3, false},
	{"", -1, 0, false},  // position should reset to len(string)
	{"，", 11, 0, false}, // Single 3byte rune
	//Only contains continuation bytes.
	{"\x82\x83", -1, 1, false}, // position should reset to len(string)
}

var rightCases = []spaceFinderCase{
	{" aaaa", 0, 0, true},
	{" aaaa", 24, 0, true}, // large position should reset to 0
	{" aaaa", -1, 0, true}, // small position should reset to 0
	{"a aaa", 0, 1, true},
	{"abcd ", 0, 4, true},
	{"01234567890 ", 0, SnipSideMax, false}, // Space after SNIP_SIDE_MAX
	// 24 bytes. Searches from far left. N.B those 'spaces' are actually
	// code-points that include the comma
	{"“啊，公爵，热那", -1, 9, false}, // Goes to 10 then count back 1
	// Start just far enough away to not hit byte 0, make sure it counts back
	// to nearest whole rune
	{"“啊，公爵，热那", 13, 21, false},
	{"", -1, 0, false},  // position should reset to 0 when searching right
	{"，", 11, 0, false}, // Single 3byte rune
	// Only contains continuation bytes.
	{"\x82\x83", 0, 0, false}, // position should reset to 0 when searching right
}

func findNearbySpaceGenericTester(t *testing.T, cases []spaceFinderCase, finder func(*fileJob, int, int) (int, bool)) {
	for _, testCase := range cases {
		idx, found := finder(&fileJob{Content: []byte(testCase.text)}, testCase.startPos, SnipSideMax)

		if idx != testCase.want {
			t.Error("Expected", testCase.want, "got", idx)
		}
		if found != testCase.found {
			t.Error("Expected", testCase.found, "got", found)
		}
	}
}

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890~!@#$%^&*()_+{}|:<>?                        "

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
