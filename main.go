// SPDX-License-Identifier: MIT OR Unlicense

package main

import (
	"github.com/boyter/cs/processor"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	//f, _ := os.Create("cs.pprof")
	////f2, _ := os.Create("mem.pprof")
	//pprof.StartCPUProfile(f)
	//
	//go func() {
	//	time.Sleep(time.Second * 10)
	////	pprof.WriteHeapProfile(f2)
	////	f2.Close()
	//	pprof.StopCPUProfile()
	//}()

	rootCmd := &cobra.Command{
		Use: "cs",
		Long: "code spelunker (cs) code search.\n" +
			"Version " + processor.Version + "\n" +
			"Ben Boyter <ben@boyter.org>" +
			"\n\n" +
			"cs recursively searches the current directory using boolean logic\n" +
			"optionally combined with regular expressions." +
			"\n",
		Version: processor.Version,
		Run: func(cmd *cobra.Command, args []string) {
			processor.SearchString = args
			p := processor.NewProcess(".")

			// If there are arguments we want to print straight out to the console
			// otherwise we should enter interactive tui mode
			if processor.HttpServer {
				// start HTTP server
				processor.StartHttpServer()
			} else if len(processor.SearchString) != 0 {
				p.StartProcess()
			} else {
				processor.Error = false // suppress writing errors in TUI mode
				//processor.ProcessTui(true)
				processor.NewTuiApplication()
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVar(
		&processor.Address,
		"address",
		":8080",
		"address and port to listen to in HTTP mode",
	)
	flags.BoolVarP(
		&processor.HttpServer,
		"http-server",
		"d",
		false,
		"start http server for search",
	)
	flags.BoolVar(
		&processor.IncludeBinaryFiles,
		"binary",
		false,
		"set to disable binary file detection and search binary files",
	)
	flags.BoolVar(
		&processor.IgnoreIgnoreFile,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&processor.IgnoreGitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.Int64VarP(
		&processor.SnippetLength,
		"snippet-length",
		"n",
		300,
		"size of the snippet to display",
	)
	flags.Int64VarP(
		&processor.SnippetCount,
		"snippet-count",
		"s",
		1,
		"number of snippets to display",
	)
	flags.BoolVar(
		&processor.IncludeHidden,
		"hidden",
		false,
		"include hidden files",
	)
	flags.StringSliceVarP(
		&processor.AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions (N.B. case sensitive) [comma separated list: e.g. go,java,js,C,cpp]",
	)
	flags.BoolVarP(
		&processor.FindRoot,
		"find-root",
		"r",
		false,
		"attempts to find the root of this repository by traversing in reverse looking for .git or .hg",
	)
	flags.StringSliceVar(
		&processor.PathDenylist,
		"exclude-dir",
		[]string{".git", ".hg", ".svn"},
		"directories to exclude",
	)
	flags.BoolVarP(
		&processor.CaseSensitive,
		"case-sensitive",
		"c",
		false,
		"make the search case sensitive",
	)
	flags.StringVar(
		&processor.SearchTemplate,
		"template-search",
		"",
		"path to search template for custom styling",
	)
	flags.StringVar(
		&processor.DisplayTemplate,
		"template-display",
		"",
		"path to display template for custom styling",
	)
	flags.StringSliceVarP(
		&processor.LocationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]",
	)
	flags.BoolVar(
		&processor.IncludeMinified,
		"min",
		false,
		"include minified files",
	)
	flags.IntVar(
		&processor.MinifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.Int64Var(
		&processor.MaxReadSizeBytes,
		"max-read-size-bytes",
		10000000,
		"number of bytes to read into a file with the remaining content ignored",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json, vimgrep]",
	)
	flags.StringVar(
		&processor.Ranker,
		"ranker",
		"bm25",
		"set ranking algorithm [simple, tfidf, tfidf2, bm25]",
	)
	flags.StringVarP(
		&processor.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
