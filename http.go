// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	str "github.com/boyter/go-string"
	"github.com/boyter/gocodewalker"

	"github.com/boyter/cs/pkg/common"
	"github.com/boyter/cs/pkg/ranker"
	"github.com/boyter/cs/pkg/snippet"
)

// HTTP template types — prefixed with "http" to avoid collision with TUI's searchResult.
type httpSearch struct {
	SearchTerm          string
	SnippetSize         int
	Results             []httpSearchResult
	ResultsCount        int
	RuntimeMilliseconds int64
	ProcessedFileCount  int64
	ExtensionFacet      []httpFacetResult
	Pages               []httpPageResult
	Ext                 string
}

type httpSearchResult struct {
	Title    string
	Location string
	Content  []template.HTML
	StartPos int
	EndPos   int
	Score    float64
}

type httpFileDisplay struct {
	Location            string
	Content             template.HTML
	RuntimeMilliseconds int64
}

type httpFacetResult struct {
	Title       string
	Count       int
	SearchTerm  string
	SnippetSize int
}

type httpPageResult struct {
	SearchTerm  string
	SnippetSize int
	Value       int
	Name        string
	Ext         string
}

func StartHttpServer(cfg *Config) {
	http.HandleFunc("/file/raw/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/file/raw/", "", 1)
		w.Header().Set("Content-Type", "text/plain")
		http.ServeFile(w, r, path)
	})

	http.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		startPos := tryParseInt(r.URL.Query().Get("sp"), 0)
		endPos := tryParseInt(r.URL.Query().Get("ep"), 0)

		path := strings.Replace(r.URL.Path, "/file/", "", 1)

		if strings.TrimSpace(cfg.Directory) != "" {
			path = "/" + path
		}

		content, err := os.ReadFile(path)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// Create a random str to define where the start and end of
		// our highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		coloredContent := str.HighlightString(string(content), [][]int{{startPos, endPos}}, fmtBegin, fmtEnd)
		coloredContent = html.EscapeString(coloredContent)
		coloredContent = strings.Replace(coloredContent, fmtBegin, fmt.Sprintf(`<strong id="%d">`, startPos), -1)
		coloredContent = strings.Replace(coloredContent, fmtEnd, "</strong>", -1)

		t := template.Must(template.New("display.tmpl").Parse(httpFileTemplate))
		if cfg.DisplayTemplate != "" {
			t = template.Must(template.New("display.tmpl").ParseFiles(cfg.DisplayTemplate))
		}

		err = t.Execute(w, httpFileDisplay{
			Location:            path,
			Content:             template.HTML(coloredContent),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
		})
		if err != nil {
			log.Printf("template execute error: %v", err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		query := r.URL.Query().Get("q")
		snippetLength := tryParseInt(r.URL.Query().Get("ss"), 300)
		ext := r.URL.Query().Get("ext")
		page := tryParseInt(r.URL.Query().Get("p"), 0)
		pageSize := 20

		var results []*common.FileJob
		var processedFileCount int64

		if query != "" {
			// Make a copy of config so we can adjust AllowListExtensions per request
			searchCfg := *cfg
			if len(ext) != 0 {
				searchCfg.AllowListExtensions = []string{ext}
			} else {
				searchCfg.AllowListExtensions = []string{}
			}

			ctx := context.Background()
			ch, stats := DoSearch(ctx, &searchCfg, query)

			for fj := range ch {
				results = append(results, fj)
			}

			processedFileCount = stats.TextFileCount.Load()
			results = ranker.RankResults(cfg.Ranker, int(processedFileCount), results)
		}

		// Create a random str to define where the start and end of
		// our highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		documentTermFrequency := ranker.CalculateDocumentTermFrequency(results)

		var searchResults []httpSearchResult
		extensionFacets := map[string]int{}

		// if we have more than the page size of results, lets just show the first page
		displayResults := results
		pages := httpCalculatePages(results, pageSize, query, snippetLength, ext)

		if displayResults != nil && len(displayResults) > pageSize {
			displayResults = displayResults[:pageSize]
		}
		if page != 0 && page <= len(pages) {
			end := page*pageSize + pageSize
			if end > len(results) {
				end = len(results)
			}
			displayResults = results[page*pageSize : end]
		}

		// loop over all results so we can get the facets
		for _, res := range results {
			extensionFacets[gocodewalker.GetExtension(res.Filename)] = extensionFacets[gocodewalker.GetExtension(res.Filename)] + 1
		}

		for _, res := range displayResults {
			snippets := snippet.ExtractRelevant(res, documentTermFrequency, snippetLength)
			if len(snippets) == 0 {
				continue
			}
			v3 := snippets[0]

			// We have the snippet so now we need to highlight it
			// we get all the locations that fall in the snippet length
			// and then remove the length of the snippet cut which
			// makes our location line up with the snippet size
			var l [][]int
			for _, value := range res.MatchLocations {
				for _, s := range value {
					if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
						s[0] = s[0] - v3.StartPos
						s[1] = s[1] - v3.StartPos
						l = append(l, s)
					}
				}
			}

			// We want to escape the output, so we highlight, then escape then replace
			// our special start and end strings with actual HTML
			coloredContent := v3.Content
			// If endpos = 0 don't highlight anything because it means its a filename match
			if v3.EndPos != 0 {
				coloredContent = str.HighlightString(v3.Content, l, fmtBegin, fmtEnd)
				coloredContent = html.EscapeString(coloredContent)
				coloredContent = strings.Replace(coloredContent, fmtBegin, "<strong>", -1)
				coloredContent = strings.Replace(coloredContent, fmtEnd, "</strong>", -1)
			}

			searchResults = append(searchResults, httpSearchResult{
				Title:    res.Location,
				Location: res.Location,
				Content:  []template.HTML{template.HTML(coloredContent)},
				StartPos: v3.StartPos,
				EndPos:   v3.EndPos,
				Score:    res.Score,
			})
		}

		t := template.Must(template.New("search.tmpl").Parse(httpSearchTemplate))
		if cfg.SearchTemplate != "" {
			t = template.Must(template.New("search.tmpl").ParseFiles(cfg.SearchTemplate))
		}

		err := t.Execute(w, httpSearch{
			SearchTerm:          query,
			SnippetSize:         snippetLength,
			Results:             searchResults,
			ResultsCount:        len(results),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
			ProcessedFileCount:  processedFileCount,
			ExtensionFacet:      httpCalculateExtensionFacet(extensionFacets, query, snippetLength),
			Pages:               pages,
			Ext:                 ext,
		})
		if err != nil {
			log.Printf("template execute error: %v", err)
		}
	})

	fmt.Printf("starting HTTP server on %s\n", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, nil))
}

