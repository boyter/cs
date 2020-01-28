package processor

import (
	"bytes"
	"fmt"
	"github.com/boyter/cs/processor/snippet"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var TotalCount int64

// How many matches of a term do we want to find in any file?
// Limited to 100 because after that is what we are looking for any more
// relevant to the user?
var MatchLimit = 100

// This is responsible for spinning up all of the jobs
// that read files from disk into memory
func FileReaderWorker(input chan *FileJob, output chan *FileJob) {
	var wg sync.WaitGroup
	TotalCount = 0
	for i := 0; i < FileReadJobWorkers; i++ {
		wg.Add(1)

		go func() {
			for res := range input {
				fi, err := os.Stat(res.Location)
				if err != nil {
					continue
				}

				var content []byte
				fileStartTime := makeTimestampNano()

				var s int64 = 1024000

				// Only read up to ~1MB of a file because anything beyond that is probably pointless
				if fi.Size() < s {
					content, err = ioutil.ReadFile(res.Location)
				} else {
					r, err := os.Open(res.Location)
					if err != nil {
						continue
					}

					var tmp [1024000]byte
					_, _ = io.ReadFull(r, tmp[:])
					_ = r.Close()
				}

				if Trace {
					printTrace(fmt.Sprintf("nanoseconds read into memory: %s: %d", res.Location, makeTimestampNano()-fileStartTime))
				}

				if err == nil {
					atomic.AddInt64(&TotalCount, 1)
					res.Content = content
					output <- res
				} else {
					if Verbose {
						printWarn(fmt.Sprintf("error reading: %s %s", res.Location, err))
					}
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(output)
}

// Does the actual processing of stats and as such contains the hot path CPU call
func FileProcessorWorker(input chan *FileJob, output chan *FileJob) {
	var wg sync.WaitGroup

	for i := 0; i < FileProcessJobWorkers; i++ {
		wg.Add(1)

		go func() {
			for res := range input {
				if bytes.IndexByte(res.Content, '\x00') != -1 {
					res.Binary = true
				} else {

					// Check if the file is minified and if so ignore it
					split := bytes.Split(res.Content, []byte("\n"))
					sumLineLength := 0
					for _, s := range split {
						sumLineLength += len(s)
					}
					averageLineLength := sumLineLength / len(split)

					if averageLineLength < MinifiedGeneratedLineByteLength {
						// what we need to do is check for each term if it exists, and then use that to determine if its a match
						contentLower := strings.ToLower(string(res.Content))
						// https://blog.gopheracademy.com/advent-2014/string-matching/
						if processMatches(res, contentLower) {
							return
						}
					} else {
						res.Minified = true
					}
				}

				if !res.Minified && !res.Binary && res.Score != 0 {
					output <- res
				} else {
					if Verbose {
						if res.Binary {
							printWarn(fmt.Sprintf("skipping file identified as binary: %s", res.Location))
						} else {
							printWarn(fmt.Sprintf("skipping file due to no match: %s", res.Location))
						}
					}
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(output)
}

func processMatches(res *FileJob, contentLower string) bool {
	for i, term := range SearchString {
		// Currently only NOT does anything as the rest
		if term == "AND" || term == "OR" || term == "NOT" {
			continue
		}

		if i != 0 && SearchString[i-1] == "NOT" {
			index := bytes.Index([]byte(contentLower), []byte(term))

			// If a negated term is found we bail out instantly as
			// this means we should not be matching at all
			if index != -1 {
				res.Score = 0
				return false
			}
		} else {
			// If someone supplies ~1 at the end of the term it means we want to expand out that
			// term to support fuzzy matches for that term where the number indicates a level
			// of fuzzyness
			if strings.HasSuffix(term, "~1") || strings.HasSuffix(term, "~2") {
				terms := makeFuzzyDistanceOne(strings.TrimRight(term, "~1"))
				if strings.HasSuffix(term, "~2") {
					terms = makeFuzzyDistanceTwo(strings.TrimRight(term, "~2"))
				}

				m := []int{}
				for _, t := range terms {
					m = append(m, snippet.ExtractLocation(t, contentLower, MatchLimit)...)
				}

				if len(m) != 0 {
					res.Locations[term] = m
					res.Score = float64(len(m))
				} else {
					res.Score = 0
					return false
				}
			} else {
				// This is a regular search, not negated where we must try and find
				res.Locations[term] = snippet.ExtractLocation(term, contentLower, MatchLimit)

				if len(res.Locations[term]) != 0 {
					res.Score += float64(len(res.Locations[term]))
				} else {
					res.Score = 0
					return false
				}
			}
		}
	}

	return false
}
