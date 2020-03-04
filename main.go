// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package main

import (
	"github.com/boyter/cs/processor"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	//f, _ := os.Create("cs.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use: "cs",
		Long: "cs code search command line.\n" +
			"Version " + processor.Version + "\n" +
			"Ben Boyter <ben@boyter.org>",
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
				processor.ProcessTui(true)
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
		"include-binary",
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
		"snippet",
		"s",
		300,
		"number of matching results to process",
	)
	flags.BoolVar(
		&processor.IncludeHidden,
		"include-hidden",
		false,
		"include hidden files",
	)
	flags.BoolVar(
		&processor.SearchPDF,
		"pdf",
		false,
		"attempt to extract text from pdf and search the result",
	)

	flags.StringSliceVarP(
		&processor.AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions case sensitive [comma separated list: e.g. go,java,js,C,cpp,]",
	)

	//flags.BoolVar(
	//	&processor.Debug,
	//	"debug",
	//	false,
	//	"enable debug output",
	//)
	//flags.Int64VarP(
	//	&processor.ResultLimit,
	//	"limit",
	//	"l",
	//	100,
	//	"number of matching results to process",
	//)


	flags.StringSliceVar(
		&processor.PathDenylist,
		"exclude-dir",
		[]string{".git", ".hg", ".svn"},
		"directories to exclude",
	)
	flags.StringVarP(
		&processor.Format,
		"format",
		"f",
		"text",
		"set output format [text, json]",
	)

	flags.StringVarP(
		&processor.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)
	//flags.BoolVarP(
	//	&processor.Trace,
	//	"trace",
	//	"t",
	//	false,
	//	"enable trace output. Not recommended when processing multiple files",
	//)
	//flags.BoolVarP(
	//	&processor.Verbose,
	//	"verbose",
	//	"v",
	//	false,
	//	"verbose output",
	//)
	flags.BoolVar(
		&processor.IncludeMinified,
		"include-min",
		false,
		"include minified files",
	)

	flags.IntVar(
		&processor.MinifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.StringSliceVarP(
		&processor.LocationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file locations matching this pattern ignoring case will be ignored",
	)

	flags.BoolVarP(
		&processor.CaseSensitive,
		"case-sensitive",
		"c",
		false,
		"make the search case sensitive",
	)
	flags.BoolVarP(
		&processor.FindRoot,
		"find-root",
		"r",
		false,
		"attempts to find the root of this repository recursively looking for .git or .hg",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