func httpCalculateExtensionFacet(extensionFacets map[string]int, query string, snippetLength int) []httpFacetResult {
	var ef []httpFacetResult

	for k, v := range extensionFacets {
		ef = append(ef, httpFacetResult{
			Title:       k,
			Count:       v,
			SearchTerm:  query,
			SnippetSize: snippetLength,
		})
	}

	sort.Slice(ef, func(i, j int) bool {
		if ef[i].Count == ef[j].Count {
			return strings.Compare(ef[i].Title, ef[j].Title) < 0
		}
		return ef[i].Count > ef[j].Count
	})

	return ef
}

func httpCalculatePages(results []*common.FileJob, pageSize int, query string, snippetLength int, ext string) []httpPageResult {
	var pages []httpPageResult

	if len(results) == 0 {
		return pages
	}

	if len(results) <= pageSize {
		pages = append(pages, httpPageResult{
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Value:       0,
			Name:        "1",
		})
		return pages
	}

	a := 1
	if len(results)%pageSize == 0 {
		a = 0
	}

	for i := 0; i < len(results)/pageSize+a; i++ {
		pages = append(pages, httpPageResult{
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Value:       i,
			Name:        strconv.Itoa(i + 1),
			Ext:         ext,
		})
	}
	return pages
}

func tryParseInt(x string, def int) int {
	t, err := strconv.Atoi(x)
	if err != nil {
		return def
	}
	return t
}

func makeTimestampMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func makeTimestampNano() int64 {
	return time.Now().UnixNano()
}

// Inline template strings — identical to the root http_helpers.go templates.

