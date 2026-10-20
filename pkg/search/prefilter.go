package search

import (
	"bytes"
	"regexp/syntax"
	"unicode/utf8"
)

// A regexPrefilter rules a file out before the regexp engine runs. Go's engine
// scans an unanchored pattern at tens of megabytes a second, and it skips ahead
// on a literal prefix only when the pattern is case-sensitive, so a folded
// /BRMP-[0-9]+/ measured 200 times slower than the same pattern case-sensitive.
// Most files contain none of a pattern's required text, and finding that out
// takes a byte scan.
//
// The rule it holds: if mayMatch returns false the regexp cannot match. Every
// match of the pattern contains at least one of the literals.
type regexPrefilter struct {
	literals []prefilterLiteral
	// unicodeFold is set when a folded literal contains k or s, the two ASCII
	// letters with a non-ASCII case partner (U+212A KELVIN SIGN, U+017F LATIN
	// SMALL LETTER LONG S). The ASCII scan cannot see those, so a file holding
	// either is handed to the regexp.
	unicodeFold bool
}

type prefilterLiteral struct {
	text []byte // lower-cased when fold is set
	fold bool
	// anchor is the offset of the byte a folded scan looks for: the rarest
	// letter, so the scan stops at as few places as it can. A literal with no
	// letter has no case to fold and is scanned for as it is.
	anchor int
}

// letterRarity ranks the lower-case letters from common to rare in source
// code and prose. Only the order matters.
const letterRarity = "etaoinsrlcdhupmfgbywvkxqjz"

// rarestLetter returns the offset of the rarest letter in lower, or -1 if it
// has none.
func rarestLetter(lower []byte) int {
	at, rank := -1, -1
	for i, c := range lower {
		if r := bytes.IndexByte([]byte(letterRarity), c); r > rank {
			at, rank = i, r
		}
	}
	return at
}

var (
	kelvinSign = []byte("K")
	longS      = []byte("ſ")
)

const (
	// A literal shorter than this rules out too little to pay for the scan.
	prefilterMinLen = 2
	// An alternation wider than this is cheaper left to the regexp.
	prefilterMaxLiterals = 32
)

// newRegexPrefilter builds the prefilter for pattern parsed with flags, or
// returns nil when the pattern has no required text worth scanning for.
func newRegexPrefilter(pattern string, flags syntax.Flags) *regexPrefilter {
	re, err := syntax.Parse(pattern, flags)
	if err != nil {
		return nil
	}
	lits := requiredLiterals(re.Simplify())
	if len(lits) == 0 {
		return nil
	}
	p := &regexPrefilter{literals: lits}
	for _, l := range lits {
		if len(l.text) < prefilterMinLen {
			return nil
		}
		if l.fold && (bytes.IndexByte(l.text, 'k') >= 0 || bytes.IndexByte(l.text, 's') >= 0) {
			p.unicodeFold = true
		}
	}
	return p
}

// RequiredLiterals returns a set of literals one of which every match of the
// regular expression pattern must contain, case-folded to lower case, or nil
// when there is no such set or the pattern does not parse.
//
// It is for a caller that narrows candidates before running the regexp, such as
// a trigram index: a document holding none of the literals cannot match. The
// set is computed case-insensitively whatever the search's own case mode, which
// is the conservative direction for an index that folds case: a literal whose
// case variants a byte comparison cannot follow (any non-ASCII letter) makes the
// whole set nil rather than narrowing on a spelling a match need not use.
//
// One literal means every match contains it. Several mean every match contains
// at least one of them, so a caller that can only intersect must not require
// them all.
func RequiredLiterals(pattern string) []string {
	re, err := syntax.Parse(pattern, syntax.Perl|syntax.FoldCase)
	if err != nil {
		return nil
	}
	lits := requiredLiterals(re.Simplify())
	if len(lits) == 0 {
		return nil
	}
	out := make([]string, 0, len(lits))
	for _, l := range lits {
		out = append(out, string(l.text))
	}
	return out
}

