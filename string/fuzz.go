package string

import (
	"crypto/md5"
	"encoding/hex"
)

// For fuzz testing...
// https://github.com/dvyukov/go-fuzz
// install both go-fuzz-build and go-fuzz
// go-fuzz-build && go-fuzz
func Fuzz(data []byte) int {

	md5_d := md5.New()
	find := hex.EncodeToString(md5_d.Sum(data))

	IndexAll(string(data), find[:2], -1)
	l := IndexAllIgnoreCaseUnicode(string(data), find[:2], -1)
	HighlightString(string(data), l, "__IN__", "__OUT__")
	FindFirstSpaceRight(string(data), 0, 100000)
	FindFirstSpaceLeft(string(data), 10, 100000)
	return 1
}
