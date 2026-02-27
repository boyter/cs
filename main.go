// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const Version = "3.1.0"

func main() {
	//f, _ := os.Create("profile.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

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
			" - OR syntax such as catch OR throw\n" +
			" - group with parentheses (cat OR dog) NOT fish\n" +
			" - note: NOT binds to next term, use () with OR\n" +
			" - regex with toothpick syntax /pr[e-i]de/\n" +
			"\n" +
			"Searches can filter which files are searched by adding\n" +
			"the following syntax\n" +
			" - file:test              (substring match on filename)\n" +
			" - filename:.go           (substring match on filename)\n" +
			" - path:pkg/search        (substring match on full file path)\n" +
			"\n" +
			"Example search that uses all current functionality\n" +
			" - darcy NOT collins wickham~1 \"ten thousand a year\" /pr[e-i]de/ file:test path:pkg\n" +
			"\n" +
			"The default input field in tui mode supports some nano commands\n" +
			"- CTRL+a move to the beginning of the input\n" +
			"- CTRL+e move to the end of the input\n" +
			"- CTRL+k to clear from the cursor location forward\n" +
			"\n" +
			"- F1 cycle ranker (simple/tfidf/bm25/structural)\n" +
			"- F2 cycle code filter (default/only-code/only-comments/only-strings/only-declarations/only-usages)\n" +
			"- F3 cycle gravity (off/low/default/logic/brain)\n" +
			"- F4 cycle noise (silence/quiet/default/loud/raw)\n",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			cfg.SearchString = args

			// Mutual exclusivity check
			count := 0
			if cfg.OnlyCode {
				count++
			}
			if cfg.OnlyComments {
				count++
			}
			if cfg.OnlyStrings {
				count++
			}
			if cfg.OnlyDeclarations {
				count++
			}
			if cfg.OnlyUsages {
				count++
			}
			if count > 1 {
				fmt.Fprintf(os.Stderr, "error: --only-code, --only-comments, --only-strings, --only-declarations, and --only-usages are mutually exclusive\n")
				os.Exit(1)
			}

			// Auto-select structural ranker when a content filter is set
			if cfg.HasContentFilter() && cfg.Ranker != "structural" {
				fmt.Fprintf(os.Stderr, "warning: --only-code/--only-comments/--only-strings requires structural ranker, setting --ranker=structural\n")
				cfg.Ranker = "structural"
			}

			if cfg.MCPServer {
				StartMCPServer(&cfg)
			} else if cfg.HttpServer {
				StartHttpServer(&cfg)
			} else if len(cfg.SearchString) != 0 {
				ConsoleSearch(&cfg)
			} else {
				p := tea.NewProgram(initialModel(&cfg), tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithOutput(os.Stderr))
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
		"structural",
		"set ranking algorithm [simple, tfidf, bm25, structural]",
	)
	flags.StringVar(
		&cfg.GravityIntent,
		"gravity",
		"default",
		"complexity gravity intent: brain (2.5), logic (1.5), default (1.0), low (0.2), off (0.0)",
	)
	flags.StringVar(
		&cfg.NoiseIntent,
		"noise",
		"default",
		"noise penalty intent: silence (0.1), quiet (0.5), default (1.0), loud (2.0), raw (off)",
	)
	flags.Float64Var(
		&cfg.TestPenalty,
		"test-penalty",
		0.4,
		"score multiplier for test files when query has no test intent (0.0-1.0, 1.0=disabled)",
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
		"snippet extraction mode: auto, snippet, lines, or grep",
	)
	flags.IntVar(
		&cfg.ResultLimit,
		"result-limit",
		-1,
		"maximum number of results to return (-1 for unlimited)",
	)
	flags.IntVar(
		&cfg.LineLimit,
		"line-limit",
		-1,
		"max matching lines per file in grep mode (-1 = unlimited)",
	)
	flags.IntVarP(
		&cfg.ContextBefore,
		"before-context",
		"B",
		0,
		"lines of context before each match (grep mode)",
	)
	flags.IntVarP(
		&cfg.ContextAfter,
		"after-context",
		"A",
		0,
		"lines of context after each match (grep mode)",
	)
	flags.IntVarP(
		&cfg.ContextAround,
		"context",
		"C",
		0,
		"lines of context before and after each match (grep mode)",
	)
	flags.BoolVar(
		&cfg.Dedup,
		"dedup",
		false,
		"collapse byte-identical search matches, keeping the highest-scored representative",
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
	flags.StringVar(
		&cfg.Color,
		"color",
		"auto",
		"color output mode [auto, always, never]",
	)
	flags.BoolVar(
		&cfg.Reverse,
		"reverse",
		false,
		"reverse the result order",
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
	flags.BoolVar(
		&cfg.OnlyStrings,
		"only-strings",
		false,
		"only rank matches in string literals (auto-selects structural ranker)",
	)
	flags.BoolVar(
		&cfg.OnlyDeclarations,
		"only-declarations",
		false,
		"only show matches on declaration lines (func, type, var, const, class, def, etc.)",
	)
	flags.BoolVar(
		&cfg.OnlyUsages,
		"only-usages",
		false,
		"only show matches on usage lines (excludes declarations)",
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
