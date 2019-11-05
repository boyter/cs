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
var StopProcessing bool

func returnEarly() bool {
	if StopProcessing || atomic.LoadInt64(&TotalCount) >= ResultLimit {
		return true
	}

	return false
}

// Reads entire file into memory and then pushes it onto the next queue
func FileReaderWorker(input chan *FileJob, output chan *FileJob) {
	//close(output)
	//return
	routineWaitGroup.Add(1)
	defer routineWaitGroup.Done()

	var startTime int64
	var wg sync.WaitGroup

	for i := 0; i < FileReadJobWorkers; i++ {
		wg.Add(1)
		go func() {
			routineWaitGroup.Add(1)
			defer routineWaitGroup.Done()
			defer wg.Done()

			for res := range input {
				if returnEarly() {
					break
				}

				atomic.CompareAndSwapInt64(&startTime, 0, makeTimestampMilli())

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
					res.Content = content
					res.Locations = map[string][]int{}
					output <- res
				} else {
					if Verbose {
						printWarn(fmt.Sprintf("error reading: %s %s", res.Location, err))
					}
				}
			}
		}()
	}

	go func() {
		// Apparently this one never finishes??
		// TODO This one never exits for some reason
		//routineWaitGroup.Add(1)
		//defer routineWaitGroup.Done()

		wg.Wait()
		close(output)

		if Debug {
			printDebug(fmt.Sprintf("milliseconds reading files into memory: %d", makeTimestampMilli()-startTime))
		}
	}()
}

// Just to work out where the goroutine leak exists
//func FileProcessorWorker(input chan *FileJob, output chan *FileJob) {
//	close(output)
//}

// Does the actual processing of stats and as such contains the hot path CPU call
func FileProcessorWorker(input chan *FileJob, output chan *FileJob) {
	routineWaitGroup.Add(1)
	defer routineWaitGroup.Done()

	var startTime int64
	var wg sync.WaitGroup

	for i := 0; i < FileProcessJobWorkers; i++ {
		wg.Add(1)
		go func() {
			routineWaitGroup.Add(1)
			defer routineWaitGroup.Done()
			defer wg.Done()

			for res := range input {
				if returnEarly() {
					return
				}

				atomic.CompareAndSwapInt64(&startTime, 0, makeTimestampMilli())
				processingStartTime := makeTimestampNano()

				if bytes.IndexByte(res.Content, '\x00') != -1 {
					res.Binary = true
				} else {
					// what we need to do is check for each term if it exists, and then use that to determine if its a match
					contentLower := strings.ToLower(string(res.Content))

					// https://blog.gopheracademy.com/advent-2014/string-matching/
					if processMatches(res, contentLower) {
						return
					}
				}

				if Trace {
					printTrace(fmt.Sprintf("nanoseconds process: %s: %d", res.Location, makeTimestampNano()-processingStartTime))
				}

				if !res.Binary && res.Score != 0 {
					atomic.AddInt64(&TotalCount, 1)
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
		}()
	}

	go func() {
		routineWaitGroup.Add(1)
		defer routineWaitGroup.Done()

		wg.Wait()
		close(output)
	}()

	if Debug {
		printDebug(fmt.Sprintf("milliseconds processing files: %d", makeTimestampMilli()-startTime))
	}
}

func processMatches(res *FileJob, contentLower string) bool {
	for i, term := range SearchString {
		if returnEarly() {
			return true
		}

		if term == "AND" || term == "OR" || term == "NOT" {
			continue
		}

		if i != 0 && SearchString[i-1] == "NOT" {
			index := bytes.Index([]byte(contentLower), []byte(term))

			// If a negated term is found we bail out instantly
			if index != -1 {
				res.Score = 0
				return false
			}
		} else {
			if strings.HasSuffix(term, "~1") || strings.HasSuffix(term, "~2") {
				terms := makeFuzzyDistanceOne(strings.TrimRight(term, "~1"))
				if strings.HasSuffix(term, "~2") {
					terms = makeFuzzyDistanceTwo(strings.TrimRight(term, "~2"))
				}

				m := []int{}
				for _, t := range terms {
					m = append(m, snippet.ExtractLocation(t, contentLower, 50)...)
				}

				if len(m) != 0 {
					res.Locations[term] = m
					res.Score = float64(len(m))
				} else {
					res.Score = 0
					return false
				}
			} else {
				res.Locations[term] = snippet.ExtractLocation(term, contentLower, 50)

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
