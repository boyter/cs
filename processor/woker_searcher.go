package processor

import (
	"bytes"
	"errors"
	str "github.com/boyter/cs/string"
	"gopkg.in/src-d/enry.v1/regex"
	"strings"
)

type SearcherWorker struct {
	input           chan *fileJob
	output          chan *fileJob
	searchParams    []searchParams
	FileCount       int64 // Count of the number of files that have been processed
	BinaryCount     int64 // Count the number of binary files
	MinfiedCount    int64
	SearchString    []string
	IncludeMinified bool
	IncludeBinary   bool
	CaseSensitive   bool
	MatchLimit      int
	InstanceId      int
}

func NewSearcherWorker(input chan *fileJob, output chan *fileJob) SearcherWorker {
	return SearcherWorker{
		input:        input,
		output:       output,
		SearchString: []string{},
		MatchLimit:   -1, // sensible default
	}
}

// Does the actual processing of stats and as such contains the hot path CPU call
func (f *SearcherWorker) Start() {
	// Build out the search params
	f.searchParams = parseArguments(f.SearchString)

	for res := range f.input {

		// Check for the presence of a null byte indicating that this
		// is likely a binary file and if so ignore it
		if !f.IncludeBinary {
			if bytes.IndexByte(res.Content, '\x00') != -1 {
				res.Binary = true
				continue
			}
		}

		// Check if the file is minified and if so ignore it
		if !f.IncludeMinified {
			split := bytes.Split(res.Content, []byte("\n"))
			sumLineLength := 0
			for _, s := range split {
				sumLineLength += len(s)
			}
			averageLineLength := sumLineLength / len(split)

			if averageLineLength > MinifiedLineByteLength {
				res.Minified = true
				continue
			}
		}

		// Now we do the actual search against the file
		// TODO needs to deal with NOT logic
		// TODO also need to try against the filename IE even with not text matches it should count
		for i, needle := range f.searchParams {
			didSearch := false
			switch needle.Type {
			case Default, Quoted:
				didSearch = true
				if f.CaseSensitive {
					res.MatchLocations[needle.Term] = str.IndexAll(string(res.Content), needle.Term, f.MatchLimit)
				} else {
					res.MatchLocations[needle.Term] = str.IndexAllIgnoreCaseUnicode(string(res.Content), needle.Term, f.MatchLimit)
				}
			case Regex:
				x, err := f.RegexSearch(needle, res)
				if err == nil { // Error indicates a regex compile fail so safe to ignore here
					didSearch = true
					res.MatchLocations[needle.Term] = x
				}
			case Fuzzy1:
				didSearch = true
				terms := makeFuzzyDistanceOne(strings.TrimRight(needle.Term, "~1"))
				matchLocations := [][]int{}
				for _, t := range terms {
					if f.CaseSensitive {
						matchLocations = append(matchLocations, str.IndexAll(string(res.Content), t, f.MatchLimit)...)
					} else {
						matchLocations = append(matchLocations, str.IndexAllIgnoreCaseUnicode(string(res.Content), t, f.MatchLimit)...)
					}
				}
				res.MatchLocations[needle.Term] = matchLocations
			case Fuzzy2:
				didSearch = true
				terms := makeFuzzyDistanceTwo(strings.TrimRight(needle.Term, "~2"))
				matchLocations := [][]int{}
				for _, t := range terms {
					if f.CaseSensitive {
						matchLocations = append(matchLocations, str.IndexAll(string(res.Content), t, f.MatchLimit)...)
					} else {
						matchLocations = append(matchLocations, str.IndexAllIgnoreCaseUnicode(string(res.Content), t, f.MatchLimit)...)
					}
				}
				res.MatchLocations[needle.Term] = matchLocations
			}

			// We currently ignore things such as NOT and as such
			// we don't want to break out if we run into them
			// so only update the score IF there was a search
			// which also makes this by default an AND search
			if didSearch {

				// If we did a search but the previous was a NOT we need to only continue if we found nothing
				if i != 0 && f.searchParams[i-1].Type == Negated {
					if len(res.MatchLocations[needle.Term]) != 0 {
						res.Score = 0
						break
					}
				} else {
					// Normal search so ensure we got something by default and logic rules
					if len(res.MatchLocations[needle.Term]) == 0 {
						res.Score = 0
						break
					}
				}

				// Without ranking this score favors the most matches which is
				// basic but better than nothing
				res.Score += float64(len(res.MatchLocations[needle.Term]))
			}
		}

		if res.Score != 0 {
			f.output <- res
		}
	}

	close(f.output)
}

func (f *SearcherWorker) RegexSearch(needle searchParams, res *fileJob) (x [][]int, err error) {
	// Its possible the user supplies an invalid regex and if so we should not crash
	// but ignore it
	defer func() {
		if recover() != nil {
			err = errors.New("regex compile failure issue")
		}
	}()

	r := regex.MustCompile(needle.Term)
	return r.FindAllIndex(res.Content, f.MatchLimit), nil
}
