codespelunker (cs)
----------------------

A command line search tool. Allows you to search over code or text files in the current directory either on
the console, via a TUI or HTTP server, using some boolean queries or regular expressions.

Consider it a similar approach to using ripgrep, silver searcher or grep coupled with fzf but in a single tool.

Dual-licensed under MIT or the UNLICENSE.

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

[![asciicast](https://asciinema.org/a/589640.svg)](https://asciinema.org/a/589640)

### Support

Using `cs` commercially? If you want priority support for `cs` you can purchase a years worth https://boyter.gumroad.com/l/vvmyi which entitles you to priority direct email support from the developer.

### Pitch

Why use cs?

 - Reasonably fast
 - Rank results on the fly helping you find things
 - Searches across multiple lines
 - Has a nice TUI interface.
 - Cross-platform (probably needs the new Windows terminal though)

The reason `cs` exists at all is because I was running into limitations using `rg TERM | fzf` and decided to solve my own
problem. 

### Install

If you want to create a package to install things please do. Let me know and ill ensure I add it here.

#### Go Get

If you have Go >= 1.20 installed

`go install github.com/boyter/cs@v1.3.0`

#### Nixos

`nix-shell -p codespelunker`

https://github.com/NixOS/nixpkgs/pull/236073

#### Manual

Binaries for Windows, GNU/Linux and macOS are available from the [releases](https://github.com/boyter/cs/releases) page.

### FAQ

#### Is this as fast as...

No.

#### You didn't let me finish, I was going to ask if it's as fast as...

The answer is probably no. It's not directly comparable. No other tool I know of works like this outside of full
indexing tools such as hound, searchcode, sourcegraph etc... None work on the fly like this does.

While `cs` does have some overlap with tools like ripgrep, grep, ack or the silver searcher the reality is it does not
work the same way, so any comparison is pointless. It is slower than most of them, but its also doing something different.

You can replicate some of what it does by piping their output into fzf though if you feel like a flawed comparison.

On my local machine which at time of writing is a Macbook Air M1 it can search a recent checkout of the linux source
code in ~2.5 seconds. While absolute performance is not a design goal, I also don't want this to be a slow tool. As such
if any obvious performance gains are on the table I will take them.

#### Does it work on normal documents?

So long as they are text. I wrote it to search code, but it works just as well on full text documents. The snippet
extraction for example was tested on Pride and Prejudice. If you had a heap of PDF's you could shell script some
use of pdftotext and get something searchable.

Note it was designed for code and as such has full .ignore and .gitignore support.

#### Where is the index?

There is none. Everything is brute force calculated on the fly. For TUI mode there are some shortcuts taken with
caching of results to speed things up.

#### How does the ranking work then?

Standard BM25 or TF/IDF or the modified TF/IDF in Lucene https://opensourceconnections.com/blog/2015/10/16/bm25-the-next-generation-of-lucene-relevation/ 
which dampens the impact of term frequency.

Technically speaking it's not accurate because it calculates the weights based on what it matched on and not everything,
but it works well enough in practice and is calculated on the fly. Try it out and report if something is not working as
you expect?

#### How do you get the snippets?

It's not fun... https://github.com/boyter/cs/blob/master/snippet.go Have a look at the code. 

It works by passing the document content to extract the snippet from and all the match locations for each term. 
It then looks through each location for each word, and checks on either side looking for terms close to it. 
It then ranks on the term frequency for the term we are checking around and rewards rarer terms. 
It also rewards more matches, closer matches, exact case matches and matches that are whole words.

For more info read the "Snippet Extraction AKA I am PHP developer" section of this blog post https://boyter.org/posts/abusing-aws-to-make-a-search-engine/

#### What does HTTP mode look like?

It's a little brutalist.

<img alt="scc" src=https://github.com/boyter/cs/raw/master/cs_http.png>

You can change its look and feel using `--template-display` and `--template-search`. See https://github.com/boyter/cs/tree/master/asset/templates
for example templates you can use to modify things.

```shell
cs -d --template-display ./asset/templates/display.tmpl --template-search ./asset/templates/search.tmpl
```

### Usage

Command line usage of `cs` is designed to be as simple as possible.
Full details can be found in `cs --help` or `cs -h`. Note that the below reflects the state of master not a release, as such
features listed below may be missing from your installation.

```
$ cs -h
code spelunker (cs) code search.
Version 1.3.0
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
 - regex with toothpick syntax /pr[e-i]de/

Searches can fuzzy match which files are searched by adding
the following syntax

 - test file:test
 - stuff filename:.go

Files that are searched will be limited to those that fuzzy
match test for the first example and .go for the second.
Example search that uses all current functionality
 - darcy NOT collins wickham~1 "ten thousand a year" /pr[e-i]de/ file:test

The default input field in tui mode supports some nano commands
- CTRL+a move to the beginning of the input
- CTRL+e move to the end of the input
- CTRL+k to clear from the cursor location forward

Usage:
  cs [flags]

Flags:
      --address string            address and port to listen to in HTTP mode (default ":8080")
      --binary                    set to disable binary file detection and search binary files
  -c, --case-sensitive            make the search case sensitive
      --dir string                directory to search, if not set defaults to current working directory
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

Searches work on single or multiple words with a logical AND applied between them. You can negate with NOT before a term.
You can do exact match with quotes, and do regular expressions using toothpicks.

Example search that uses all current functionality

```shell
cs t NOT something test~1 "ten thousand a year" "/pr[e-i]de/" file:test
```

You can use it in a similar manner to `fzf` in TUI mode if you like, since `cs` will return the matching document path
if you hit the enter key.

```shell
cat `cs`
```
