package search

// Document represents a single file to be searched.
type Document struct {
	Path       string
	Filename   string
	Language   string
	Extension  string
	Content    []byte
	Complexity int64
}

// SearchResult holds the outcome of a search operation.
// It includes the matching documents and any informational notices
// generated during the query parsing and transformation phases.
type SearchResult struct {
	Documents        []*Document
	Notices          []string
	TermsToHighlight []string
}
