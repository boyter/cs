codesearch (cs)
----------------------

[![Build Status](https://travis-ci.org/boyter/cs.svg?branch=master)](https://travis-ci.org/boyter/cs)
[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

<img alt="cs" src=https://github.com/boyter/cs/raw/master/sc.gif>

example, vendor/github.com/rivo/

searching for tab key usage with shift modifier, searched for keytab using ag/rg/ack and nothing useful
try using cs and its right at the top 

1. The runaway goroutines
2. the highlight snippet issue
3. a help for tui like tig???
4. filter on filename
5. ignore minified/generated files
6. display the number of processed files
7. be able to look though the results using keys

https://github.com/BurntSushi/ripgrep/issues/95


"ten thousand a year"

interesting bug

# bboyter @ SurfaceBook2 in ~/Go/src/bitbucket.org/ on git:master x [11:28:56]
$ cs javascript
fargate/api/integration_tests/blns.json (1.980)
…86expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\xE2\\x80\\x85expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\xE2\\x80\\x82expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\x0Bexpression(javascript:alert(1…

fargate/xray-api/integration_tests/blns.json (1.980)
…86expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\xE2\\x80\\x85expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\xE2\\x80\\x82expression(javascript:alert(1)\">DEF",
    "ABC<div style=\"x:\\x0Bexpression(javascript:alert(1…

Snippet generation

https://stackoverflow.com/questions/2829303/given-a-document-select-a-relevant-snippet
https://stackoverflow.com/questions/282002/c-sharp-finding-relevant-document-snippets-for-search-result-display


https://levelup.gitconnected.com/create-your-own-expression-parser-d1f622077796


https://blog.golang.org/normalization
https://news.ycombinator.com/item?id=6806062
https://groups.google.com/forum/#!topic/golang-nuts/Il2DX4xpW3w
https://www.reddit.com/r/javascript/comments/9i455b/why_is_%C3%9Ftouppercase_equal_to_ss/
$ rg -i --debug ß
DEBUG|grep_regex::literal|/home/bboyter/.cargo/registry/src/github.com-1ecc6299db9ec823/grep-regex-0.1.5/src/literal.rs:59: literal prefixes detected: Literals { lits: [Complete(ß), Complete(ẞ)], limit_size: 250, limit_class: 10 }


 head -c200000000 /dev/urandom > 200mb.txt
 
 https://about.sourcegraph.com/blog/going-beyond-regular-expressions-with-structural-code-search/
 
 Look into using the below to read  PDF, DOC, DOCX, XML, HTML, RTF, into plain text for searching
 https://github.com/sajari/docconv
