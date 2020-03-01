package processor

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

// PathExclude sets the paths that should be skipped
var PathDenylist = []string{}

// Allow ignoring files by location
var LocationExcludePattern = []string{}

// Allow turning on case sensitive search
var CaseSensitive = false

// Allow turning on case sensitive search
var FindRoot = false

// AllowListExtensions is a list of extensions which are whitelisted to be processed
var AllowListExtensions = []string{}

// Search string if set to anything is what we want to run the search for against the current directory
var SearchString = []string{}

var SearchBytes = [][]byte{}

// Number of results to process before bailing out
var ResultLimit int64 = 0

// How many characters out of the file to display in snippets
var SnippetLength int64 = 0

var Address string = ":8080"
var HttpServer bool = false