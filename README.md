codespelunker (cs)
----------------------

[![Build Status](https://travis-ci.org/boyter/cs.svg?branch=master)](https://travis-ci.org/boyter/cs)
[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

<img alt="cs" src=https://github.com/boyter/cs/raw/master/sc.gif>

```
Example search that uses all current functionality
cs t NOT something test~1 "ten thousand a year" "/pr[e-i]de/"
```

```
BUGS
search for cs --hidden --no-gitignore --no-ignore 英文 cuts in the middle of a rune
score from TF/IDF appears to be negative in some cases (overflow??)
searches in TUI for very large directories clobber each other making UI unresponsive
highlight on windows command line not escaped correctly
search for "collins lizzy mr" in TUI view does not highlight mr but does pure command line

TODO
add a snippet overlap detector to allow multiple snippets to be displayed
clean up parser so multiple spaces aren't tokens or flag em to be ignored
add proximity search "this is"~5 which means they need to be within 5 bytes of each other
add limit to number of results
investigate string match limit (might be wrong for unicode insensitive)
JSON endpoint for HTTP
Save to disk output
Actually have a real parser supporting proper boolean syntax

MAYBE

HTML parser option
https://stackoverflow.com/questions/44441665/how-to-extract-only-text-from-html-in-golang
```

template example (from root)

cs -d --template-display ./asset/templates/display.tmpl --template-search ./asset/templates/search.tmpl


searching for tab key usage with shift modifier, searched for keytab using ag/rg/ack and nothing useful
try using cs and its right at the top 


https://github.com/BurntSushi/ripgrep/issues/95

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
 
https://github.com/sourcegraph/src-cli

https://arxiv.org/pdf/1904.03061.pdf



https://www.researchgate.net/publication/4004411_Topic_extraction_from_news_archive_using_TFPDF_algorithm

A number of term-weighting schemes have derived from tf–idf. One of them is TF–PDF (Term Frequency * Proportional Document Frequency).[14] TF–PDF was introduced in 2001 in the context of identifying emerging topics in the media. The PDF component measures the difference of how often a term occurs in different domains. Another derivate is TF–IDuF. In TF–IDuF,[15] idf is not calculated based on the document corpus that is to be searched or recommended. Instead, idf is calculated on users' personal document collections. The authors report that TF–IDuF was equally effective as tf–idf but could also be applied in situations when, e.g., a user modeling system has no access to a global document corpus.



Mostly about ranking/highlighting snippet extraction links

NB problem with most snippet stuff is that it is designed to work on whole words or full word matches not partial matches such as the ones cs supports
However this is designed to work like that https://www.forrestthewoods.com/blog/reverse_engineering_sublime_texts_fuzzy_match/

https://www.hathitrust.org/blogs/large-scale-search/practical-relevance-ranking-11-million-books-part-3-document-length-normali
https://www.quora.com/How-does-BM25-work
https://github.com/apache/lucene-solr/blob/master/lucene/highlighter/src/java/org/apache/lucene/search/uhighlight/UnifiedHighlighter.java
https://lucene.apache.org/core/7_0_0/highlighter/org/apache/lucene/search/vectorhighlight/package-summary.html
https://www.compose.com/articles/how-scoring-works-in-elasticsearch/
https://blog.softwaremill.com/6-not-so-obvious-things-about-elasticsearch-422491494aa4
https://github.com/elastic/elasticsearch/blob/master/docs/reference/search/request/highlighting.asciidoc#unified-highlighter
https://www.elastic.co/guide/en/elasticsearch/reference/6.8/search-request-highlighting.html#unified-highlighter
http://www.public.asu.edu/~candan/papers/wi07.pdf
https://faculty.ist.psu.edu/jessieli/Publications/WWW10-ZLi-KeywordExtract.pdf
https://www.researchgate.net/publication/221299008_Fast_generation_of_result_snippets_in_web_search
https://arxiv.org/pdf/1904.03061.pdf
https://web.archive.org/web/20141230232527/http://rcrezende.blogspot.com/2010/08/smallest-relevant-text-snippet-for.html
https://stackoverflow.com/questions/282002/c-sharp-finding-relevant-document-snippets-for-search-result-display
https://stackoverflow.com/questions/2829303/given-a-document-select-a-relevant-snippet


$ cs "ten thousand a year" && cs "Ten thousand a year"
prideandprejudice.txt (-1.386)
…  features, noble mien, and the report which was in general
      circulation within five minutes after his entrance, of his having
      ten thousand a year. The gentlemen pronounced him to be a fine
      figure of a man, the ladies declared he was much handsomer than
      Mr. Bingley, and h…

prideandprejudice.txt (-1.386)
…before. I hope he will overlook
      it. Dear, dear Lizzy. A house in town! Every thing that is
      charming! Three daughters married! Ten thousand a year! Oh, Lord!
      What will become of me. I shall go distracted.”

      This was enough to prove that her approbation need not be
   …