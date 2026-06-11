// SPDX-License-Identifier: MIT

package str

import (
	"math"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// IndexAll extracts all the locations of a string inside another string
// up-to the defined limit and does so without regular expressions,
// which makes it faster than FindAllIndex in most situations while
// not being any slower. It performs worst when working against random
// data.
//
// Some benchmark results to illustrate the point (find more in index_benchmark_test.go)
//
// BenchmarkFindAllIndex-8                         2458844	       480.0 ns/op
// BenchmarkIndexAll-8                            14819680	        79.6 ns/op
//
// For pure literal searches IE no regular expression logic this method
// is a drop-in replacement for re.FindAllIndex but generally much faster.
//
// Similarly to how FindAllIndex the limit option can be passed -1
// to get all matches.
//
// Note that this method is explicitly case-sensitive in its matching.
// A return value of nil indicates no match.
func IndexAll(haystack string, needle string, limit int) [][]int {
	// The below needed to avoid timeout crash found using go-fuzz
	if len(haystack) == 0 || len(needle) == 0 {
		return nil
	}

	// Return contains a slice of slices where index 0 is the location of the match in bytes
	// and index 1 contains the end location in bytes of the match
	var locs [][]int

	// Perform the first search outside the main loop to make the method
	// easier to understand
	searchText := haystack
	offSet := 0
	loc := strings.Index(searchText, needle)

	if limit <= -1 {
		// Similar to how regex FindAllString works
		// if we have -1 as the limit set to max to
		//  try to get everything
		limit = math.MaxInt32
	} else {
		// Increment by one because we do count++ at the start of the loop
		// and as such there is a off by 1 error in the return otherwise
		limit++
	}

	var count int
	for loc != -1 {
		count++

		if count == limit {
			break
		}

		// trim off the portion we already searched, and look from there
		searchText = searchText[loc+len(needle):]
		locs = append(locs, []int{loc + offSet, loc + offSet + len(needle)})

		// We need to keep the offset of the match so we continue searching
		offSet += loc + len(needle)

		// strings.Index does checks of if the string is empty so we don't need
		// to explicitly do it ourselves
		loc = strings.Index(searchText, needle)
	}

	// Retain compatibility with the FindAllIndex method
	if len(locs) == 0 {
		return nil
	}

	return locs
}

// Rarity rank for ASCII letters (0 = rarest, 25 = most common).
// Based on English + programming keyword frequencies.
// Non-letter bytes default to 100 (treated as common, so we prefer ASCII letter trigrams).
var _charRarity [256]int

// I am against init generally, but this should not have much impact.
// We build a lookup table for character rarity to speed up trigram scoring.
// Allowing us to pick an uncommon trigram as the starting point for searching.
// Which can greatly improve search performance for common needles.
func init() {
	for i := range _charRarity {
		_charRarity[i] = 100
	}
	// 's' and 'k' are deliberately omitted from this ordering. They are the only
	// ASCII letters that case-fold to a third, non-ASCII form ('ſ' U+017F and the
	// KELVIN SIGN 'K' U+212A), giving them three fold variants, so the fast
	// two-byte indexByteTwo SIMD scan in IndexAllIgnoreCase cannot be used when
	// they are the anchor. By English frequency 'k' is rare and would otherwise
	// be the preferred anchor for any needle containing it (e.g. "kelvin"),
	// forcing the slow multi-pass fallback. Leaving them at the common default
	// (100) keeps bestCharOffset from preferring them, so such needles anchor on
	// a SIMD-capable letter instead. If a needle contains no other letter we
	// still anchor on them and the slower unicode-aware fallback handles it.
	order := "zjqxvbpygfwmucldrhnioate"
	for i, ch := range order {
		_charRarity[ch] = i
		if ch >= 'a' && ch <= 'z' {
			_charRarity[ch-32] = i // uppercase
		}
	}
}

// bestCharOffset returns the rune offset of the rarest character window
// in needleRune. Scores each window by summing character rarity (lower = rarer).
// Falls back to offset 0 if needle is shorter than width.
func bestCharOffset(needleRune []rune, width int) int {
	if len(needleRune) <= width {
		return 0
	}

	bestOffset := 0
	bestScore := math.MaxInt32

	for i := 0; i <= len(needleRune)-width; i++ {
		score := 0
		for j := 0; j < width; j++ {
			r := needleRune[i+j]
			if r < 256 {
				score += _charRarity[r]
			} else {
				score += 100
			}
		}
		if score < bestScore {
			bestScore = score
			bestOffset = i
		}
	}

	return bestOffset
}

// if the IndexAllIgnoreCase method is called frequently with the same patterns
// (which is a common case) this is here to speed up the case permutations
// it is limited to a size of 10 so it never gets that large but really
// allows things to run faster
var _permuteCache = map[string][]string{}
var _permuteCacheLock = sync.RWMutex{}

// CacheSize this is public so it can be modified depending on project needs
// you can increase this value to cache more of the case permutations which
// can improve performance if doing the same searches over and over
var CacheSize = 10

// IndexAllIgnoreCase extracts all the locations of a string inside another string
// up-to the defined limit. It is designed to be faster than uses of FindAllIndex with
// case-insensitive matching enabled, by looking for string literals first and then
// checking for exact matches. It also does so in a unicode aware way such that a search
// for S will search for S s and ſ which a simple strings.ToLower over the haystack
// and the needle will not.
//
// The result is the ability to search for literals without hitting the regex engine
// which can at times be horribly slow. This, by contrast, is much faster. See
// index_ignorecase_benchmark_test.go for some head-to-head results. Generally,
// so long as we aren't dealing with random data, this method should be considerably
// faster (in some cases thousands of times) or just as fast. Of course, it cannot
// do regular expressions, but that's by design.
//
// Performance notes: IndexAllIgnoreCase is effectively memory-bandwidth-limited
// for large haystacks. The permutation approach (2^min(len,3) passes of
// strings.Index) is near-optimal because strings.Index compiles to SIMD
// assembly. Attempts to reduce passes (2-char prefix with byte-level |0x20
// verification) yielded only ~13% improvement on 1.87GB at the cost of
// significant complexity. Single-pass approaches using strings.IndexByte
// are slower due to poor selectivity (too many candidates per byte match).
// Further gains would require assembly-level SIMD or parallelism, which
// is out of scope for this library
//
// For pure literal searches IE no regular expression logic this method
// is a drop-in replacement for re.FindAllIndex but generally much faster.
func IndexAllIgnoreCase(haystack string, needle string, limit int) [][]int {
	// The below needed to avoid timeout crash found using go-fuzz
	if len(haystack) == 0 || len(needle) == 0 {
		return nil
	}

	// One of the problems with finding locations ignoring case is that
	// the different case representations can have different byte counts
	// which means the locations using strings or bytes Index can be off
	// if you apply strings.ToLower to your haystack then use strings.Index.
	//
	// This can be overcome using regular expressions but suffers the penalty
	// of hitting the regex engine and paying the price of case-
	// insensitive match there.
	//
	// This method tries something else which is used by some regex engines
	// such as the one in Rust where given a str literal if you get
	// all the case options of that such as turning foo into foo Foo fOo FOo foO FoO fOO FOO
	// and then use Boyer-Moore or some such for those. Of course using something
	// like Aho-Corasick or Rabin-Karp to get multi match would be a better idea so you
	// can match all of the input in one pass.
	//
	// If the needle is over some amount of characters long you chop off the first few
	// and then search for those. However this means you are not finding actual matches and as such
	// you the need to validate a potential match after you have found one.
	// The confirmation match is done in a loop because for some literals regular expression
	// is still to slow, although for most its a valid option.
	var locs [][]int

	// Char limit is the cut-off where we switch from all case permutations
	// to just the first 3 and then check for an actual match
	// in my tests 3 speeds things up the most against test data
	// of many famous books concatenated together and large
	// amounts of data from /dev/urandom
	var charLimit = 3

	if utf8.RuneCountInString(needle) <= charLimit {
		// We are below the limit we set, so get all the search
		// terms and search for that

		// Generally, I am against caches inside libraries. However, in this case...
		// when the IndexAllIgnoreCase method is called repeatedly, it quite often
		// ends up performing case folding on the same thing over and over again, which
		// can become the most expensive operation. So we keep a VERY small cache
		// to avoid that being an issue.
		// Try a read lock first.
		// This allows multiple goroutines to hit the cache simultaneously
		// without blocking each other.
		_permuteCacheLock.RLock()
		searchTerms, ok := _permuteCache[needle]
		_permuteCacheLock.RUnlock()

		if !ok {
			// we now have to do something about it
			_permuteCacheLock.Lock()

			// Another goroutine might have
			// inserted the key while we were waiting for the lock.
			searchTerms, ok = _permuteCache[needle]
			if !ok {
				if len(_permuteCache) > CacheSize {
					_permuteCache = map[string][]string{}
				}
				searchTerms = PermuteCaseFolding(needle)
				_permuteCache[needle] = searchTerms
			}
			_permuteCacheLock.Unlock()
		}

		// Stream the permutation matches out in regexp.FindAllIndex order,
		// stopping as soon as we have limit of them, so a small limit on a huge
		// haystack stays cheap. collectFoldedMatches already returns ordered,
		// non-overlapping results, so we can return directly.
		locs = collectFoldedMatches(haystack, searchTerms, limit)

		if len(locs) == 0 {
			return nil
		}
		return locs
	} else {
		// Over the character limit so look for potential matches and only then check to find real ones

		// Note that we have to use runes here to avoid cutting bytes off so
		// cast things around to ensure it works
		needleRune := []rune(needle)

		searchStart := bestCharOffset(needleRune, 1)
		searchChar := string(needleRune[searchStart : searchStart+1])

		_permuteCacheLock.RLock()
		searchTerms, ok := _permuteCache[searchChar]
		_permuteCacheLock.RUnlock()

		if !ok {
			// nothing found so we have to do something
			_permuteCacheLock.Lock()
			searchTerms, ok = _permuteCache[searchChar]
			if !ok {
				if len(_permuteCache) > CacheSize {
					_permuteCache = map[string][]string{}
				}
				searchTerms = PermuteCaseFolding(searchChar)
				_permuteCache[searchChar] = searchTerms
			}
			_permuteCacheLock.Unlock()
		}

		// Fast path: when the search character is an ASCII letter with exactly
		// two fold variants (upper + lower), use indexByteTwo to find both in
		// a single SIMD pass instead of two separate IndexAll scans.
		// This covers all ASCII letters except s/S (which also fold to ſ)
		// and k/K (which also fold to K); those have three fold variants so
		// the len == 2 guard below excludes them. The rarity table (see init)
		// omits s and k so bestCharOffset avoids anchoring on them whenever the
		// needle has a SIMD-capable letter, keeping needles like "kelvin" on
		// this fast path.
		searchRune := needleRune[searchStart]
		useSIMD := len(searchTerms) == 2 && len(searchTerms[0]) == 1 && len(searchTerms[1]) == 1 &&
			searchRune < utf8.RuneSelf &&
			unicode.ToLower(searchRune) != unicode.ToUpper(searchRune)

		if useSIMD {
			b1 := searchTerms[0][0]
			b2 := searchTerms[1][0]
			haystackBytes := unsafe.Slice(unsafe.StringData(haystack), len(haystack))
			offset := 0
			for offset < len(haystackBytes) {
				idx := indexByteTwo(haystackBytes[offset:], b1, b2)
				if idx < 0 {
					break
				}
				bytePos := offset + idx

				needleStart := bytePos
				skip := false
				for k := 0; k < searchStart; k++ {
					if needleStart <= 0 {
						skip = true
						break
					}
					_, size := utf8.DecodeLastRuneInString(haystack[:needleStart])
					needleStart -= size
				}
				if !skip {
					pos := needleStart
					isMatch := true

					for i := 0; i < len(needleRune); i++ {
						if pos >= len(haystack) {
							isMatch = false
							break
						}

						r, size := utf8.DecodeRuneInString(haystack[pos:])

						if r != needleRune[i] {
							foldMatch := false
							for _, f := range AllSimpleFold(r) {
								if f == needleRune[i] {
									foldMatch = true
									break
								}
							}
							if !foldMatch {
								isMatch = false
								break
							}
						}
						pos += size
					}

					if isMatch {
						locs = append(locs, []int{needleStart, pos})
					}
				}
				offset = bytePos + 1
			}
		} else {
			// Slow path for characters with >2 fold variants (e.g. s/S/ſ, k/K/K)
			// or non-ASCII characters. Uses multiple IndexAll passes.
			for _, term := range searchTerms {
				potentialMatches := IndexAll(haystack, term, -1)

				for _, match := range potentialMatches {
					needleStart := match[0]
					skip := false
					for k := 0; k < searchStart; k++ {
						if needleStart <= 0 {
							skip = true
							break
						}
						_, size := utf8.DecodeLastRuneInString(haystack[:needleStart])
						needleStart -= size
					}
					if skip {
						continue
					}
					pos := needleStart
					isMatch := true

					for i := 0; i < len(needleRune); i++ {
						if pos >= len(haystack) {
							isMatch = false
							break
						}

						r, size := utf8.DecodeRuneInString(haystack[pos:])

						if r != needleRune[i] {
							foldMatch := false
							for _, f := range AllSimpleFold(r) {
								if f == needleRune[i] {
									foldMatch = true
									break
								}
							}
							if !foldMatch {
								isMatch = false
								break
							}
						}
						pos += size
					}

					if isMatch {
						locs = append(locs, []int{needleStart, pos})
					}
				}
			}
		}

	}

	// The case-fold search above runs each permutation independently and
	// concatenates the hits, so the slice can be unordered (a lowercase match
	// found after an uppercase one further left) and can contain overlaps (e.g.
	// needle "aa" matching both [0,2] and [1,3] in "aAa", or different byte
	// lengths via the ſ/KELVIN SIGN folds). regexp.FindAllIndex, which this
	// method mirrors, returns matches left-to-right and non-overlapping, so we
	// order and resolve overlaps before applying the limit.
	locs = sortAndDedupe(locs)

	// Match regexp.FindAllIndex's limit semantics: a negative limit returns all
	// matches, while a non-negative limit returns at most that many (so 0 returns
	// none). The case-sensitive IndexAll already behaves this way.
	if limit >= 0 && len(locs) > limit {
		locs = locs[:limit]
	}

	// Retain compatibility with the FindAllIndex method
	if len(locs) == 0 {
		return nil
	}

	return locs
}

// collectFoldedMatches searches haystack for every case-fold permutation in
// terms and returns the matches in left-to-right, non-overlapping order,
// stopping once it has limit of them (limit < 0 means all, limit == 0 means
// none). It is the case-folded equivalent of regexp.FindAllIndex.
//
// It works as a lazy k-way merge: each term keeps its next pending match, and
// each step takes the earliest-starting one, keeps it only if it begins at or
// after the end of the previously kept match (regexp's consume-and-advance
// rule), then advances that term to its next match. Crucially the next match is
// found one byte past the current one, so overlapping occurrences of a periodic
// permutation (e.g. "aa" at [0,2] and [1,3] in "aaa") are still produced. At
// most one permutation can match at any given byte offset, so two pending
// matches never share a start. Because results are emitted in order we can stop
// at limit without scanning the rest of the haystack.
func collectFoldedMatches(haystack string, terms []string, limit int) [][]int {
	if limit == 0 {
		return nil
	}

	// next[i] is the start offset of term i's pending match (-1 once exhausted);
	// from[i] is where its following search resumes (one byte later).
	next := make([]int, len(terms))
	from := make([]int, len(terms))
	for i, t := range terms {
		next[i] = strings.Index(haystack, t)
		if next[i] >= 0 {
			from[i] = next[i] + 1
		}
	}

	var out [][]int
	lastEnd := 0
	for {
		best := -1
		for i := range next {
			if next[i] >= 0 && (best < 0 || next[i] < next[best]) {
				best = i
			}
		}
		if best < 0 {
			break
		}

		pos := next[best]
		end := pos + len(terms[best])
		if pos >= lastEnd {
			out = append(out, []int{pos, end})
			lastEnd = end
			if limit > 0 && len(out) == limit {
				break
			}
		}

		if idx := strings.Index(haystack[from[best]:], terms[best]); idx < 0 {
			next[best] = -1
		} else {
			next[best] = from[best] + idx
			from[best] = next[best] + 1
		}
	}

	return out
}

// sortAndDedupe orders matches left-to-right and removes overlaps so the result
// matches regexp.FindAllIndex semantics: a match is kept only if it starts at or
// after the end of the previously kept match (leftmost wins, scanning resumes
// after it). Exact duplicates produced by different case permutations are
// dropped by the same rule. The slice is sorted and filtered in place.
func sortAndDedupe(locs [][]int) [][]int {
	if len(locs) < 2 {
		return locs
	}

	slices.SortFunc(locs, func(a, b []int) int {
		if a[0] != b[0] {
			return a[0] - b[0]
		}
		return a[1] - b[1]
	})

	// Compact non-overlapping matches into the front of the slice; the write
	// index never overtakes the read index, so this is safe in place.
	n := 1
	lastEnd := locs[0][1]
	for _, m := range locs[1:] {
		if m[0] >= lastEnd {
			locs[n] = m
			n++
			lastEnd = m[1]
		}
	}

	return locs[:n]
}
