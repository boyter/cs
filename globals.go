// SPDX-License-Identifier: MIT OR Unlicense

package main

// Flags set via the CLI which control how the output is displayed

// Verbose enables verbose logging output
var Verbose = false

// Debug enables debug logging output
var Debug = false

// Trace enables trace logging output which is extremely verbose
var Trace = false

// Trace enables error logging output
var Error = true

// Include minified files
var IncludeMinified = false

// Number of bytes per average line to determine file is minified
var MinifiedLineByteLength = 255

// Maximum depth to read into any text file
var MaxReadSizeBytes int64 = 1_000_000

// Disables .gitignore checks
var IgnoreGitIgnore = false

// Disables ignore file checks
var IgnoreIgnoreFile = false

// IncludeBinaryFiles toggles checking for binary files using NUL bytes
var IncludeBinaryFiles = false

// Format sets the output format of the formatter
var Format = ""

// Ranker sets which ranking algorithm to use
var Ranker = "bm25" // seems to be the best default

// FileOutput sets the file that output should be written to
var FileOutput = ""

// PathExclude sets the paths that should be skipped
var PathDenylist = []string{}

// Allow ignoring files by location
var LocationExcludePattern = []string{}

// Allow including files by location
var LocationIncludePattern = []string{}

// CaseSensitive allows tweaking of case in/sensitive search
var CaseSensitive = false

// FindRoot flag to check for the root of git or hg when run from a deeper directory and search from there
var FindRoot = false

// AllowListExtensions is a list of extensions which are whitelisted to be processed
var AllowListExtensions []string

// SearchString str if set to anything is what we want to run the search for against the current directory
var SearchString []string

// SnippetLength contains many characters out of the file to display in snippets
var SnippetLength int64 = 300

// SnippetCount is the number of snippets per file to display
var SnippetCount int64 = 1

// Include hidden files and directories in search
var IncludeHidden = false

// Address is the address to listen on when in HTTP mode
var Address string = ":8080"

// HttpServer indicates if we should fork into HTTP mode or not
var HttpServer bool = false

// SearchTemplate is the location to the search page template
var SearchTemplate = ""

// DisplayTemplate is the location to the display page template
var DisplayTemplate = ""
