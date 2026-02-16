//go:build go1.18

package search

import (
	"testing"
)

func FuzzSearch(f *testing.F) {
	// Seed corpus with the existing test cases to guide the fuzzer.
	f.Add("cat", false)
	f.Add("brown AND cat", false)
	f.Add("lazy cat", false)
	f.Add("dog OR fox", false)
	f.Add("brown NOT dog", false)
	f.Add("(lazy OR house) AND cat", false)
	f.Add("/[cb]at/", false)
	f.Add("complexity=5", false)
	f.Add("complexity>=8", false)
	f.Add("complexity!=3", false)
	f.Add("lazy AND complexity<=3", false)
	f.Add("lang=go", false)
	f.Add("ext=py", false)
	f.Add("lang=go,python", false)
	f.Add("Cat", false)
	f.Add("cat", true)
	f.Add("Cat", true)
	f.Add("lang=go,,python", false)
	f.Add("lang=,go,python", false)
	f.Add("lang=go,python,", false)
	f.Add("ext=go,py", false)
	f.Add("lang=go,py+thon,java", false)
	f.Add("lang=go NOT python", false)
	f.Add("(lang=go OR lang=python) AND complexity>=5", false)
	f.Add("lang=GO,PYTHON", false)

	// Add known-bad or tricky inputs we've fixed.
	f.Add("AND", false)
	f.Add("OR", false)
	f.Add("NOT", false)
	f.Add(">", false)
	f.Add("lang=", false)
	f.Add("lang=>", false)
	f.Add("lang>>5", false)
	f.Add(",", false)
	f.Add("cat AND", false)
	f.Add(" ", false)
	f.Add("", false)
	f.Add("cat)", false)
	f.Add("(cat", false)
	f.Add("lazy AND NOT dog", false)
	f.Add("NOT lang=go", false)

	// The test documents for the fuzzer to run against.
	testDocs := []*Document{
		{Path: "file1.go", Filename: "file1.go", Language: "Go", Extension: "go", Content: []byte("A brown cat is in the house."), Complexity: 2},
		{Path: "file2.go", Filename: "file2.go", Language: "Go", Extension: "go", Content: []byte("A quick brown dog jumps over the lazy fox."), Complexity: 5},
		{Path: "file3.py", Filename: "file3.py", Language: "Python", Extension: "py", Content: []byte("The lazy cat sat on the mat."), Complexity: 3},
		{Path: "file4.py", Filename: "file4.py", Language: "Python", Extension: "py", Content: []byte("A bat and a cat are friends."), Complexity: 8},
		{Path: "file5.rs", Filename: "file5.rs", Language: "Rust", Extension: "rs", Content: []byte("This is a complex document about Go programming."), Complexity: 9},
	}
	se := NewSearchEngine(testDocs)

	f.Fuzz(func(t *testing.T, query string, caseSensitive bool) {
		// The primary goal of this fuzz test is to ensure that the Search
		// function does not panic on unexpected inputs. We don't need to
		// validate the correctness of the search results here, as that is
		// covered by the unit tests.
		_, _ = se.Search(query, caseSensitive)
	})
}
