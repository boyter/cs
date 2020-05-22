package processor

import (
	"crypto/md5"
	"encoding/hex"
	str "github.com/boyter/cs/string"
)

// For fuzz testing...
// https://github.com/dvyukov/go-fuzz
// install both go-fuzz-build and go-fuzz
// go-fuzz-build && go-fuzz
func Fuzz(data []byte) int {

	md5_d := md5.New()
	find := hex.EncodeToString(md5_d.Sum(data))

	loc := map[string][][]int{}
	loc[find[:2]] = str.IndexAllIgnoreCaseUnicode(string(data), find[:2], -1)

	freq := map[string]int{}
	freq[find[:2]] = 5

	res := &fileJob{
		Content:        data,
		MatchLocations: loc,
	}

	extractRelevantV3(res, freq, 300, "...")

	findSpaceRight(string(data), 0, 10000)
	findSpaceRight(string(data), len(data) - 1, 10000)

	return 1
}

