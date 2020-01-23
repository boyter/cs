package processor

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

// Flags set via the CLI which control how the output is displayed

// Files indicates if there should be file output or not when formatting
var Files = false


// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// Trace enables error logging output
var Error = true

// Disables .gitignore checks
var GitIgnore = false

// Disables ignore file checks
var Ignore = false

// DisableCheckBinary toggles checking for binary files using NUL bytes
var DisableCheckBinary = false

// Exclude is a regular expression which is used to exclude files from being processed
var Exclude = []string{}

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// PathDenylist sets the paths that should be skipped
var PathDenylist = []string{}

// FileListQueueSize is the queue of files found and ready to be read into memory
var FileListQueueSize = runtime.NumCPU()

// FileReadJobWorkers is the number of processes that read files off disk into memory
var FileReadJobWorkers = runtime.NumCPU() * 4

// FileReadContentJobQueueSize is a queue of files ready to be processed
var FileReadContentJobQueueSize = runtime.NumCPU()

// FileProcessJobWorkers is the number of workers that process the file collecting stats
var FileProcessJobWorkers = runtime.NumCPU()

// FileSummaryJobQueueSize is the queue used to hold processed file statistics before formatting
var FileSummaryJobQueueSize = runtime.NumCPU()

// WhiteListExtensions is a list of extensions which are whitelisted to be processed
var WhiteListExtensions = []string{}

// Search string if set to anything is what we want to run the search for against the current directory
var SearchString = []string{}

// Number of results to process before bailing out
var ResultLimit int64 = 0

// How many characters out of the file to display in snippets
var SnippetLength int64 = 0

// Clean up the input
func CleanSearchString() {
	tmp := []string{}

	for _, s := range SearchString {
		s = strings.Trim(s, " ")

		if s != "AND" && s != "OR" && s != "NOT" {
			s = strings.ToLower(s)
		}

		if s != "" {
			tmp = append(tmp, s)
		}
	}

	SearchString = tmp
}

// Process is the main entry point of the command line it sets everything up and starts running
func Process() {
	CleanSearchString()

	fileListQueue := make(chan *FileJob, FileListQueueSize)                     // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, FileReadContentJobQueueSize) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, FileSummaryJobQueueSize)         // Files ready to be summarised


	go walkDirectory(".", fileListQueue)
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
