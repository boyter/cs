package processor

import (
	"fmt"
	"math/rand"
	"testing"
)

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

var rightCases = []spaceFinderCase{
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

// Try to find bugs by fuzzing the input to all sorts of random things
func TestFindNearbySpaceFuzzy(t *testing.T) {
	for i := 0; i < 100000; i++ {
		// Opt1 - Fails!! - Try running this instead of the way bellow.  ATTN Boyter
		//findSpaceLeft(&fileJob{Content: []byte(randStringBytes(1000))}, rand.Intn(1000), rand.Intn(10000))
		//findSpaceRight(&fileJob{Content: []byte(randStringBytes(1000))}, rand.Intn(1000), rand.Intn(10000))

		// Opt2 - Passes
		bs := []byte(randStringBytes(1000))
		startPos := rand.Intn(1000)
		dist := rand.Intn(10000)
		findSpaceLeft(&fileJob{Content: bs}, startPos, dist)
		findSpaceRight(&fileJob{Content: bs}, startPos, dist)
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

/* Invocation below produces non UTF-8 byte at begining of the returned string. The highlight also
doesn't terminate

(general) asset ❯ cs --hidden --no-gitignore --no-ignore 是波
corpus/snippet/chinese_war_and_peace.txt (3.296)
�

英文

“啊，公爵，热那亚和卢加现在是波拿巴家族的领地，不过，我得事先对您说，如果您不对我说我们这里处于战争状态，如果您 还敢袒护这个基督的

(general) asset ❯

*/
var tempCase = spaceFinderCase{`第一章

英文

“啊，公爵，热那亚和卢加现在是波拿巴家族的领地，不过，我得事先对您说，如果您不对我说我们这里处于战争状态，如果您还敢袒护这个基督的敌人（我确乎相信，他是一个基督的敌人）的种种卑劣行径和他一手造成的灾祸，那么我就不再管您了。您`, 65, 10, true}

func TestTims(t *testing.T) {
	fmt.Println("Starting test main ...")

	var left_idx, right_idx int
	var found bool
	l := tempCase
	bs := []byte(l.text)
	distance := 35
	//fmt.Println("Before space search:", string(bs))

	fmt.Println()
	fmt.Printf("LOOKING LEFT from byte %d (← %d)...\n", l.startPos, distance)
	left_idx, found = findSpaceLeft(&fileJob{Content: bs}, l.startPos, distance)
	fmt.Println("Space found: ", found, ". Returned index:", left_idx)
	//fmt.Println(string(bs[idx:l.startPos]))

	fmt.Println()
	fmt.Printf("LOOKING RIGHT from byte %d (→ %d)...\n", l.startPos, distance)
	right_idx, found = findSpaceRight(&fileJob{Content: bs}, l.startPos, distance)
	fmt.Println("Space found: ", found, ". Returned index:", right_idx)
	//fmt.Println(string(bs[l.startPos:idx]))

	fmt.Println()
	fmt.Println("COMPLETE String is:")
	fmt.Println(string(bs[left_idx:right_idx]))
}
