package processor

import (
	"fmt"
	"github.com/boyter/cs/file"
	str "github.com/boyter/cs/string"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type search struct {
	SearchTerm   string
	SnippetCount int
	SnippetSize  int
	Results      []searchResult
}

type searchResult struct {
	Title   string
	Content template.HTML
}

func StartHttpServer() {

	// Serves up a file
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, r.URL.Path[1:])
	//})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query().Get("q")

		// If the user asks we should look back till we find the .git or .hg directory and start the search
		// or in case of SVN go back till we don't find it
		directory := "."
		if FindRoot {
			directory = file.FindRepositoryRoot(directory)
		}

		fileQueue := make(chan *file.File, 1000)                // Files ready to be read from disk NB we buffer here because CLI runs till finished or the process is cancelled
		toProcessQueue := make(chan *fileJob, runtime.NumCPU()) // Files to be read into memory for processing
		summaryQueue := make(chan *fileJob, runtime.NumCPU())   // Files that match and need to be displayed

		fileWalker := file.NewFileWalker(directory, fileQueue)
		fileWalker.PathExclude = PathDenylist
		fileWalker.EnableIgnoreFile = true

		fileReader := NewFileReaderWorker(fileQueue, toProcessQueue)

		fileSearcher := NewSearcherWorker(toProcessQueue, summaryQueue)
		fileSearcher.SearchString = strings.Split(strings.TrimSpace(query), " ")
		fileSearcher.IncludeMinified = IncludeMinified
		fileSearcher.CaseSensitive = CaseSensitive
		fileSearcher.IncludeBinary = DisableCheckBinary

		go fileWalker.Start()
		go fileReader.Start()
		go fileSearcher.Start()

		// First step is to collect results so we can rank them
		results := []*fileJob{}
		for res := range summaryQueue {
			results = append(results, res)
		}

		rankResults(int(fileReader.GetFileCount()), results)

		fmtBegin := "<strong>"
		fmtEnd := "</strong>"

		documentFrequency := calculateDocumentFrequency(results)

		searchResults := []searchResult{}

		for _, res := range results {
			v3 := extractRelevantV3(res, documentFrequency, int(SnippetLength), "…")

			// We have the snippet so now we need to highlight it
			// we get all the locations that fall in the snippet length
			// and then remove the length of the snippet cut which
			// makes out location line up with the snippet size
			l := [][]int{}
			for _, value := range res.MatchLocations {
				for _, s := range value {
					if s[0] >= v3.StartPos && s[1] <= v3.EndPos {
						s[0] = s[0] - v3.StartPos
						s[1] = s[1] - v3.StartPos
						l = append(l, s)
					}
				}
			}

			coloredContent := str.HighlightString(v3.Content, l, fmtBegin, fmtEnd)

			searchResults = append(searchResults, searchResult{
				Title:   fmt.Sprintf("%s (%.3f)", res.Location, res.Score),
				Content: template.HTML(coloredContent),
			})
		}

		fmap := template.FuncMap{
			"findreplace": func(s1 string, s2 string, s3 string) string {
				return strings.Replace(s3, s1, s2, -1)
			},
		}

		t := template.Must(template.New("search.tmpl").Funcs(fmap).Parse(`<html>
	<head>
		<style>
			strong {
				background-color: #FFFF00
			}
pre {
    white-space: pre-wrap;       /* Since CSS 2.1 */
    white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
    white-space: -pre-wrap;      /* Opera 4-6 */
    white-space: -o-pre-wrap;    /* Opera 7 */
    word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
		</style>
	</head>
	<body>
		<div>
		<form method="get" action="/" >
			<input type="text" name="q" value="{{ .SearchTerm }}" />
			<input type="submit" value="search" />
			<select name="s" id="s">
				<option value="1">1</option>
				<option value="2">2</option>
				<option value="3">3</option>
				<option value="4">4</option>
				<option value="5">5</option>
				<option value="6">6</option>
				<option value="7">7</option>
				<option value="8">8</option>
				<option value="9">9</option>
				<option value="10">10</option>
			</select>
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
		</form>
		</div>
		{{if .Results -}}
		<div>
			<ul>
    			{{- range .Results }}
				<li>
					<h3>{{ .Title }}</h3>
					<pre>{{ .Content }}</pre>
				</li>
				 {{- end }}
			</ul>
		</div>
		{{- end}}
	</body>
</html>`))

		err := t.Execute(w, search{
			SearchTerm:   query,
			SnippetCount: 0,
			SnippetSize:  0,
			Results:      searchResults,
		})

		if err != nil {
			panic(err)
		}

	})

	log.Fatal(http.ListenAndServe(":8081", nil))

	//https://tutorialedge.net/golang/creating-simple-web-server-with-golang/
	//	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	//	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	//
	//	app := handlers.Application{
	//		ErrorLog: errorLog,
	//		InfoLog:  infoLog,
	//	}
	//
	//	srv := &http.Server{
	//		Addr:     ":8080",
	//		ErrorLog: errorLog,
	//		Handler:  app.Routes(),
	//	}
	//
	//	// Use the http.ListenAndServe() function to start a new web server. We pass in
	//	// two parameters: the TCP network address to listen on (in this case ":8080")
	//	// and the servemux we just created. If http.ListenAndServe() returns an error
	//	// we use the log.Fatal() function to log the error message and exit.
	//	infoLog.Println("Starting server on :8080")
	//	err := srv.ListenAndServe()
	//	errorLog.Fatal(err)
}
