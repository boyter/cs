// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const Version = "2.0.0"

func main() {
	cfg := DefaultConfig()

	initLanguageDatabase()

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
			cfg.SearchString = args

			// Auto-select structural ranker when only-code or only-comments is set
			if (cfg.OnlyCode || cfg.OnlyComments) && cfg.Ranker != "structural" {
				fmt.Fprintf(os.Stderr, "warning: --only-code/--only-comments requires structural ranker, setting --ranker=structural\n")
				cfg.Ranker = "structural"
			}

			if cfg.MCPServer {
				StartMCPServer(&cfg)
			} else if cfg.HttpServer {
				StartHttpServer(&cfg)
			} else if len(cfg.SearchString) != 0 {
				ConsoleSearch(&cfg)
			} else {
				p := tea.NewProgram(initialModel(&cfg), tea.WithAltScreen(), tea.WithOutput(os.Stderr))
				m, err := p.Run()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if fm, ok := m.(model); ok && fm.chosen != "" {
					fmt.Println(fm.chosen)
				}
			}
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.BoolVar(
		&cfg.IncludeBinaryFiles,
		"binary",
		false,
		"set to disable binary file detection and search binary files",
	)
	flags.BoolVar(
		&cfg.IgnoreIgnoreFile,
		"no-ignore",
		false,
		"disables .ignore file logic",
	)
	flags.BoolVar(
		&cfg.IgnoreGitIgnore,
		"no-gitignore",
		false,
		"disables .gitignore file logic",
	)
	flags.IntVarP(
		&cfg.SnippetLength,
		"snippet-length",
		"n",
		300,
		"size of the snippet to display",
	)
	flags.IntVarP(
		&cfg.SnippetCount,
		"snippet-count",
		"s",
		1,
		"number of snippets to display",
	)
	flags.BoolVar(
		&cfg.IncludeHidden,
		"hidden",
		false,
		"include hidden files",
	)
	flags.StringSliceVarP(
		&cfg.AllowListExtensions,
		"include-ext",
		"i",
		[]string{},
		"limit to file extensions (N.B. case sensitive) [comma separated list: e.g. go,java,js,C,cpp]",
	)
	flags.StringSliceVarP(
		&cfg.LanguageTypes,
		"type",
		"t",
		[]string{},
		"limit to language types [comma separated list: e.g. Go,Java,Python]",
	)
	flags.BoolVarP(
		&cfg.FindRoot,
		"find-root",
		"r",
		false,
		"attempts to find the root of this repository by traversing in reverse looking for .git or .hg",
	)
	flags.StringSliceVar(
		&cfg.PathDenylist,
		"exclude-dir",
		[]string{".git", ".hg", ".svn"},
		"directories to exclude",
	)
	flags.BoolVarP(
		&cfg.CaseSensitive,
		"case-sensitive",
		"c",
		false,
		"make the search case sensitive",
	)
	flags.StringSliceVarP(
		&cfg.LocationExcludePattern,
		"exclude-pattern",
		"x",
		[]string{},
		"file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]",
	)
	flags.BoolVar(
		&cfg.IncludeMinified,
		"min",
		false,
		"include minified files",
	)
	flags.IntVar(
		&cfg.MinifiedLineByteLength,
		"min-line-length",
		255,
		"number of bytes per average line for file to be considered minified",
	)
	flags.Int64Var(
		&cfg.MaxReadSizeBytes,
		"max-read-size-bytes",
		1_000_000,
		"number of bytes to read into a file with the remaining content ignored",
	)
	flags.StringVarP(
		&cfg.Format,
		"format",
		"f",
		"text",
		"set output format [text, json, vimgrep]",
	)
	flags.StringVar(
		&cfg.Ranker,
		"ranker",
		"bm25",
		"set ranking algorithm [simple, tfidf, tfidf2, bm25, structural]",
	)
	flags.StringVarP(
		&cfg.FileOutput,
		"output",
		"o",
		"",
		"output filename (default stdout)",
	)
	flags.StringVar(
		&cfg.Directory,
		"dir",
		"",
		"directory to search, if not set defaults to current working directory",
	)
	flags.StringVar(
		&cfg.SnippetMode,
		"snippet-mode",
		"auto",
		"snippet extraction mode: auto, snippet, or lines",
	)
	flags.IntVar(
		&cfg.ResultLimit,
		"result-limit",
		-1,
		"maximum number of results to return (-1 for unlimited)",
	)
	flags.BoolVar(
		&cfg.MCPServer,
		"mcp",
		false,
		"start as an MCP (Model Context Protocol) server over stdio",
	)
	flags.BoolVarP(
		&cfg.HttpServer,
		"http-server",
		"d",
		false,
		"start the HTTP server",
	)
	flags.StringVar(
		&cfg.Address,
		"address",
		":8080",
		"address and port to listen on",
	)
	flags.StringVar(
		&cfg.SearchTemplate,
		"template-search",
		"",
		"path to a custom search template",
	)
	flags.StringVar(
		&cfg.DisplayTemplate,
		"template-display",
		"",
		"path to a custom display template",
	)
	flags.StringVar(
		&cfg.TemplateStyle,
		"template-style",
		"dark",
		"built-in theme for the HTTP server UI [dark, light, bare]",
	)
	flags.BoolVar(
		&cfg.NoSyntax,
		"no-syntax",
		false,
		"disable syntax highlighting in output",
	)
	flags.Float64Var(
		&cfg.WeightCode,
		"weight-code",
		1.0,
		"structural ranker: weight for matches in code (default 1.0)",
	)
	flags.Float64Var(
		&cfg.WeightComment,
		"weight-comment",
		0.2,
		"structural ranker: weight for matches in comments (default 0.2)",
	)
	flags.Float64Var(
		&cfg.WeightString,
		"weight-string",
		0.5,
		"structural ranker: weight for matches in strings (default 0.5)",
	)
	flags.BoolVar(
		&cfg.OnlyCode,
		"only-code",
		false,
		"only rank matches in code (auto-selects structural ranker)",
	)
	flags.BoolVar(
		&cfg.OnlyComments,
		"only-comments",
		false,
		"only rank matches in comments (auto-selects structural ranker)",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
