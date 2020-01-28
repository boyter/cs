package main

import (
	"github.com/boyter/cs/processor"
	sccprocessor "github.com/boyter/scc/processor"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	//f, _ := os.Create("sc.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	rootCmd := &cobra.Command{
		Use:     "cs",
		Long:    "cs code search command line.\nBen Boyter <ben@boyter.org>",
		Version: "0.0.4",
		Run: func(cmd *cobra.Command, args []string) {
			processor.SearchString = args
			sccprocessor.ProcessConstants()

			if len(processor.SearchString) != 0 {
				processor.Process()
			} else {
				processor.Error = false // suppress writing errors in TUI mode
				processor.ProcessTui(true)
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.BoolVar(
		&processor.DisableCheckBinary,
		"binary",
		false,
		"disable binary file detection",
	)
	flags.BoolVar(
		&processor.Ignore,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&processor.GitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.BoolVar(
		&processor.Debug,
		"debug",
		false,
		"enable debug output",
	)
	flags.Int64VarP(
		&processor.ResultLimit,
		"limit",
		"l",
		100,
		"number of matching results to process",
	)
	flags.Int64VarP(
		&processor.SnippetLength,
		"snippet",
		"s",
		300,
		"number of matching results to process",
	)
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
	flags.StringSliceVarP(
		&processor.AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions [comma separated list: e.g. go,java,js]",
	)
	flags.StringArrayVarP(
		&processor.Exclude,
		"not-match",
		"M",
		[]string{},
		"ignore files and directories matching regular expression",
	)
	flags.StringVarP(
		&processor.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)
	flags.BoolVarP(
		&processor.Trace,
		"trace",
		"t",
		false,
		"enable trace output. Not recommended when processing multiple files",
	)
	flags.BoolVarP(
		&processor.Verbose,
		"verbose",
		"v",
		false,
		"verbose output",
	)
	flags.IntVar(
		&processor.MinifiedGeneratedLineByteLength,
		"min-gen-line-length",
		255,
		"number of bytes per average line for file to be considered minified or generated",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
