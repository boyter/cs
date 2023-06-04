// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/boyter/gocodewalker"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

var dirFilePaths = []string{}
var searchToFileMatchesCache = map[string][]string{}

func FindFiles(query string) chan *gocodewalker.File {
	// get the keys for the cache
	var keys []string
	for k, _ := range searchToFileMatchesCache {
		keys = append(keys, k)
	}

	// clear from the cache anything longer than the search since its will not help us
	for _, k := range keys {
		if len(k) > len(query) || query[0] != k[0] { // if cached is longer OR the first char does not match...
			delete(searchToFileMatchesCache, k)
		}
	}

	// check if the files we expect are in the cache...
	files := make(chan *gocodewalker.File, 100000)
	for i := len(query); i > 0; i-- {
		r, ok := searchToFileMatchesCache[query[:i]]
		if ok {
			go func() {
				for _, f := range r {
					files <- &gocodewalker.File{Location: f}
				}
				close(files)
			}()
			return files
		}
	}

	return walkFiles()
}

func walkFiles() chan *gocodewalker.File {
	// Now we need to run through every file closed by the filewalker when done
	fileListQueue := make(chan *gocodewalker.File, 1000)

	if FindRoot {
		dirFilePaths[0] = gocodewalker.FindRepositoryRoot(dirFilePaths[0])
	}

	fileWalker := gocodewalker.NewFileWalker(dirFilePaths[0], fileListQueue)
	fileWalker.AllowListExtensions = AllowListExtensions
	fileWalker.IgnoreIgnoreFile = IgnoreIgnoreFile
	fileWalker.IgnoreGitIgnore = IgnoreGitIgnore
	fileWalker.LocationExcludePattern = LocationExcludePattern

	go func() { _ = fileWalker.Start() }()

	return fileListQueue
}

// Reads the supplied file into memory, but only up to a certain size
func readFileContent(f *gocodewalker.File) []byte {
	fi, err := os.Lstat(f.Location)
	if err != nil {
		return nil
	}

	var content []byte

	// Only read up to point of a file because anything beyond that is probably pointless
	if fi.Size() < MaxReadSizeBytes {
		var err error
		content, err = os.ReadFile(f.Location)
		if err != nil {
			return nil
		}
	} else {
		fil, err := os.Open(f.Location)
		if err != nil {
			return nil
		}
		defer fil.Close()

		byteSlice := make([]byte, MaxReadSizeBytes)
		_, err = fil.Read(byteSlice)
		if err != nil {
			return nil
		}

		content = byteSlice
	}

	return content
}

// Given a file to read will read the contents into memory and determine if we should process it
// based on checks such as if its binary or minified
func processFile(f *gocodewalker.File) ([]byte, error) {
	content := readFileContent(f)

	if len(content) == 0 {
		if Verbose {
			fmt.Println(fmt.Sprintf("empty file so moving on %s", f.Location))
		}
		return nil, errors.New("empty file so moving on")
	}

	// Check if this file is binary by checking for nul byte and if so bail out
	// this is how GNU Grep, git and ripgrep binaryCheck for binary files
	if !IncludeBinaryFiles {
		isBinary := false

		binaryCheck := content
		if len(binaryCheck) > 10_000 {
			binaryCheck = content[:10_000]
		}
		if bytes.IndexByte(binaryCheck, 0) != -1 {
			isBinary = true
		}

		if isBinary {
			if Verbose {
				fmt.Println(fmt.Sprintf("file determined to be binary so moving on %s", f.Location))
			}
			return nil, errors.New("binary file")
		}
	}

	if !IncludeMinified {
		// Check if this file is minified
		// Check if the file is minified and if so ignore it
		split := bytes.Split(content, []byte("\n"))
		sumLineLength := 0
		for _, s := range split {
			sumLineLength += len(s)
		}
		averageLineLength := sumLineLength / len(split)

		if averageLineLength > MinifiedLineByteLength {
			if Verbose {
				fmt.Println(fmt.Sprintf("file determined to be minified so moving on %s", f.Location))
			}
			return nil, errors.New("file determined to be minified")
		}
	}

	return content, nil
}

// FileReaderWorker reads files from disk in parallel
type FileReaderWorker struct {
	input            chan *gocodewalker.File
	output           chan *FileJob
	fileCount        int64 // Count of the number of files that have been read
	InstanceId       int
	MaxReadSizeBytes int64
}

func NewFileReaderWorker(input chan *gocodewalker.File, output chan *FileJob) *FileReaderWorker {
	return &FileReaderWorker{
		input:            input,
		output:           output,
		fileCount:        0,
		MaxReadSizeBytes: 1_000_000, // sensible default of 1MB
	}
}

func (f *FileReaderWorker) GetFileCount() int64 {
	return atomic.LoadInt64(&f.fileCount)
}

// Start is responsible for spinning up jobs
// that read files from disk into memory
func (f *FileReaderWorker) Start() {
	var wg sync.WaitGroup
	for i := 0; i < maxInt(2, runtime.NumCPU()); i++ {
		wg.Add(1)
		go func() {
			for res := range f.input {
				fil, err := processFile(res)
				if err == nil {
					atomic.AddInt64(&f.fileCount, 1)
					f.output <- &FileJob{
						Filename:       res.Filename,
						Extension:      "",
						Location:       res.Location,
						Content:        fil,
						Bytes:          len(fil),
						Score:          0,
						MatchLocations: map[string][][]int{},
					}
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(f.output)
}
