# codespelunker (cs)

### CLI code search tool that understands code structure and ranks results by relevance. No indexing required

Ever searched for `authenticate` and gotten 200 results from config files, comments, and test stubs before finding the actual implementation? `cs` fixes that.

It combines the speed of CLI tools with the relevance ranking usually reserved for heavy indexed search engines like Sourcegraph or Zoekt, but without needing to maintain an index.

```shell
cs "authenticate" --gravity=brain           # Find the complex implementation, not the interface
cs "FIXME OR TODO OR HACK" --only-comments  # Search only in comments, not code or strings
cs "error" --only-strings                   # Find where error messages are defined
cs "handleRequest" --only-declarations      # Jump straight to where it's defined
cs "handleRequest" --only-usages            # Find every call site, skip the definition
cs "error" --dedup                          # Collapse duplicated matches into one result
```

Licensed under MIT.

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

[//]: # ([![asciicast]&#40;https://asciinema.org/a/589640.svg&#41;]&#40;https://asciinema.org/a/589640&#41;)


<img alt="cs tui" src=https://github.com/boyter/cs/raw/master/cs_tui.png>

### Pitch: Why use cs?

Most search tools treat code as plain text. `cs` doesn't.

It parses every file on the fly to understand what is a comment, what is a string, and what is code 
then uses that structure to rank by relevance, not just list them by occurrence.

```shell
cs "authenticate"                    # BM25-ranked results, best match first
cs "authenticate" --gravity=brain    # Boost complex implementations over interfaces
cs "TODO" --only-comments            # Only matches inside comments
cs "error" --only-strings            # Only matches inside string literals
cs "config OR setup" lang:go         # Boolean queries with language filters
cs "handleRequest" --only-declarations  # Jump to definitions (func, class, def, etc.)
cs "config" --dedup                     # Collapse byte-identical matches
```

#### What makes it different from ripgrep or grep?

`ripgrep` is a fast text matcher. It finds lines and prints them. It's the best at what it does.

`cs` is a search engine. It finds files, ranks by relevance, extracts the best snippet, 
and shows the most relevant results. Think Sourcegraph-quality ranked search as a CLI tool, no index required.

They solve different problems. You'll probably want both.

#### Key capabilities

- Structural Awareness: A match in code ranks higher than the same word in a comment (1.0 vs 0.2) - and it's configurable. Or filter strictly: `--only-code`, `--only-comments`, `--only-strings`.
- Complexity Gravity: Uses [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) as a ranking signal. Searching for `Authenticate`? The complex implementation file ranks above the interface definition. (`--gravity=brain`)
- Smart Ranking: BM25 relevance scoring, file-location boosting, noise penalty for data blobs, and automatic test-file dampening — all on the fly with no pre-built index.
- Multiple interfaces: Console output, a built-in TUI, an HTTP server with syntax highlighting, or an MCP server for LLM tooling.

### Key Features

#### Structural Filtering

Stop grepping through false positives.
```shell
cs "database" --only-code        # Ignore matches in comments/docs
cs "FIXME" --only-comments       # Ignore matches in code/strings
cs "error" --only-strings        # Find where error messages are defined
cs "handleRequest" --only-declarations  # Jump to where it's defined (func, class, def, etc.)
cs "handleRequest" --only-usages        # Every call site, skipping the definition
```

These are mutually exclusive with `--only-code`, `--only-comments`, and `--only-strings`.

The structural ranker also uses declaration detection to boost matches that appear on declaration lines 
(e.g. `func`, `class`, `def`) over plain usages. This currently works for the following languages:

Go, Python, JavaScript, TypeScript, TSX, Rust, Java, C, C++, C#, Ruby, PHP, Kotlin, Swift,
Shell, Lua, Scala, Elixir, Haskell, Perl, Zig, Dart, Julia, Clojure, Erlang, Groovy, OCaml,
MATLAB, Powershell, Nim, Crystal, V

For unsupported languages, all matches are treated as usages and ranked by text relevance only. 
Structural filtering (`--only-code`, `--only-comments`, `--only-strings`) still works for any language recognised 
by [scc](https://github.com/boyter/scc).

#### Complexity Gravity

Find where the work happens.
```shell
cs "login" --gravity=brain       # Boosts complex files (the implementation)
cs "login" --gravity=low         # Boosts simple files (configs/interfaces)
```

#### Deduplication

Collapse byte-identical matches into a single result.
```shell
cs "Copyright" --dedup                  # One result per unique copyright notice
cs "error" --dedup                      # Skip vendored/copied duplicates
```

**Smart Ranking**
Results are sorted by BM25 (relevance), dampened by file length, and boosted by code structure. Some effort to dampen 
test files (when you are not looking for them) is taken into account as well.

**Non-Smart Ranking**
You can switch the ranking algorithm to pure BM25, TFIDF, or simple most match ranking on the fly.

### Install

If you want to create a package to install, please make it. Let me know, and I will ensure I add it here.

#### Go Get

If you have Go >= 1.25.2 installed

`go install github.com/boyter/cs@v2.0.0`

#### Nixos

`nix-shell -p codespelunker`

https://github.com/NixOS/nixpkgs/pull/236073

#### Manual

Binaries for Windows, GNU/Linux, and macOS are available from the [releases](https://github.com/boyter/cs/releases) page.

### FAQ

#### Is this as fast as...

No.

#### You didn't let me finish, I was going to ask if it's as fast as...

The answer is probably no. It's not directly comparable. No other tool I know of works like this outside of full
indexing tools such as hound, searchcode, sourcegraph etc... None work on the fly like this does.

As far as I know what `cs` does is unique for a command line tool.

`cs` runs a full lexical analysis and complexity calculation from [scc](https://github.com/boyter/scc) on every matching file. 
This is expensive compared to the raw byte-scanning of `ripgrep`, but probably not as slow as you may think.

On a modern machine (such as Apple Silicon M1), it can search and rank the entire Linux kernel source in ~2.5 seconds.

#### Does it work on normal documents?

So long as they are text. I wrote it to search code, but it works just as well on full text documents. The snippet
extraction, for example, was tested on Pride and Prejudice, a text I know more about than I probably should considering I'm male.

#### Where is the index?

There is none. Everything is brute force calculated on the fly. There is some caching to speed things up, but should in 
practice never affect the results.

#### How does the ranking work?

`cs` uses a weighted BM25 algorithm.

Standard BM25 weights matches based on "fields" (so title, body, category). `cs` generates fields dynamically 
by parsing the code syntax.
- A match in code gets full weight (1.0).
- A match in a string gets partial weight (0.5).
- A match in a comment gets lower weight (0.2).

This means a file where your search term appears in the logic will rank higher than a file where the term only appears 
in the documentation, even if the word count is the same. 

You can tweak the values as needed via the CLI, or on the fly change what fields `cs` searches.

#### What is complexity gravity?

Complexity gravity is a ranking boost that uses each file's cyclomatic complexity to influence result ordering.

In code search, the best result is usually where the logic is implemented. These files usually have higher 
algorithmic density (branches, loops, conditions). `cs` uses this so implementation files generally outrank 
data/config/interface files all things being equal. 

The `--gravity` flag accepts named intent:

| Intent    | Strength | Purpose                                 |
|-----------|----------|-----------------------------------------|
| `brain`   | 2.5      | Aggressively surface complex core logic |
| `logic`   | 1.5      | Standard boost toward complex code      |
| `default` | 1.0      | Balanced (applied when flag not set)    |
| `low`     | 0.2      | Flatten gravity, find simple boilerplate|
| `off`     | 0.0      | Pure text relevance, no complexity boost|

```shell
cs --gravity=brain "search term"   # find complex implementations
cs --gravity=off "search term"     # pure text relevance
```

#### How do you get the snippets?

It's not fun... see https://github.com/boyter/cs/blob/master/pkg/snippet/snippet.go and https://github.com/boyter/cs/blob/master/pkg/snippet/snippet_lines.go

It works by passing the document content to extract the snippet from and all the match locations for each term.
It then looks through each location for each word, and checks on either side looking for terms close to it.
It then ranks on the term frequency for the term we are checking around and rewards rarer terms.
It also rewards more matches, closer matches, exact case matches, and matches that are whole words.

For more info read the "Snippet Extraction AKA I am PHP developer" section of this blog post https://boyter.org/posts/abusing-aws-to-make-a-search-engine/

#### What does HTTP mode look like?

It's a little brutalist.

<img alt="cs http" src=https://github.com/boyter/cs/raw/master/cs_http.png>

You can change its look and feel using `--template-style` for built-in themes (`dark`, `light`, `bare`), or provide
custom templates with `--template-display` and `--template-search`. See https://github.com/boyter/cs/tree/master/asset/templates
for example templates you can use to modify the look and feel.

```shell
cs -d --template-style light
cs -d --template-display ./asset/templates/display.tmpl --template-search ./asset/templates/search.tmpl
```


### Usage

Command line usage of `cs` is designed to be as simple as possible.
Full details can be found in `cs --help` or `cs -h`. Note that the below reflects the state of master not a release, as such
features listed below may be missing from your installation.

```
$ cs -h
code spelunker (cs) code search.
Version 2.1.0
Ben Boyter <ben@boyter.org>

cs recursively searches the current directory using some boolean logic
optionally combined with regular expressions.

Works via command line where passed in arguments are the search terms
or in a TUI mode with no arguments. Can also run in HTTP mode with
the -d or --http-server flag.

Searches by default use AND boolean syntax for all terms
 - exact match using quotes "find this"
 - fuzzy match within 1 or 2 distance fuzzy~1 fuzzy~2
 - negate using NOT such as pride NOT prejudice
 - OR syntax such as catch OR throw
 - group with parentheses (cat OR dog) NOT fish
 - note: NOT binds to next term, use () with OR
 - regex with toothpick syntax /pr[e-i]de/

Searches can filter which files are searched by adding
the following syntax
 - file:test              (substring match on filename)
 - filename:.go           (substring match on filename)
 - path:pkg/search        (substring match on full file path)

Example search that uses all current functionality
 - darcy NOT collins wickham~1 "ten thousand a year" /pr[e-i]de/ file:test path:pkg

The default input field in tui mode supports some nano commands
- CTRL+a move to the beginning of the input
- CTRL+e move to the end of the input
- CTRL+k to clear from the cursor location forward

- F1 cycle ranker (simple/tfidf/bm25/structural)
- F2 cycle code filter (default/only-code/only-comments/only-strings/only-declarations/only-usages)
- F3 cycle gravity (off/low/default/logic/brain)
- F4 cycle noise (silence/quiet/default/loud/raw)

Usage:
  cs [flags]

Flags:
      --address string            address and port to listen on (default ":8080")
      --binary                    set to disable binary file detection and search binary files
  -c, --case-sensitive            make the search case sensitive
      --dedup                     collapse byte-identical search matches, keeping the highest-scored representative
      --dir string                directory to search, if not set defaults to current working directory
      --exclude-dir strings       directories to exclude (default [.git,.hg,.svn])
  -x, --exclude-pattern strings   file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]
  -r, --find-root                 attempts to find the root of this repository by traversing in reverse looking for .git or .hg
  -f, --format string             set output format [text, json, vimgrep] (default "text")
      --gravity string            complexity gravity intent: brain (2.5), logic (1.5), default (1.0), low (0.2), off (0.0) (default "default")
  -h, --help                      help for cs
      --hidden                    include hidden files
  -d, --http-server               start the HTTP server
  -i, --include-ext strings       limit to file extensions (N.B. case sensitive) [comma separated list: e.g. go,java,js,C,cpp]
      --max-read-size-bytes int   number of bytes to read into a file with the remaining content ignored (default 1000000)
      --mcp                       start as an MCP (Model Context Protocol) server over stdio
      --min                       include minified files
      --min-line-length int       number of bytes per average line for file to be considered minified (default 255)
      --no-gitignore              disables .gitignore file logic
      --no-ignore                 disables .ignore file logic
      --no-syntax                 disable syntax highlighting in output
      --noise string              noise penalty intent: silence (0.1), quiet (0.5), default (1.0), loud (2.0), raw (off) (default "default")
      --only-code                 only rank matches in code (auto-selects structural ranker)
      --only-comments             only rank matches in comments (auto-selects structural ranker)
      --only-declarations         only show matches on declaration lines (func, type, var, const, class, def, etc.)
      --only-strings              only rank matches in string literals (auto-selects structural ranker)
      --only-usages               only show matches on usage lines (excludes declarations)
  -o, --output string             output filename (default stdout)
      --ranker string             set ranking algorithm [simple, tfidf, bm25, structural] (default "structural")
      --result-limit int          maximum number of results to return (-1 for unlimited) (default -1)
  -s, --snippet-count int         number of snippets to display (default 1)
  -n, --snippet-length int        size of the snippet to display (default 300)
      --snippet-mode string       snippet extraction mode: auto, snippet, or lines (default "auto")
      --template-display string   path to a custom display template
      --template-search string    path to a custom search template
      --template-style string     built-in theme for the HTTP server UI [dark, light, bare] (default "dark")
      --test-penalty float        score multiplier for test files when query has no test intent (0.0-1.0, 1.0=disabled) (default 0.4)
  -t, --type strings              limit to language types [comma separated list: e.g. Go,Java,Python]
  -v, --version                   version for cs
      --weight-code float         structural ranker: weight for matches in code (default 1.0) (default 1)
      --weight-comment float      structural ranker: weight for matches in comments (default 0.2) (default 0.2)
      --weight-string float       structural ranker: weight for matches in strings (default 0.5) (default 0.5)
```

Searches work on single or multiple words with a logical AND applied between them. You can negate with NOT before a term.
You can combine terms with OR and use parentheses to control grouping.
You can do an exact match with quotes and do regular expressions using toothpicks.

Example searches,

```shell
cs t NOT something test~1 "ten thousand a year" "/pr[e-i]de/" file:test
cs (cat OR dog) AND NOT bird
cs path:vendor main           # search only under vendor/
cs "func main" path:cmd       # find main functions under cmd/
cs handler lang:go            # search only Go files
cs TODO lang:go,python        # search Go and Python files
cs NOT lang:go test           # search all languages except Go
cs handler complexity:>=50    # find complex files containing "handler"
cs "json" --only-code         # find "json" in code, ignoring string literals
cs "hack" --only-comments     # find "hack" in comments only
cs "func main" --only-declarations      # find main function declarations
cs "logger" --only-usages               # find where logger is called, not defined
cs "Copyright" --dedup                   # collapse identical copyright headers
```

You can use it in a similar manner to `fzf` in TUI mode if you like, since `cs` will return the matching document path
if you hit the enter key one you have highlighted a result.

```shell
cat `cs`  # cat out the matching file
vi `cs`   # edit the selected file
```

### MCP Server Mode

`cs` can run as an [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server over stdio, allowing LLM 
tools like Claude Desktop, Claude Code, Cursor, and others to use it as a code search tool.

```shell
cs --mcp --dir /path/to/codebase
```

#### Claude Desktop Configuration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "codespelunker": {
      "command": "/path/to/cs",
      "args": ["--mcp", "--dir", "/path/to/codebase"]
    }
  }
}
```

#### Claude Code Configuration

Add to your `.mcp.json`:

```json
{
  "mcpServers": {
    "codespelunker": {
      "command": "/path/to/cs",
      "args": ["--mcp", "--dir", "/path/to/codebase"]
    }
  }
}
```

#### Exposed Tools

The MCP server exposes two tools:

**`search`** — Search code files recursively with relevance ranking.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `query` | string | yes | Search query (supports boolean logic, quotes, regex, fuzzy) |
| `max_results` | number | no | Maximum results to return (default 20) |
| `snippet_length` | number | no | Snippet size in characters |
| `case_sensitive` | boolean | no | Case sensitive search |
| `include_ext` | string | no | Comma-separated file extensions (e.g. `go,js,py`) |
| `language` | string | no | Comma-separated language types (e.g. `Go,Python`) |
| `gravity` | string | no | Complexity gravity intent: `brain`, `logic`, `default`, `low`, `off` |

Results are returned as JSON with the same fields as `--format json`: filename, location, score, snippet content, match locations, language, and code statistics.

**`get_file`** — Read the contents of a file within the project directory.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `path` | string | yes | File path relative to the project directory, or absolute path within the project |
| `start_line` | number | no | 1-based start line number (reads from beginning if omitted) |
| `end_line` | number | no | 1-based end line number, inclusive (reads to end if omitted) |

Returns JSON with line-numbered file content and, for recognised source files, language, lines, code, comment, blank, and complexity fields.



### Support

Using `cs` commercially? If you want priority support for `cs` you can purchase a years worth https://boyter.gumroad.com/l/vvmyi which entitles you to priority direct email support from the developer.

If not, raise a bug report... or don't. I'm not the boss of you.