// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/boyter/cs/file"
	str "github.com/boyter/cs/str"
	"github.com/rs/zerolog/log"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func StartHttpServer() {
	http.HandleFunc("/file/raw/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Replace(r.URL.Path, "/file/raw/", "", 1)

		log.Info().
			Str("unique_code", "f24a4b1d").
			Str("path", path).
			Msg("raw page")

		w.Header().Set("Content-Type", "text/plain")
		http.ServeFile(w, r, path)
		return
	})

	http.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		startPos := tryParseInt(r.URL.Query().Get("sp"), 0)
		endPos := tryParseInt(r.URL.Query().Get("ep"), 0)

		path := strings.Replace(r.URL.Path, "/file/", "", 1)

		log.Info().
			Str("unique_code", "9212b49c").
			Int("startpos", startPos).
			Int("endpos", endPos).
			Str("path", path).
			Msg("file view page")

		var content []byte
		var err error

		content, err = ioutil.ReadFile(path)

		if err != nil {
			log.Error().
				Str("unique_code", "d063c1fd").
				Int("startpos", startPos).
				Int("endpos", endPos).
				Str("path", path).
				Msg("error reading file")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}

		// Create a random str to define where the start and end of
		// out highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		coloredContent := str.HighlightString(string(content), [][]int{{startPos, endPos}}, fmtBegin, fmtEnd)

		coloredContent = html.EscapeString(coloredContent)
		coloredContent = strings.Replace(coloredContent, fmtBegin, fmt.Sprintf(`<strong id="%d">`, startPos), -1)
		coloredContent = strings.Replace(coloredContent, fmtEnd, "</strong>", -1)

		t := template.Must(template.New("display.tmpl").Parse(httpFileTemplate))

		if DisplayTemplate != "" {
			t = template.Must(template.New("display.tmpl").ParseFiles(DisplayTemplate))
		}

		err = t.Execute(w, fileDisplay{
			Location:            path,
			Content:             template.HTML(coloredContent),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
		})

		if err != nil {
			panic(err)
		}

		return
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := makeTimestampMilli()
		query := r.URL.Query().Get("q")
		snippetLength := tryParseInt(r.URL.Query().Get("ss"), 300)
		ext := r.URL.Query().Get("ext")
		page := tryParseInt(r.URL.Query().Get("p"), 0)
		pageSize := 20

		var results []*FileJob
		var fileCount int64

		log.Info().
			Str("unique_code", "1e38548a").
			Msg("search page")

		if query != "" {
			log.Info().
				Str("unique_code", "1a54b0cd").
				Str("query", query).
				Int("snippetlength", snippetLength).
				Str("ext", ext).
				Msg("search query")

			// If the user asks we should look back till we find the .git or .hg directory and start the search from there
			directory := "."
			if FindRoot {
				directory = file.FindRepositoryRoot(directory)
			}

			fileQueue := make(chan *file.File, 1000)                // Files ready to be read from disk NB we buffer here because http runs till finished or the process is cancelled
			toProcessQueue := make(chan *FileJob, runtime.NumCPU()) // Files to be read into memory for processing
			summaryQueue := make(chan *FileJob, runtime.NumCPU())   // Files that match and need to be displayed

			fileWalker := file.NewFileWalker(directory, fileQueue)
			fileWalker.PathExclude = PathDenylist
			fileWalker.IgnoreIgnoreFile = IgnoreIgnoreFile
			fileWalker.IgnoreGitIgnore = IgnoreGitIgnore
			fileWalker.IncludeHidden = IncludeHidden
			fileWalker.LocationExcludePattern = LocationExcludePattern
			fileWalker.AllowListExtensions = AllowListExtensions
			if ext != "" {
				found := false
				for _, v := range AllowListExtensions {
					if ext == v {
						found = true
					}
				}

				if len(AllowListExtensions) == 0 {
					found = true
				}

				if found {
					fileWalker.AllowListExtensions = []string{ext}
				}
			}

			fileReader := NewFileReaderWorker(fileQueue, toProcessQueue)
			fileReader.MaxReadSizeBytes = MaxReadSizeBytes

			fileSearcher := NewSearcherWorker(toProcessQueue, summaryQueue)
			fileSearcher.SearchString = strings.Split(strings.TrimSpace(query), " ")
			fileSearcher.IncludeMinified = IncludeMinified
			fileSearcher.CaseSensitive = CaseSensitive
			fileSearcher.MatchLimit = -1
			fileSearcher.IncludeBinary = IncludeBinaryFiles
			fileSearcher.MinifiedLineByteLength = MinifiedLineByteLength

			go fileWalker.Start()
			go fileReader.Start()
			go fileSearcher.Start()

			// First step is to collect results so we can rank them
			for res := range summaryQueue {
				results = append(results, res)
			}

			fileCount = fileReader.GetFileCount()
			rankResults(int(fileReader.GetFileCount()), results)
		}

		// Create a random str to define where the start and end of
		// out highlight should be which we swap out later after we have
		// HTML escaped everything
		md5Digest := md5.New()
		fmtBegin := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("begin_%d", makeTimestampNano()))))
		fmtEnd := hex.EncodeToString(md5Digest.Sum([]byte(fmt.Sprintf("end_%d", makeTimestampNano()))))

		documentTermFrequency := calculateDocumentTermFrequency(results)

		var searchResults []searchResult
		extensionFacets := map[string]int{}

		// if we have more than the page size of results, lets just show the first page
		displayResults := results
		pages := calculatePages(results, pageSize, query, snippetLength)

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
			extensionFacets[file.GetExtension(res.Filename)] = extensionFacets[file.GetExtension(res.Filename)] + 1
		}

		for _, res := range displayResults {
			v3 := extractRelevantV3(res, documentTermFrequency, snippetLength, "…")[0]

			// We have the snippet so now we need to highlight it
			// we get all the locations that fall in the snippet length
			// and then remove the length of the snippet cut which
			// makes out location line up with the snippet size
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

			searchResults = append(searchResults, searchResult{
				Title:    res.Location,
				Location: res.Location,
				Content:  []template.HTML{template.HTML(coloredContent)},
				StartPos: v3.StartPos,
				EndPos:   v3.EndPos,
				Score:    res.Score,
			})
		}

		t := template.Must(template.New("search.tmpl").Parse(httpSearchTemplate))
		if SearchTemplate != "" {
			// If we have been supplied a template then load it up
			t = template.Must(template.New("search.tmpl").ParseFiles(SearchTemplate))
		}

		err := t.Execute(w, search{
			SearchTerm:          query,
			SnippetSize:         snippetLength,
			Results:             searchResults,
			ResultsCount:        len(results),
			RuntimeMilliseconds: makeTimestampMilli() - startTime,
			ProcessedFileCount:  fileCount,
			ExtensionFacet:      calculateExtensionFacet(extensionFacets, query, snippetLength),
			Pages:               pages,
		})

		if err != nil {
			panic(err)
		}
		return
	})

	log.Info().
		Str("unique_code", "03148801").
		Str("address", Address).
		Msg("ready to serve requests")

	log.Fatal().Msg(http.ListenAndServe(Address, nil).Error())
}

func calculateExtensionFacet(extensionFacets map[string]int, query string, snippetLength int) []facetResult {
	var ef []facetResult

	for k, v := range extensionFacets {
		ef = append(ef, facetResult{
			Title:       k,
			Count:       v,
			SearchTerm:  query,
			SnippetSize: snippetLength,
		})
	}

	sort.Slice(ef, func(i, j int) bool {
		// If the same count sort by the name to ensure it's consistent on the display
		if ef[i].Count == ef[j].Count {
			return strings.Compare(ef[i].Title, ef[j].Title) < 0
		}
		return ef[i].Count > ef[j].Count
	})

	return ef
}

func calculatePages(results []*FileJob, pageSize int, query string, snippetLength int) []pageResult {
	var pages []pageResult

	if len(results) == 0 {
		return pages
	}

	if len(results) <= pageSize {
		pages = append(pages, pageResult{
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
		pages = append(pages, pageResult{
			SearchTerm:  query,
			SnippetSize: snippetLength,
			Value:       i,
			Name:        strconv.Itoa(i + 1),
		})
	}
	return pages
}
