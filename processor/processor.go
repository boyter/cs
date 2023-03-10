// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"github.com/boyter/cs/file"
	"runtime"
)

var Version = "0.12.0 beta"

type Process struct {
	Directory string // What directory are we searching
	FindRoot  bool
}

func NewProcess(directory string) Process {
	return Process{
		Directory: directory,
	}
}

// StartProcess is the main entry point of the command line output it sets everything up and starts running
func (process *Process) StartProcess() {
	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	if process.FindRoot {
		process.Directory = file.FindRepositoryRoot(process.Directory)
	}

	fileQueue := make(chan *file.File, 1000)                // Files ready to be read from disk NB we buffer here because CLI runs till finished or the process is cancelled
	toProcessQueue := make(chan *FileJob, runtime.NumCPU()) // Files to be read into memory for processing
	summaryQueue := make(chan *FileJob, runtime.NumCPU())   // Files that match and need to be displayed

	fileWalker := file.NewFileWalker(process.Directory, fileQueue)
	fileWalker.PathExclude = PathDenylist
	fileWalker.IgnoreIgnoreFile = IgnoreIgnoreFile
	fileWalker.IgnoreGitIgnore = IgnoreGitIgnore
	fileWalker.IncludeHidden = IncludeHidden
	fileWalker.AllowListExtensions = AllowListExtensions
	fileWalker.LocationExcludePattern = LocationExcludePattern

	fileReader := NewFileReaderWorker(fileQueue, toProcessQueue)
	fileReader.MaxReadSizeBytes = MaxReadSizeBytes

	fileSearcher := NewSearcherWorker(toProcessQueue, summaryQueue)
	fileSearcher.SearchString = SearchString
	fileSearcher.IncludeMinified = IncludeMinified
	fileSearcher.CaseSensitive = CaseSensitive
	fileSearcher.IncludeBinary = IncludeBinaryFiles
	fileSearcher.MinifiedLineByteLength = MinifiedLineByteLength

	resultSummarizer := NewResultSummarizer(summaryQueue)
	resultSummarizer.FileReaderWorker = fileReader
	resultSummarizer.SnippetCount = SnippetCount

	go func() { _ = fileWalker.Start() }()
	go fileReader.Start()
	go fileSearcher.Start()
	resultSummarizer.Start()
}
