// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	"io/ioutil"
	"runtime"
	"strings"
)

var Version = "0.0.7 alpha"

// Clean up the input to avoid searching for spaces etc...
// Take the string cut it up, lower case everything except
// boolean operators and join it all back into the same slice
func CleanSearchString() {
	tmp := [][]byte{}

	for _, s := range SearchString {
		s = strings.TrimSpace(s)

		if s != "AND" && s != "OR" && s != "NOT" {
			if !CaseSensitive {
				s = strings.ToLower(s)
			}
		}

		if s != "" {
			tmp = append(tmp, []byte(s))
		}
	}

	SearchBytes = tmp
}

type Process struct {
	Directory string // What directory are we searching
	FindRoot bool
}

func NewProcess(directory string) Process {
	return Process{
		Directory: directory,
	}
}

// Process is the main entry point of the command line output it sets everything up and starts running
func (process *Process) StartProcess() {
	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	if process.FindRoot {
		process.Directory = file.FindRepositoryRoot(process.Directory)
	}

	fileQueue := make(chan *file.File, 1000)    // Files ready to be read from disk NB we buffer here because CLI runs till finished or the process is cancelled
	toProcessQueue := make(chan *fileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *fileJob, runtime.NumCPU())   // Files that match and need to be displayed

	fileWalker := file.NewFileWalker(process.Directory, fileQueue)
	fileWalker.PathExclude = PathDenylist
	fileWalker.EnableIgnoreFile = true
	fileReader := NewFileReaderWorker(fileQueue, toProcessQueue)
	fileSearcher := NewSearcherWorker(toProcessQueue, summaryQueue)
	fileSearcher.SearchString = SearchString
	fileSearcher.MatchLimit = MatchLimit
	fileSearcher.IncludeMinified = IncludeMinified
	resultSummarizer := NewResultSummarizer(summaryQueue)
	resultSummarizer.FileReaderWorker = &fileReader

	go fileWalker.Start()
	go fileReader.Start()
	go fileSearcher.Start()
	result := resultSummarizer.Start()


	if FileOutput == "" {
		fmt.Println(result)
	} else {
		_ = ioutil.WriteFile(FileOutput, []byte(result), 0600)
		fmt.Println("results written to " + FileOutput)
	}
}
