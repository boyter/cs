package search

import (
	"bytes"
	"regexp/syntax"
	"strings"
	"testing"
	"unicode"
)

// The one rule: the prefilter never rules out content the regexp matches.
func TestThePrefilterNeverRulesOutAMatch(t *testing.T) {
	patterns := []string{
		`BRMP-[0-9]+`,
		`(AKIA[0-9A-Z]{16}|sk_live|-----BEGIN|xox[bp]-)`,
		`(secret|api[_-]?key|token)\s*[:=]\s*["'][A-Za-z0-9_\-]{12,}`,
		`^\s{8}- Name: `,
		`20(25|26)-[01][0-9]-[0-3][0-9]`,
		`(?i)Kelvin`,
		`(?-i)API_KEY`,
		`colou?r`,
		`(ab)+cd`,
		`x(yz){2,}`,
		`(foo|b)ar`,
		`a.*b`,
		`\bdef\s+\w+`,
		`(?s)start.end`,
		`[Ss]ecret`,
		`café`,
		`(?i)café`,
		`(?:ss|t)ask`,
		// A branch with no required text, and an optional part that is the
		// longest literal in the pattern: neither may become required.
		`token|[0-9]{3}`,
		`x(longword)?y`,
		`q(abcdefgh)*z`,
	}
	contents := []string{
		"", "nothing here", "BRMP-123", "brmp-9", "ticket Brmp-42",
		"AKIAABCDEFGHIJKLMNOP", "key = sk_live_123", "-----begin RSA", "XOXB-token",
		`api_key = "abcdefghijklmnop"`, `SECRET: 'abcdefghijklmnop'`, `token="AAAAAAAAAAAA"`,
		"stages:\n        - Name: build\n", "released 2026-01-15",
		"KELVIN", "Kelvin", "api_key", "API_KEY", "color", "colour",
		"ababcd", "xyzyz", "far bar", "a long way to b", "def main", "start\nend",
		"Secret", "ſecret", "CAFÉ", "café", "ſſask", "task",
		"123", "xy", "qz",
	}
	for _, p := range patterns {
		for _, caseSensitive := range []bool{true, false} {
			n := &RegexNode{Pattern: p}
			re := n.RegexpFor(caseSensitive)
			if re == nil {
				t.Fatalf("%s did not compile", p)
			}
			for _, c := range contents {
				if re.MatchString(c) && !n.MayMatch([]byte(c), caseSensitive) {
					t.Errorf("/%s/ caseSensitive=%v matches %q and the prefilter ruled it out", p, caseSensitive, c)
				}
			}
		}
	}
}

// And it does rule things out, which is what it is for.
func TestThePrefilterRulesOutFilesWithoutTheText(t *testing.T) {
	cases := []struct {
		pattern       string
		caseSensitive bool
		content       string
		want          bool
	}{
		{`BRMP-[0-9]+`, false, "package main\nfunc main() {}\n", false},
		{`BRMP-[0-9]+`, false, "see brmp-7", true},
		{`BRMP-[0-9]+`, true, "see brmp-7", false},
		{`(AKIA[0-9A-Z]{16}|sk_live|-----BEGIN|xox[bp]-)`, false, "plain go code", false},
		{`(AKIA[0-9A-Z]{16}|sk_live|-----BEGIN|xox[bp]-)`, false, "Sk_Live", true},
		// A k or s in a folded literal: a file holding the Kelvin sign or the
		// long s goes to the regexp.
		{`sk_live`, false, "K and nothing else", true},
		{`sk_live`, true, "K and nothing else", false},
	}
	for _, c := range cases {
		n := &RegexNode{Pattern: c.pattern}
		if got := n.MayMatch([]byte(c.content), c.caseSensitive); got != c.want {
			t.Errorf("/%s/ caseSensitive=%v on %q = %v, want %v", c.pattern, c.caseSensitive, c.content, got, c.want)
		}
	}
}

// No required text means no prefilter, and a file is never ruled out.
func TestAPatternWithNoRequiredTextHasNoPrefilter(t *testing.T) {
	for _, p := range []string{`a*`, `[0-9]+`, `x?y?`, `.`, `(foo|[a-z])?z`, `(?i)café`} {
		if pre := newRegexPrefilter(p, syntax.Perl); pre != nil {
			t.Errorf("/%s/ built a prefilter: %+v", p, pre.literals)
		}
	}
	// A non-ASCII literal is scanned for as bytes when it is not folded.
	if newRegexPrefilter(`café`, syntax.Perl) == nil {
		t.Error("/café/ case-sensitive built no prefilter")
	}
}

