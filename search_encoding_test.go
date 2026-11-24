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
		// FF FE 00 00 is the UTF-32 LE mark, which stays out of scope.
		{"utf32 le bom", []byte{0xFF, 0xFE, 0, 0, 'a', 0, 0, 0}, "\xff\xfe\x00\x00a\x00\x00\x00", ""},
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

func runSearch(t *testing.T, cfg *Config, query string) ([]*common.FileJob, *SearchStats) {
	t.Helper()
	ch, stats, err := DoSearch(context.Background(), cfg, query, nil)
	if err != nil {
		t.Fatal(err)
	}
	var out []*common.FileJob
	for fj := range ch {
		out = append(out, fj)
	}
	return out, stats
}

// A file longer than --max-read-size-bytes is searched only up to the limit,
// and its result says so. UTF-16 is cut on raw bytes, before decoding.
func TestDoSearch_ReportsTruncatedFiles(t *testing.T) {
	dir := t.TempDir()
	long := "onair\n" + strings.Repeat("filler line\n", 50)
	files := map[string][]byte{
		"short.txt":      []byte("onair\n"),
		"long.txt":       []byte(long),
		"long_utf16.txt": encodeUTF16(long, binary.LittleEndian, true),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cfg := DefaultConfig()
	cfg.Directory = dir
	cfg.ResultLimit = -1
	cfg.MaxReadSizeBytes = 100
	got, _ := runSearch(t, &cfg, "onair")

	want := map[string]bool{"short.txt": false, "long.txt": true, "long_utf16.txt": true}
	seen := map[string]bool{}
	for _, fj := range got {
		seen[fj.Filename] = true
		if fj.Truncated != want[fj.Filename] {
			t.Errorf("%s: truncated %v, want %v", fj.Filename, fj.Truncated, want[fj.Filename])
		}
		if fj.Filename == "long_utf16.txt" && len(fj.Content) != 49 {
			// 100 raw bytes: a 2 byte BOM, then 49 code units of ASCII.
			t.Errorf("%s: %d decoded bytes, want 49", fj.Filename, len(fj.Content))
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s: no match, want a match", name)
		}
	}

	// A file exactly at the limit was read whole, so it is not truncated.
	exact := strings.Repeat("x", 94) + "onair\n"
	if err := os.WriteFile(filepath.Join(dir, "exact.txt"), []byte(exact), 0o600); err != nil {
		t.Fatal(err)
	}
	got, _ = runSearch(t, &cfg, "onair")
	for _, fj := range got {
		if fj.Filename == "exact.txt" && fj.Truncated {
			t.Errorf("exact.txt: truncated, want not truncated")
		}
	}
}

// Files that are read but not searched are counted by reason.
func TestDoSearch_CountsSkippedFiles(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		"text.sql":     []byte(encodingFixtureText),
		"le_nobom.sql": encodeUTF16(encodingFixtureText, binary.LittleEndian, false),
		"image.bin":    {0x89, 'P', 'N', 'G', 0, 0, 0, 0},
		"min.js":       []byte(strings.Repeat("a", 1000)),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unreadable := filepath.Join(dir, "locked.sql")
	if err := os.WriteFile(unreadable, []byte("onair"), 0o000); err != nil {
		t.Fatal(err)
	}
	wantUnreadable := int64(1)
	if f, err := os.Open(unreadable); err == nil {
		// Running as root: the file can be read after all.
		f.Close()
		wantUnreadable = 0
	}

	cfg := DefaultConfig()
	cfg.Directory = dir
	cfg.ResultLimit = -1
	cfg.IgnoreIgnoreFile = true
	_, stats := runSearch(t, &cfg, "onair")

	for _, c := range []struct {
		name      string
		got, want int64
	}{
		{"binary", stats.SkippedBinary.Load(), 2},
		{"minified", stats.SkippedMinified.Load(), 1},
		{"unreadable", stats.SkippedUnreadable.Load(), wantUnreadable},
		{"text", stats.TextFileCount.Load(), 1},
	} {
		if c.got != c.want {
			t.Errorf("%s: %d files, want %d", c.name, c.got, c.want)
		}
	}
}
