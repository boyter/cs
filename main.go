// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"github.com/spf13/cobra"
	"os"
)

const (
	Version = "1.1.0"
)

func main() {
	//f, _ := os.Create("profile.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use: "cs",
		Long: "code spelunker (cs) code search.\n" +
			"Version " + Version + "\n" +
			"Ben Boyter <ben@boyter.org>" +
			"\n\n" +
			"cs recursively searches the current directory using some boolean logic\n" +
			"optionally combined with regular expressions." +
			"\n",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			SearchString = args
			dirFilePaths = []string{"."}

			// If there are arguments we want to print straight out to the console
			// otherwise we should enter interactive tui mode
			if HttpServer {
				// start HTTP server
				StartHttpServer()
			} else if len(SearchString) != 0 {
				NewConsoleSearch()
			} else {
				NewTuiSearch()
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVar(
		&Address,
		"address",
		":8080",
		"address and port to listen to in HTTP mode",
	)
	flags.BoolVarP(
		&HttpServer,
		"http-server",
		"d",
		false,
		"start http server for search",
	)
	flags.BoolVar(
		&IncludeBinaryFiles,
		"binary",
		false,
		"set to disable binary file detection and search binary files",
	)
	flags.BoolVar(
		&IgnoreIgnoreFile,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&IgnoreGitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.Int64VarP(
		&SnippetLength,
		"snippet-length",
		"n",
		300,
		"size of the snippet to display",
	)
	flags.Int64VarP(
		&SnippetCount,
		"snippet-count",
		"s",
		1,
		"number of snippets to display",
	)
	flags.BoolVar(
		&IncludeHidden,
		"hidden",
		false,
		"include hidden files",
	)
	flags.StringSliceVarP(
		&AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions (N.B. case sensitive) [comma separated list: e.g. go,java,js,C,cpp]",
	)
	flags.BoolVarP(
		&FindRoot,
		"find-root",
		"r",
		false,
		"attempts to find the root of this repository by traversing in reverse looking for .git or .hg",
	)
	flags.StringSliceVar(
		&PathDenylist,
		"exclude-dir",
		[]string{".git", ".hg", ".svn"},
		"directories to exclude",
	)
	flags.BoolVarP(
		&CaseSensitive,
		"case-sensitive",
		"c",
		false,
		"make the search case sensitive",
	)
	flags.StringVar(
		&SearchTemplate,
		"template-search",
		"",
		"path to search template for custom styling",
	)
	flags.StringVar(
		&DisplayTemplate,
		"template-display",
		"",
		"path to display template for custom styling",
	)
	flags.StringSliceVarP(
		&LocationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]",
	)
	flags.BoolVar(
		&IncludeMinified,
		"min",
		false,
		"include minified files",
	)
	flags.IntVar(
		&MinifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.Int64Var(
		&MaxReadSizeBytes,
		"max-read-size-bytes",
		1_000_000,
		"number of bytes to read into a file with the remaining content ignored",
	)
	flags.StringVarP(
		&Format,
		"format",
		"f",
		"text",
		"set output format [text, json, vimgrep]",
	)
	flags.StringVar(
		&Ranker,
		"ranker",
		"bm25",
		"set ranking algorithm [simple, tfidf, tfidf2, bm25]",
	)
	flags.StringVarP(
		&FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
