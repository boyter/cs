codespelunker (cs)
----------------------

[![Go Report Card](https://goreportcard.com/badge/github.com/boyter/cs)](https://goreportcard.com/report/github.com/boyter/cs)
[![Coverage Status](https://coveralls.io/repos/github/boyter/cs/badge.svg?branch=master)](https://coveralls.io/github/boyter/cs?branch=master)
[![Cs Count Badge](https://sloc.xyz/github/boyter/cs/)](https://github.com/boyter/cs/)

<img alt="cs" src=https://github.com/boyter/cs/raw/master/sc.gif>

```
Example search that uses all current functionality
cs t NOT something test~1 "ten thousand a year" "/pr[e-i]de/"
```


template example (from root)

```
cs -d --template-display ./asset/templates/display.tmpl --template-search ./asset/templates/search.tmpl
```


```
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
```