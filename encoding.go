// SPDX-License-Identifier: MIT

package main

import (
	"unicode/utf16"
	"unicode/utf8"
)

// Encodings reported for files that are not UTF-8. A file reported with one of
// these is decoded to UTF-8 before it is searched, so every line number and
// byte offset cs returns for it refers to the decoded text, not the bytes on disk.
const (
	EncodingUTF16LE = "utf-16le"
	EncodingUTF16BE = "utf-16be"
)

// decodeText returns raw as UTF-8 text plus the encoding it was decoded from.
// Only UTF-16 that starts with a byte order mark is decoded; the mark is
// dropped. Anything else, UTF-8 with or without a BOM included, is returned
// as it is with an empty encoding, after a check of its first two bytes only.
//
// Decoded text is appended to dst[:0], so a caller can reuse one buffer
// across files. A trailing odd byte, as left when a read limit cuts a file
// in the middle of a code unit, is dropped.
func decodeText(raw []byte, dst []byte) ([]byte, string) {
	if len(raw) < 2 {
		return raw, ""
	}
	var bigEndian bool
	var enc string
	switch {
	case raw[0] == 0xFF && raw[1] == 0xFE:
		enc = EncodingUTF16LE
	case raw[0] == 0xFE && raw[1] == 0xFF:
		bigEndian, enc = true, EncodingUTF16BE
	default:
		return raw, ""
	}

	body := raw[2:]
	n := len(body) / 2
	unit := func(i int) rune {
		if bigEndian {
			return rune(body[2*i])<<8 | rune(body[2*i+1])
		}
		return rune(body[2*i+1])<<8 | rune(body[2*i])
	}

	// Each code unit becomes at most three bytes of UTF-8; a surrogate pair
	// is two units and four bytes.
	out := dst[:0]
	if cap(out) < n*3 {
		out = make([]byte, 0, n*3)
	}
	for i := 0; i < n; i++ {
		r := unit(i)
		if utf16.IsSurrogate(r) && i+1 < n {
			if dec := utf16.DecodeRune(r, unit(i+1)); dec != utf8.RuneError {
				out = utf8.AppendRune(out, dec)
				i++
				continue
			}
		}
		// A lone surrogate is not a valid rune, so AppendRune writes U+FFFD.
		out = utf8.AppendRune(out, r)
	}
	return out, enc
}