// requiredLiterals returns a set of literals one of which every match of re
// contains, or nil when there is no such set.
func requiredLiterals(re *syntax.Regexp) []prefilterLiteral {
	switch re.Op {
	case syntax.OpLiteral:
		fold := re.Flags&syntax.FoldCase != 0
		text := make([]byte, 0, len(re.Rune))
		for _, r := range re.Rune {
			if fold {
				if r >= utf8.RuneSelf {
					// A non-ASCII rune folds onto partners of other byte
					// lengths, which a byte scan cannot follow.
					return nil
				}
				if 'A' <= r && r <= 'Z' {
					r += 'a' - 'A'
				}
			}
			text = utf8.AppendRune(text, r)
		}
		if fold && rarestLetter(text) < 0 {
			fold = false
		}
		anchor := 0
		if fold {
			anchor = rarestLetter(text)
		}
		return []prefilterLiteral{{text: text, fold: fold, anchor: anchor}}
	case syntax.OpCapture, syntax.OpPlus:
		return requiredLiterals(re.Sub[0])
	case syntax.OpRepeat:
		if re.Min >= 1 {
			return requiredLiterals(re.Sub[0])
		}
	case syntax.OpConcat:
		// Every part is required, so any one part's set will do. The best is
		// the one whose shortest literal is longest.
		var best []prefilterLiteral
		for _, sub := range re.Sub {
			if lits := requiredLiterals(sub); lits != nil && shortest(lits) > shortest(best) {
				best = lits
			}
		}
		return best
	case syntax.OpAlternate:
		// A match takes one branch, so every branch must contribute.
		var all []prefilterLiteral
		for _, sub := range re.Sub {
			lits := requiredLiterals(sub)
			if lits == nil {
				return nil
			}
			all = append(all, lits...)
		}
		if len(all) > prefilterMaxLiterals {
			return nil
		}
		return all
	}
	return nil
}

func shortest(lits []prefilterLiteral) int {
	if len(lits) == 0 {
		return 0
	}
	n := len(lits[0].text)
	for _, l := range lits[1:] {
		n = min(n, len(l.text))
	}
	return n
}

// mayMatch reports whether content could hold a match. A nil prefilter rules
// nothing out.
func (p *regexPrefilter) mayMatch(content []byte) bool {
	if p == nil {
		return true
	}
	for _, l := range p.literals {
		if l.fold {
			if indexFoldASCII(content, l.text, l.anchor) >= 0 {
				return true
			}
		} else if bytes.Contains(content, l.text) {
			return true
		}
	}
	return p.unicodeFold && (bytes.Contains(content, kelvinSign) || bytes.Contains(content, longS))
}

// indexFoldASCII returns the first index of lower in s ignoring ASCII case, or
// -1. lower must already be lower-case, and anchor is the offset of the byte to
// scan for, either case of which may stand at that place in a match.
func indexFoldASCII(s, lower []byte, anchor int) int {
	if len(lower) == 0 {
		return 0
	}
	first := lower[anchor]
	upper := first
	if 'a' <= first && first <= 'z' {
		upper = first - ('a' - 'A')
	}
	// The next position of each case of the anchor byte, kept between rounds
	// so a common one does not make the scan for a rare one start again.
	nextLower, nextUpper := -2, -2
	if upper == first {
		nextUpper = -1
	}
	next := func(pos *int, c byte, from int) {
		if *pos != -1 && *pos < from {
			if k := bytes.IndexByte(s[from:], c); k >= 0 {
				*pos = from + k
			} else {
				*pos = -1
			}
		}
	}
	for at := anchor; at+len(lower)-anchor <= len(s); at++ {
		next(&nextLower, first, at)
		if upper != first {
			next(&nextUpper, upper, at)
		}
		switch {
		case nextLower < 0 && nextUpper < 0:
			return -1
		case nextLower < 0:
			at = nextUpper
		case nextUpper < 0:
			at = nextLower
		default:
			at = min(nextLower, nextUpper)
		}
		start := at - anchor
		if start+len(lower) > len(s) {
			return -1
		}
		if equalFoldASCII(s[start:start+len(lower)], lower) {
			return start
		}
	}
	return -1
}

func equalFoldASCII(s, lower []byte) bool {
	for i, c := range s {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != lower[i] {
			return false
		}
	}
	return true
}
