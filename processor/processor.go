// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	"io/ioutil"
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
}

func NewProcess(directory string) Process {
	return Process{
		Directory: directory,
	}
}

// Process is the main entry point of the command line output it sets everything up and starts running
func (process *Process) StartProcess() {
	CleanSearchString()
	fileListQueue := make(chan *FileJob)                                        // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, FileReadContentJobQueueSize) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, FileSummaryJobQueueSize)         // Files ready to be summarised

	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	if FindRoot {
		process.Directory = file.FindRepositoryRoot(process.Directory)
	}


	// Walk Directory -> WhiteLister -> Reader/Scanner -> Searcher -> Summarise

	//fileWalker := file.NewFileWalker(process.Directory, fileListQueue)

	go walkDirectory(process.Directory, fileListQueue)
	go FileReaderWorker(fileListQueue, fileReadContentJobQueue)
	go FileProcessorWorker(fileReadContentJobQueue, fileSummaryJobQueue)

	result := fileSummarize(fileSummaryJobQueue)

	if FileOutput == "" {
		fmt.Println(result)
	} else {
		_ = ioutil.WriteFile(FileOutput, []byte(result), 0600)
		fmt.Println("results written to " + FileOutput)
	}
}