func TestIndexFoldASCII(t *testing.T) {
	cases := []struct {
		s, lower string
		want     int
	}{
		{"hello BRMP-1", "brmp-", 6},
		{"bRmP", "brmp", 0},
		{"brm", "brmp", -1},
		{"xxxxBRMx brmp", "brmp", 9},
		{"", "a", -1},
		{"--x--", "-x", 1},
		// The anchor is k, mid-literal: a match at the start, in the middle
		// and at the very end, and a k with no match around it.
		{"AKIA0000", "akia", 0},
		{"zzAkIazz", "akia", 2},
		{"key fooAKIA", "akia", 7},
		{"kkkk akik", "akia", -1},
		{"-----BEGIN", "-----begin", 0},
	}
	for _, c := range cases {
		if got := indexFoldASCII([]byte(c.s), []byte(c.lower), max(0, rarestLetter([]byte(c.lower)))); got != c.want {
			t.Errorf("indexFoldASCII(%q, %q) = %d, want %d", c.s, c.lower, got, c.want)
		}
	}
}

// A first byte whose one case is everywhere and the other nowhere must not make
// the scan quadratic: 8 MB of 'A' with a literal starting 'a' that never matches.
func BenchmarkIndexFoldASCIIOneCaseEverywhere(b *testing.B) {
	s := bytes.Repeat([]byte("A"), 8<<20)
	for b.Loop() {
		indexFoldASCII(s, []byte("az"), 0)
	}
}

func BenchmarkFoldedLiteralRegexOverCode(b *testing.B) {
	content := bytes.Repeat([]byte("func main() { fmt.Println(\"hello world\") }\n"), 20000)
	n := &RegexNode{Pattern: `BRMP-[0-9]+`}
	re := n.RegexpFor(false)
	b.Run("regexp only", func(b *testing.B) {
		for b.Loop() {
			re.FindAllIndex(content, -1)
		}
	})
	b.Run("prefilter first", func(b *testing.B) {
		for b.Loop() {
			if n.MayMatch(content, false) {
				re.FindAllIndex(content, -1)
			}
		}
	})
}

// RequiredLiterals carries the same rule outward: every case-insensitive match
// contains one of the literals, once both are folded the way a case-folding
// index folds (ToLower of ToUpper, so the long s and the Kelvin sign land on s
// and k).
func TestRequiredLiteralsHoldForEveryMatch(t *testing.T) {
	fold := func(s string) string {
		return strings.Map(func(r rune) rune { return unicode.ToLower(unicode.ToUpper(r)) }, s)
	}
	patterns := []string{
		`BRMP-[0-9]+`, `(AKIA[0-9A-Z]{16}|sk_live|-----BEGIN|xox[bp]-)`, `[Ss]ecret`, `(?:ss|t)ask`,
		`(?i)Kelvin`, `colou?r`, `(ab)+cd`, `x(longword)?y`, `[a-z]+Handler`, `aws_[a-z_]+`,
	}
	contents := []string{
		"BRMP-123", "brmp-9", "AKIAABCDEFGHIJKLMNOP", "key = sk_live_123", "XOXB-token", "Secret",
		"ſecret", "ſſask", "task", "KELVIN", "Kelvin", "color", "ababcd", "xlongwordy", "xy",
		"func myHandler()", "ResourceHANDLER", "aws_s3_bucket", "AWS_REGION", "awſ_x",
	}
	for _, p := range patterns {
		lits := RequiredLiterals(p)
		if lits == nil {
			t.Errorf("/%s/ has required text and RequiredLiterals returned none", p)

			continue
		}
		re := (&RegexNode{Pattern: p}).RegexpFor(false)
		for _, c := range contents {
			if !re.MatchString(c) {
				continue
			}
			found := false
			for _, l := range lits {
				if strings.Contains(fold(c), l) {
					found = true
				}
			}
			if !found {
				t.Errorf("/%s/ matches %q, which holds none of %q once folded", p, c, lits)
			}
		}
	}

	for p, want := range map[string]string{`[a-z]+Handler`: "handler", `aws_[a-z_]+`: "aws_", `BRMP-[0-9]+`: "brmp-"} {
		if got := RequiredLiterals(p); len(got) != 1 || got[0] != want {
			t.Errorf("RequiredLiterals(%q) = %q, want [%q]", p, got, want)
		}
	}
	for _, p := range []string{`a*`, `[0-9]+`, `.`, `café`, `(unclosed`} {
		if got := RequiredLiterals(p); got != nil {
			t.Errorf("RequiredLiterals(%q) = %q, want nil", p, got)
		}
	}
}
