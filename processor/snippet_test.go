package processor

import (
	"math/rand"
	"testing"
)

type spaceFinderCase = struct {
	text           string
	startPos, want int
	found          bool
}

var leftCases []spaceFinderCase = []spaceFinderCase{
	{" aaaa", 4, 0, true},
	{" aaaa", 24, 0, true}, // large position should reset to len(string)
	{" aaaa", -1, 0, true}, // small position should reset to len(string)
	{"a aaa", 4, 1, true},
	{"aaaa ", 4, 4, true},
	{" 12345678901", 11, 11 - SNIP_SIDE_MAX, false}, // Space after SNIP_SIDE_MAX
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

var rightCases []spaceFinderCase = []spaceFinderCase{
	{" aaaa", 0, 0, true},
	{" aaaa", 24, 0, true}, // large position should reset to 0
	{" aaaa", -1, 0, true}, // small position should reset to 0
	{"a aaa", 0, 1, true},
	{"abcd ", 0, 4, true},
	{"01234567890 ", 0, SNIP_SIDE_MAX, false}, // Space after SNIP_SIDE_MAX
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
		idx, found := finder(&fileJob{Content: []byte(testCase.text)}, testCase.startPos, SNIP_SIDE_MAX)

		if idx != testCase.want {
			t.Error("Expected", testCase.want, "got", idx)
		}
		if found != testCase.found {
			t.Error("Expected", testCase.found, "got", found)
		}
	}
}

func TestFindNearbySpaceLeftX(t *testing.T) {
	findNearbySpaceGenericTester(t, leftCases, findSpaceLeft)
}

func TestFindNearbySpaceRightX(t *testing.T) {
	findNearbySpaceGenericTester(t, rightCases, findSpaceRight)
}

// TODO: Check all callers of `findSpaceLeft/Right` and make sure they can
// deal with new return value (it used to return pos to indicate no match)
func TestFindNearbySpaceLeftChangedTrue(t *testing.T) {
	space, changed := findSpaceLeft(&fileJob{Content: []byte(` aaaa`)}, 4, SNIP_SIDE_MAX)

	if space != 0 || changed != true {
		t.Error("Expected 0 and changed true got", space, changed)
	}
}

func TestFindNearbySpaceLeftChangedFalse(t *testing.T) {
	space, changed := findSpaceLeft(&fileJob{Content: []byte(` aaaa`)}, 4, 1)

	if space != 4 || changed != false {
		t.Error("Expected 4 and changed true got", space, changed)
	}
}

func TestFindNearbySpaceRightChangedTrue(t *testing.T) {
	space, changed := findSpaceRight(&fileJob{Content: []byte(`aaaa `)}, 0, SNIP_SIDE_MAX)

	if space != 4 || changed != true {
		t.Error("Expected 4 and changed true got", space, changed)
	}
}

func TestFindNearbySpaceRightChangedFalse(t *testing.T) {
	space, changed := findSpaceRight(&fileJob{Content: []byte(`aaaa `)}, 0, 1)

	if space != 0 || changed != false {
		t.Error("Expected 0 and changed true got", space, changed)
	}
}

// Try to find bugs by fuzzing the input to all sorts of random things
func TestFindNearbySpaceFuzzy(t *testing.T) {
	for i := 0; i < 100000; i++ {
		findNearbySpace(&fileJob{Content: []byte(randStringBytes(1000))}, rand.Intn(1000), rand.Intn(10000))
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
