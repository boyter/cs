// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	Version = "1.4.0"
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
			"optionally combined with regular expressions.\n" +
			"\n" +
			"Works via command line where passed in arguments are the search terms\n" +
			"or in a TUI mode with no arguments. Can also run in HTTP mode with\n" +
			"the -d or --http-server flag.\n" +
			"\n" +
			"Searches by default use AND boolean syntax for all terms\n" +
			" - exact match using quotes \"find this\"\n" +
			" - fuzzy match within 1 or 2 distance fuzzy~1 fuzzy~2\n" +
			" - negate using NOT such as pride NOT prejudice\n" +
			" - regex with toothpick syntax /pr[e-i]de/\n" +
			"\n" +
			"Searches can fuzzy match which files are searched by adding\n" +
			"the following syntax\n" +
			"\n" +
			" - test file:test\n" +
			" - stuff filename:.go\n" +
			"\n" +
			"Files that are searched will be limited to those that fuzzy\n" +
			"match test for the first example and .go for the second." +
			"\n" +
			"Example search that uses all current functionality\n" +
			" - darcy NOT collins wickham~1 \"ten thousand a year\" /pr[e-i]de/ file:test\n" +
			"\n" +
			"The default input field in tui mode supports some nano commands\n" +
			"- CTRL+a move to the beginning of the input\n" +
			"- CTRL+e move to the end of the input\n" +
			"- CTRL+k to clear from the cursor location forward\n",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			SearchString = args

			dirFilePaths = []string{"."}
			if strings.TrimSpace(Directory) != "" {
				dirFilePaths = []string{Directory}
			}

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
	flags.StringVar(
		&Directory,
		"dir",
		"",
		"directory to search, if not set defaults to current working directory",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
