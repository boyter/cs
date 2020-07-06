// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package processor

import (
	str "github.com/boyter/cs/str"
	"strings"
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
		pos, found := findSpaceRight(&FileJob{Content: []byte(c.s)}, c.startpos, c.distance)

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
		pos, found := findSpaceLeft(&FileJob{Content: []byte(c.s)}, c.startpos, c.distance)

		if pos != c.want {
			t.Error("  pos for", i, "wanted", c.want, "got", pos)
		}

		if found != c.found {
			t.Error("found for", i, "wanted", c.found, "got", found)
		}
	}
}

func TestExtractRelevantV3PaintedShip(t *testing.T) {
	terms := []string{
		"painted",
		"ship",
		"ocean",
	}

	res := &FileJob{
		Content:        []byte(rhymeOfTheAncient),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(rhymeOfTheAncient, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `Day after day, day after day,
We stuck, nor breath nor motion;
As idle as a painted ship
Upon a painted ocean.`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3WaterWaterEverywhere(t *testing.T) {
	terms := []string{
		"water",
		"every",
		"where",
		"drink",
	}

	res := &FileJob{
		Content:        []byte(rhymeOfTheAncient),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(rhymeOfTheAncient, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `Water, water, every where,
And all the boards did shrink;
Water, water, every where,
Nor any drop to drink.`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3GroanedDead(t *testing.T) {
	terms := []string{
		"groaned",
		"dead",
	}

	res := &FileJob{
		Content:        []byte(rhymeOfTheAncient),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(rhymeOfTheAncient, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `They groaned, they stirred, they all uprose,
Nor spake, nor moved their eyes;
It had been strange, even in a dream,
To have seen those dead men rise.`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3DeathFires(t *testing.T) {
	terms := []string{
		"death",
		"fires",
	}

	res := &FileJob{
		Content:        []byte(rhymeOfTheAncient),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(rhymeOfTheAncient, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `About, about, in reel and rout
The death-fires danced at night;
The water, like a witch's oils,
Burnt green, and blue and white.`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3PoorNerves(t *testing.T) {
	terms := []string{
		"poor",
		"nerves",
	}

	res := &FileJob{
		Content:        []byte(prideAndPrejudice),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(prideAndPrejudice, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `You take delight in vexing me. You have no compassion for my poor
      nerves`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3TenThousandAYear(t *testing.T) {
	terms := []string{
		"ten",
		"thousand",
		"a",
		"year",
	}

	res := &FileJob{
		Content:        []byte(prideAndPrejudice),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(prideAndPrejudice, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `of his having
      ten thousand a year. The gentlemen pronounced him to be a fine`) {
		t.Error("expected to have snippet")
	}
}

func TestExtractRelevantV3StrangerParents(t *testing.T) {
	terms := []string{
		"stranger",
		"parents",
	}

	res := &FileJob{
		Content:        []byte(prideAndPrejudice),
		MatchLocations: map[string][][]int{},
	}

	for _, t := range terms {
		res.MatchLocations[t] = str.IndexAllIgnoreCase(prideAndPrejudice, t, -1)
	}

	df := calculateDocumentTermFrequency([]*FileJob{res})
	snippets := extractRelevantV3(res, df, 300, "")

	if !strings.Contains(snippets[0].Content, `An unhappy alternative is before you, Elizabeth. From this day
      you must be a stranger to one of your parents. Your mother will
      never see you again if you`) {
		t.Error("expected to have snippet")
	}
}
