package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// keeps track of files stored in the index so we can open them to find matches\
var id = 0
var idToFile []string
var fileToId = map[string]int{}

func main() {
	// walk the directory getting files and indexing
	_ = filepath.Walk(".", func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil // we only care about files
		}

		res, err := os.ReadFile(root)
		if err != nil {
			return nil // swallow error
		}

		// don't index binary files by looking for nul byte, similar to how grep does it
		if bytes.IndexByte(res, 0) != -1 {
			return nil
		}

		// only index up to about 5kb
		if len(res) > 5000 {
			res = res[:5000]
		}

		// add the document to the index
		_ = Add(Itemise(Tokenize(string(res))))
		// store the association from what's in the index to the filename, we know its 0 to whatever so this works
		idToFile = append(idToFile, root)
		fileToId[root] = id
		id++
		return nil
	})

	fmt.Printf("currentBlockDocumentCount:%v currentDocumentCount:%v currentBlockStartDocumentCount:%v\n", currentBlockDocumentCount, currentDocumentCount, currentBlockStartDocumentCount)

	var searchTerm string
	for {
		fmt.Println("enter search term: ")
		_, _ = fmt.Scanln(&searchTerm)

		res := Search(Queryise(searchTerm))
		fmt.Println("--------------")
		fmt.Println(len(res), "index result(s)")
		fmt.Println("")
		for _, r := range res {
			fmt.Println(idToFile[r])
			matching := findMatchingLines(idToFile[r], searchTerm, 5)
			for _, l := range matching {
				fmt.Println(l)
			}
			if len(matching) == 0 {
				fmt.Println("false positive match")
			}
			fmt.Println("")
		}
	}
}

// Given a file and a query try to open the file, then look through its lines
// and see if any of them match something from the query up to a limit
// Note this will return partial matches as if any term matches its considered a match
// and there is no accounting for better matches...
// In other words it's a very dumb way of doing this and probably has horrible runtime
// performance to match
func findMatchingLines(filename string, query string, limit int) []string {
	res, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}

	terms := strings.Fields(strings.ToLower(query))
	var cleanTerms []string
	for _, t := range terms {
		if len(t) >= 3 {
			cleanTerms = append(cleanTerms, t)
		}
	}

	var matches []string
	for i, l := range strings.Split(string(res), "\n") {

		low := strings.ToLower(l)
		found := false
		for _, t := range terms {
			if strings.Contains(low, t) {
				if !found {
					matches = append(matches, fmt.Sprintf("%v. %v", i+1, l))
				}
				found = true
			}
		}

		if len(matches) >= limit {
			return matches
		}
	}

	return matches
}
