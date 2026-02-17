// SPDX-License-Identifier: MIT

package main

import (
	"strings"

	"github.com/boyter/scc/v3/processor"
)

// initLanguageDatabase initialises the scc language database.
// Must be called before detectLanguage or languageExtensions.
func initLanguageDatabase() {
	processor.ProcessConstants()
}

// detectLanguage returns the language name for the given filename and content
// using the scc language database.
func detectLanguage(filename string, content []byte) string {
	detected, _ := processor.DetectLanguage(filename)
	if len(detected) >= 2 {
		return processor.DetermineLanguage(filename, detected[0], detected, content)
	}
	if len(detected) == 1 {
		return detected[0]
	}
	return ""
}

// fileCodeStats detects the language and computes SCC code stats for a file
// in a single call. Returns empty language and zero stats for unrecognised files.
func fileCodeStats(filename string, content []byte) (language string, lines, code, comment, blank, complexity int64) {
	language = detectLanguage(filename, content)
	if language == "" {
		return
	}
	sccJob := &processor.FileJob{
		Filename: filename,
		Language: language,
		Content:  content,
		Bytes:    int64(len(content)),
	}
	processor.CountStats(sccJob)
	return language, sccJob.Lines, sccJob.Code, sccJob.Comment, sccJob.Blank, sccJob.Complexity
}

// languageExtensions resolves language names to file extensions using the scc
// language database. Lookup is case-insensitive. It uses the ExtensionToLanguage
// map (extension → []languageName) built by ProcessConstants to invert the mapping.
func languageExtensions(languageNames []string) []string {
	// Build set of desired language names (lowercased) for fast lookup
	wanted := make(map[string]struct{}, len(languageNames))
	for _, name := range languageNames {
		wanted[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}

	var exts []string
	for ext, langs := range processor.ExtensionToLanguage {
		for _, lang := range langs {
			if _, ok := wanted[strings.ToLower(lang)]; ok {
				exts = append(exts, ext)
				break
			}
		}
	}
	return exts
}