var httpFileTemplate = `<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{ .Location }}</title>
		<style>
			strong {
				background-color: #FFFF00;
			}
			pre {
				white-space: pre-wrap;
				white-space: -moz-pre-wrap;
				white-space: -pre-wrap;
				white-space: -o-pre-wrap;
				word-wrap: break-word;
			}
			body {
				width: 80%;
				margin: 10px auto;
				display: block;
			}
		</style>
	</head>
	<body>
		<div>
		<form method="get" action="/" >
			<input type="text" name="q" value="" autofocus="autofocus" onfocus="this.select()" />
			<input type="submit" value="search" />
			<select name="ss" id="ss">
				<option value="100">100</option>
				<option value="200">200</option>
				<option selected value="300">300</option>
				<option value="400">400</option>
				<option value="500">500</option>
				<option value="600">600</option>
				<option value="700">700</option>
				<option value="800">800</option>
				<option value="900">900</option>
				<option value="1000">1000</option>
			</select>
			<small>[processed in {{ .RuntimeMilliseconds }} (ms)]</small>
		</form>
		</div>
		<div>
			<h4>{{ .Location }}</h4>
			<small>[<a href="/file/raw/{{ .Location }}">raw file</a>]</small>
			<pre>{{ .Content }}</pre>
		</div>
	</body>
</html>`

var httpSearchTemplate = `<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>{{ .SearchTerm }}</title>
		<style>
			strong {
				background-color: #FFFF00;
			}
			pre {
				white-space: pre-wrap;
				white-space: -moz-pre-wrap;
				white-space: -pre-wrap;
				white-space: -o-pre-wrap;
				word-wrap: break-word;
			}
			body {
				width: 80%;
				margin: 10px auto;
				display: block;
			}
		</style>
	</head>
	<body>
		<div>
		<form method="get" action="/" >
			<input type="text" name="q" value="{{ .SearchTerm }}" autofocus="autofocus" onfocus="this.select()" />
			<input type="submit" value="search" />
			<select name="ss" id="ss" onchange="this.form.submit()">
				<option value="100" {{ if eq .SnippetSize 100 }}selected{{ end }}>100</option>
				<option value="200" {{ if eq .SnippetSize 200 }}selected{{ end }}>200</option>
				<option value="300" {{ if eq .SnippetSize 300 }}selected{{ end }}>300</option>
				<option value="400" {{ if eq .SnippetSize 400 }}selected{{ end }}>400</option>
				<option value="500" {{ if eq .SnippetSize 500 }}selected{{ end }}>500</option>
				<option value="600" {{ if eq .SnippetSize 600 }}selected{{ end }}>600</option>
				<option value="700" {{ if eq .SnippetSize 700 }}selected{{ end }}>700</option>
				<option value="800" {{ if eq .SnippetSize 800 }}selected{{ end }}>800</option>
				<option value="900" {{ if eq .SnippetSize 900 }}selected{{ end }}>900</option>
				<option value="1000" {{ if eq .SnippetSize 1000 }}selected{{ end }}>1000</option>
			</select>
			<small>[{{ .ResultsCount }} results from {{ .ProcessedFileCount }} files in {{ .RuntimeMilliseconds }} (ms)]</small>
		</form>
		</div>
		<div style="display:flex;">
			{{if .Results -}}
			<div style="flex-grow: 4;">
				<ul>
					{{- range .Results }}
					<li>
						<h4><a href="/file/{{ .Location }}?sp={{ .StartPos }}&ep={{ .EndPos }}">{{ .Title }}</a></h4>
						{{- range .Content }}
							<pre>{{ . }}</pre>
						{{- end }}
					</li><small>[<a href="/file/{{ .Location }}?sp={{ .StartPos }}&ep={{ .EndPos }}#{{ .StartPos }}">jump to location</a>]</small>
					{{- end }}
				</ul>

				<div>
				{{- range .Pages }}
					<a href="/?q={{ .SearchTerm }}&ss={{ .SnippetSize }}&p={{ .Value }}&ext={{ .Ext }}">{{ .Name }}</a>
				{{- end }}
				</div>
			</div>
			{{- end}}
			{{if .ExtensionFacet }}
			<div style="flex-grow: 1;">
				<h4>extensions</h4>
				<ol>
				{{- range .ExtensionFacet }}
					<li value="{{ .Count }}"><a href="/?q={{ .SearchTerm }}&ss={{ .SnippetSize }}&ext={{ .Title }}">{{ .Title }}</a></li>
				{{- end }}
				</ol>
			</div>
			{{- end }}
		</div>
	</body>
</html>`
