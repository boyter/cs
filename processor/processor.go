package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	"io/ioutil"
	"runtime"
	"strings"
)

var Version = "0.0.7 alpha"

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

// Fuzzy makes all searches fuzzy by default
var Fuzzy = false

// Include minified files
var IncludeMinified = false

// Number of bytes per average line to determine file is minified
var MinifiedLineByteLength = 255

// Disables .gitignore checks
var GitIgnore = false

// Disables ignore file checks
var Ignore = false

// DisableCheckBinary toggles checking for binary files using NUL bytes
var DisableCheckBinary = false

// Format sets the output format of the formatter
var Format = ""

// FileOutput sets the file that output should be written to
var FileOutput = ""

// PathDenylist sets the paths that should be skipped
var PathDenylist = []string{}

// Allow ignoring files by location
var LocationExcludePattern = []string{}

// Allow turning on case sensitive search
var CaseSensitive = false

// Allow turning on case sensitive search
var FindRoot = false

// FileReadJobWorkers is the number of processes that read files off disk into memory
var FileReadJobWorkers = runtime.NumCPU() * 4

// FileReadContentJobQueueSize is a queue of files ready to be processed
var FileReadContentJobQueueSize = runtime.NumCPU()

// FileProcessJobWorkers is the number of workers that process the file collecting stats
var FileProcessJobWorkers = runtime.NumCPU()

// FileSummaryJobQueueSize is the queue used to hold processed file statistics before formatting
var FileSummaryJobQueueSize = runtime.NumCPU()

// AllowListExtensions is a list of extensions which are whitelisted to be processed
var AllowListExtensions = []string{}

// Search string if set to anything is what we want to run the search for against the current directory
var SearchString = []string{}

var SearchBytes = [][]byte{}

// Number of results to process before bailing out
var ResultLimit int64 = 0

// How many characters out of the file to display in snippets
var SnippetLength int64 = 0

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
}

func NewProcess() Process {
	return Process{}
}

// Process is the main entry point of the command line output it sets everything up and starts running
func (process *Process) StartProcess() {
	CleanSearchString()
	fileListQueue := make(chan *FileJob)                                        // Files ready to be read from disk
	fileReadContentJobQueue := make(chan *FileJob, FileReadContentJobQueueSize) // Files ready to be processed
	fileSummaryJobQueue := make(chan *FileJob, FileSummaryJobQueueSize)         // Files ready to be summarised

	// If the user asks we should look back till we find the .git or .hg directory and start the search
	// or in case of SVN go back till we don't find it
	startDirectory := "."
	if FindRoot {
		startDirectory = file.FindRepositoryRoot(startDirectory)
	}

	go walkDirectory(startDirectory, fileListQueue)
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
