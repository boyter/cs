// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/boyter/cs/v3/pkg/common"
)

// encodingFixtureText is a SQL Server stored procedure line as SSMS saves it:
// CRLF line endings, with the search term on the first line.
const encodingFixtureText = "SELECT * FROM ONAIR.PROGRAM\r\nWHERE id = 1\r\n"

func encodeUTF16(s string, order binary.ByteOrder, bom bool) []byte {
	var buf bytes.Buffer
	if bom {
		_ = binary.Write(&buf, order, uint16(0xFEFF))
	}
	for _, u := range utf16.Encode([]rune(s)) {
		_ = binary.Write(&buf, order, u)
	}
	return buf.Bytes()
}

// writeEncodingFixtures writes the same text in each encoding the bug report
// names, and returns the directory that holds them.
func writeEncodingFixtures(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string][]byte{
		"utf8.sql":     []byte(encodingFixtureText),
		"utf8_bom.sql": append([]byte{0xEF, 0xBB, 0xBF}, encodingFixtureText...),
		"le_bom.sql":   encodeUTF16(encodingFixtureText, binary.LittleEndian, true),
		"be_bom.sql":   encodeUTF16(encodingFixtureText, binary.BigEndian, true),
		"le_nobom.sql": encodeUTF16(encodingFixtureText, binary.LittleEndian, false),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func searchEncodingFixtures(t *testing.T, includeBinary bool) map[string]*common.FileJob {
	t.Helper()
	initLanguageDatabase()
	cfg := DefaultConfig()
	cfg.Directory = writeEncodingFixtures(t)
	cfg.ResultLimit = -1
	cfg.IncludeBinaryFiles = includeBinary

	ch, _, err := DoSearch(context.Background(), &cfg, "onair", nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]*common.FileJob{}
	for fj := range ch {
		got[fj.Filename] = fj
		for term, locs := range fj.MatchLocations {
			for _, loc := range locs {
				// Match offsets must index the decoded text, so the bytes at
				// each location are the term itself.
				if m := string(fj.Content[loc[0]:loc[1]]); !strings.EqualFold(m, term) {
					t.Errorf("%s: location %v holds %q, want %q", fj.Filename, loc, m, term)
				}
				// The term is on the first line of every fixture.
				if line := bytes.Count(fj.Content[:loc[0]], []byte("\n")) + 1; line != 1 {
					t.Errorf("%s: match on line %d, want 1", fj.Filename, line)
				}
			}
		}
	}
	return got
}

// UTF-16 files with a byte order mark are decoded and searched; UTF-16 with no
// mark stays out of scope, as in ripgrep.
func TestDoSearch_UTF16WithBOM(t *testing.T) {
	for _, includeBinary := range []bool{false, true} {
		name := "default"
		if includeBinary {
			name = "binary"
		}
		t.Run(name, func(t *testing.T) {
			got := searchEncodingFixtures(t, includeBinary)

			for _, want := range []string{"utf8.sql", "utf8_bom.sql", "le_bom.sql", "be_bom.sql"} {
				if _, ok := got[want]; !ok {
					t.Errorf("%s: no match, want a match", want)
				}
			}

			// The decoded text, its offsets and its counts are those of the
			// UTF-8 file, and the result names the encoding it came from.
			ref, ok := got["utf8.sql"]
			if !ok {
				return
			}
			wantEnc := map[string]string{
				"utf8.sql":     "",
				"utf8_bom.sql": "",
				"le_bom.sql":   EncodingUTF16LE,
				"be_bom.sql":   EncodingUTF16BE,
			}
			for name, enc := range wantEnc {
				fj, ok := got[name]
				if !ok {
					continue
				}
				if fj.Encoding != enc {
					t.Errorf("%s: encoding %q, want %q", name, fj.Encoding, enc)
				}
				if fj.Language != "SQL" || fj.Lines != ref.Lines || fj.Code != ref.Code {
					t.Errorf("%s: %s %d lines %d code, want SQL %d lines %d code",
						name, fj.Language, fj.Lines, fj.Code, ref.Lines, ref.Code)
				}
				if enc == "" {
					continue
				}
				if !bytes.Equal(fj.Content, ref.Content) {
					t.Errorf("%s: content %q, want %q", name, fj.Content, ref.Content)
				}
				if !reflect.DeepEqual(fj.MatchLocations, ref.MatchLocations) {
					t.Errorf("%s: match locations %v, want %v", name, fj.MatchLocations, ref.MatchLocations)
				}
			}
		})
	}
}

func TestDecodeText(t *testing.T) {
	le := func(s string) []byte { return encodeUTF16(s, binary.LittleEndian, true) }
	be := func(s string) []byte { return encodeUTF16(s, binary.BigEndian, true) }

	cases := []struct {
		name    string
		raw     []byte
		want    string
		wantEnc string
	}{
		{"empty", nil, "", ""},
		{"one byte", []byte{0xFF}, "\xff", ""},
		{"utf8 unchanged", []byte("héllo"), "héllo", ""},
		{"utf8 bom kept", []byte("\xEF\xBB\xBFhi"), "\xEF\xBB\xBFhi", ""},
		{"le bom only", []byte{0xFF, 0xFE}, "", EncodingUTF16LE},
		{"le ascii", le("onair"), "onair", EncodingUTF16LE},
		{"be ascii", be("onair"), "onair", EncodingUTF16BE},
		{"le non-ascii", le("café ☃"), "café ☃", EncodingUTF16LE},
		{"le surrogate pair", le("a😀b"), "a😀b", EncodingUTF16LE},
		{"be surrogate pair", be("a😀b"), "a😀b", EncodingUTF16BE},
		// A read limit can cut a file in the middle of a code unit or a pair.
		{"le odd trailing byte", append(le("ab"), 'c'), "ab", EncodingUTF16LE},
		{"le cut surrogate pair", le("a😀")[:6], "a\uFFFD", EncodingUTF16LE},
		{"le lone low surrogate", []byte{0xFF, 0xFE, 0x00, 0xDC, 'x', 0}, "\uFFFDx", EncodingUTF16LE},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, enc := decodeText(c.raw, nil)
			if string(got) != c.want || enc != c.wantEnc {
				t.Errorf("got %q %q, want %q %q", got, enc, c.want, c.wantEnc)
			}
		})
	}
}

// --max-read-size-bytes limits the raw bytes read, before decoding.
func TestDoSearch_UTF16ReadLimitAppliesToRawBytes(t *testing.T) {
	dir := t.TempDir()
	// The BOM is raw bytes 0 to 2 and "x" bytes 2 to 4, so "ONAIR" ends at byte 14.
	raw := encodeUTF16("xONAIRx", binary.LittleEndian, true)
	if err := os.WriteFile(filepath.Join(dir, "p.sql"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		limit int64
		match bool
	}{{14, true}, {13, false}} {
		cfg := DefaultConfig()
		cfg.Directory = dir
		cfg.ResultLimit = -1
		cfg.MaxReadSizeBytes = tc.limit

		ch, _, err := DoSearch(context.Background(), &cfg, "onair", nil)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for range ch {
			n++
		}
		if (n == 1) != tc.match {
			t.Errorf("limit %d: %d matches, want match=%v", tc.limit, n, tc.match)
		}
	}
}
