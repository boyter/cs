codespelunker (cs)
----------------------

A command line search tool. Allows you to search over code or text files in the current directory either on
the console, via a TUI or HTTP server, using some boolean queries or regular expressions.

Consider it a similar approach to using ripgrep, silver searcher or grep coupled with fzf but in a single tool.

Dual-licensed under MIT or the UNLICENSE.

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)


### Pitch

Why use cs?

 - Reasonably fast
 - Rank results on the fly helping you find things
 - Searches across multiple lines
 - Has a nice TUI interface.

### FAQ

#### Is this as fast as...

No.

#### You didn't let me finish, I was going to ask if it's as fast as...

The answer is probably no. It's not directly comparable. No other tool I know of works like this outside of full
indexing tools such as hound, searchcode, sourcegraph etc... None work on the fly like this does.

While `cs` does have some overlap with tools like ripgrep, grep, ack or the silver searcher the reality is it does not
work the same way, so any comparison is pointless. It is slower than most of them, but its also doing something different.

You can replicate some of what it does by piping their output into fzf though.

### Usage

Command line usage of `cs` is designed to be as simple as possible.
Full details can be found in `cs --help` or `cs -h`. Note that the below reflects the state of master not a release, as such
features listed below may be missing from your installation.

```shell
$ cs -h
code spelunker (cs) code search.
Version 1.0.0
Ben Boyter <ben@boyter.org>

cs recursively searches the current directory using some boolean logic
optionally combined with regular expressions.

Usage:
  cs [flags]

Flags:
      --address string            address and port to listen to in HTTP mode (default ":8080")
      --binary                    set to disable binary file detection and search binary files
  -c, --case-sensitive            make the search case sensitive
      --exclude-dir strings       directories to exclude (default [.git,.hg,.svn])
  -x, --exclude-pattern strings   file and directory locations matching case sensitive patterns will be ignored [comma separated list: e.g. vendor,_test.go]
  -r, --find-root                 attempts to find the root of this repository by traversing in reverse looking for .git or .hg
  -f, --format string             set output format [text, json, vimgrep] (default "text")
  -h, --help                      help for cs
      --hidden                    include hidden files
  -d, --http-server               start http server for search
  -i, --include-ext strings       limit to file extensions (N.B. case sensitive) [comma separated list: e.g. go,java,js,C,cpp]
      --max-read-size-bytes int   number of bytes to read into a file with the remaining content ignored (default 1000000)
      --min                       include minified files
      --min-line-length int       number of bytes per average line for file to be considered minified (default 255)
      --no-gitignore              disables .gitignore file logic
      --no-ignore                 disables .ignore file logic
  -o, --output string             output filename (default stdout)
      --ranker string             set ranking algorithm [simple, tfidf, tfidf2, bm25] (default "bm25")
  -s, --snippet-count int         number of snippets to display (default 1)
  -n, --snippet-length int        size of the snippet to display (default 300)
      --template-display string   path to display template for custom styling
      --template-search string    path to search template for custom styling
  -v, --version                   version for cs
```

```
Example search that uses all current functionality
cs t NOT something test~1 "ten thousand a year" "/pr[e-i]de/"
```





template example (from root)

```
cs -d --template-display ./asset/templates/display.tmpl --template-search ./asset/templates/search.tmpl
```

```shell
cs powernow_dmi_table acer
```


Release 1.0.0 list

 - Write README with details about what its for and examples
 - Write blog post about it
 - Ensure all command line params are supported and working
 - give ability to set directory via params for http mode