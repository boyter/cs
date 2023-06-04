codespelunker (cs)
----------------------

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

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


Release 1.0.0 list

 - Improve the file walking, with the cache for queries (in tui currently)
 - Write README with details about what its for and examples
 - Write blog post about it
 - Ensure all command line params are supported and working
 - resolve issue in tui view where /[cb]at/ displays incorrectly
 - ensure paging preserves extension in http mode
 - give ability to set directory via params for http mode