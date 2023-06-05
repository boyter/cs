// SPDX-License-Identifier: MIT OR Unlicense

package main

import "html/template"

type search struct {
	SearchTerm          string
	SnippetSize         int
	Results             []searchResult
	ResultsCount        int
	RuntimeMilliseconds int64
	ProcessedFileCount  int64
	ExtensionFacet      []facetResult
	Pages               []pageResult
	Ext                 string
}

type pageResult struct {
	SearchTerm  string
	SnippetSize int
	Value       int
	Name        string
	Ext         string
}

type searchResult struct {
	Title    string
	Location string
	Content  []template.HTML
	StartPos int
	EndPos   int
	Score    float64
}

type fileDisplay struct {
	Location            string
	Content             template.HTML
	RuntimeMilliseconds int64
}

type facetResult struct {
	Title       string
	Count       int
	SearchTerm  string
	SnippetSize int
}

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
