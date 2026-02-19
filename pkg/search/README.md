
# Go In-Memory Search Engine

This project is a robust, extensible, in-memory search engine written in Go. It supports a rich text-based query language that combines boolean keyword search with structured metadata filtering. The engine is designed to be self-contained with minimal dependencies, featuring a "self-healing" parser that gracefully handles common user syntax errors.

## Features

-   **Rich Query Syntax**:
    -   **Boolean Logic**: `AND`, `OR`, `NOT` operators. `AND` is the default operator between terms (e.g., `cat dog` is the same as `cat AND dog`).
    -   **Operator Precedence**: NOT (tightest) > AND > OR (loosest). `a OR b AND c` parses as `a OR (b AND c)`. `a OR b NOT path:vendor` parses as `a OR (b AND NOT path:vendor)`. Use parentheses to override: `(a OR b) NOT path:vendor`.
    -   **Grouping**: Use parentheses `()` for controlling order of operations.
    -   **Phrase Search**: Exact phrases using double quotes (e.g., `"lazy fox"`).
    -   **Regex Search**: Pattern matching using `/.../` syntax (e.g., `/[cb]at/`).
-   **Structured Metadata Filtering**:
    -   Filter on metadata fields using operators: `=`, `!=`, `>=`, `<=`.
    -   Multi-value filtering with commas: `lang=go,python`.
    -   Example: `complexity>=5 AND lang=go`.
-   **Semantic Aliases**:
    -   Define user-friendly aliases that are transformed into concrete queries.
    -   Example: `complexity=high` is automatically rewritten to `complexity>=8`.
-   **Robust "Self-Healing" Parser**:
    -   Automatically corrects common syntax errors and informs the user.
    -   Handles dangling operators (`cat AND`).
    -   Handles mismatched parentheses (`(cat OR dog` or `cat)`).
-   **Optimized Execution**:
    -   A simplified query planner reorders clauses to execute the most restrictive filters first, ensuring better performance.

## How to Use

This is a library package. Import it and use the `SearchEngine` API:

```go
import "github.com/boyter/cs/pkg/search"

// Create documents to search over.
docs := []*search.Document{
    {
        Path:       "/src/main.go",
        Filename:   "main.go",
        Language:   "Go",
        Extension:  "go",
        Content:    []byte("package main\nfunc main() {}"),
        Complexity: 2,
    },
}

// Create a search engine and run a query.
engine := search.NewSearchEngine(docs)
result, err := engine.Search("main AND lang=go", false) // false = case-insensitive
if err != nil {
    log.Fatal(err)
}

for _, doc := range result.Documents {
    fmt.Println(doc.Path)
}
// result.Notices contains any parser warnings/corrections.
// result.TermsToHighlight contains terms for highlighting in output.
```

## Architecture and Flow

The search engine processes a query through a multi-stage pipeline. This decoupled architecture makes the system easier to understand, maintain, and extend.

### Architectural Diagram

```
+---------------+      +-------------+      +-------------------+      +-------------+      +------------------+      +------------------+      +------------------+
|  Query String |----->|    Lexer    |----->|      Parser       |----->| Transformer |----->|     Planner     |----->|     Executor     |----->|    Extractor     |
| "cat AND > 5" |      | (Tokens)    |      | (Initial AST)     |      | (Final AST) |      | (Optimized AST) |      | (Filtering Data) |      | (Highlight Terms)|
+---------------+      +-------------+      +-------------------+      +-------------+      +------------------+      +------------------+      +------------------+
                                                                                                                                                       |
                                                                                                                                                       V
                                                                                                                                              +--------------------+
                                                                                                                                              |   Search Results   |
                                                                                                                                              +--------------------+
```

### Breakdown of Stages

1.  **Lexer (Tokenizer)** - `lexer.go`
    -   **Responsibility**: Scans the raw query string and breaks it down into a sequence of "tokens" (e.g., `KEYWORD`, `AND`, `OPERATOR`, `NUMBER`, `COMMA`). It has no understanding of grammar; it only identifies the pieces.

2.  **Parser** - `parser.go`
    -   **Responsibility**: Takes the stream of tokens from the Lexer and builds an **Abstract Syntax Tree (AST)** based on a defined grammar. The AST is a tree structure that represents the logical meaning of the query.
    -   **Key Feature**: This is where the **self-healing** logic resides. If the parser encounters a syntax error (like a missing parenthesis), it attempts to correct it and adds a `Notice` to the search result.

3.  **Transformer** - `transformer.go`
    -   **Responsibility**: Walks the initial AST and applies semantic transformations. Its job is to translate user-friendly aliases into concrete, executable filter logic.
    -   **Example**: It finds a `FilterNode` for `complexity=high` and replaces it with a new `FilterNode` for `complexity>=8`.

4.  **Planner** - `planner.go`
    -   **Responsibility**: Performs a simplified query optimization step. It analyzes the AST, specifically looking for `AND` clauses, and reorders the nodes to be more efficient.
    -   **Logic**: It reorders clauses so that low-cost operations (like metadata filters) are executed before high-cost operations (like regex searches). This dramatically reduces the amount of data that needs to be processed in later stages.

5.  **Executor** - `executor.go`
    -   **Responsibility**: This is the engine's workhorse. It walks the final, optimized AST and executes the search logic against the in-memory slice of `Document` structs.
    -   **Core Search Logic**: The `evaluate` function contains the `switch` statement that handles each node type (e.g., `AndNode`, `KeywordNode`, `FilterNode`). The actual text matching (`strings.Contains`, `regexp.Match`) happens here.

6.  **Extractor** - `extractor.go`
    -   **Responsibility**: Walks the AST to extract positive search terms for highlighting in results. It skips terms inside `NOT` subtrees so that negated terms are not highlighted.

## Built-in Filters

The engine ships with the following metadata filters, registered in `executor.go`:

| Filter Name          | Aliases              | Type    | Operators        | Multi-value | Description                         |
|----------------------|----------------------|---------|------------------|-------------|-------------------------------------|
| `complexity`         | —                    | Numeric | `=` `!=` `>=` `<=` | No       | Matches `Document.Complexity`       |
| `lang` / `language`  | Each is an alias     | String  | `=` `!=`         | Yes         | Matches `Document.Language` (case-insensitive) |
| `file` / `filename`  | Each is an alias     | String  | `=` `!=`         | No          | Matches `Document.Filename` (case-insensitive substring, or glob when `*`, `?`, `[` present) |
| `path` / `filepath`  | Each is an alias     | String  | `=` `!=`         | No          | Matches `Document.Path` (case-insensitive substring, or glob when `*`, `?`, `[` present) |
| `ext` / `extension`  | Each is an alias     | String  | `=` `!=`         | Yes         | Matches `Document.Extension` (case-insensitive) |

**Semantic alias**: `complexity=high` is rewritten to `complexity>=8`.

**Glob pattern support**: The `file` and `path` filters support glob patterns when the value contains `*`, `?`, or `[` characters. Without these characters, the existing substring matching is used for backward compatibility.

| Pattern | Meaning |
|---------|---------|
| `*` | Matches any sequence of non-separator characters |
| `?` | Matches any single non-separator character |
| `[abc]` | Matches any character in the set |

Examples:
- `file:*.go` — files ending in `.go`
- `file:*_test.go` — Go test files only
- `file:file?.py` — `file1.py`, `file2.py`, etc.
- `path:*/search/*` — files in any `search` directory
- `path:src/main/*` — files directly under `src/main/`
- `NOT path:vendor/*/*` — exclude vendor directory

For `path:` globs, a sliding-window approach matches pattern segments against contiguous segments of the file path. `**` (recursive glob) is not supported; use substring matching without glob characters for broad directory matching (e.g., `path:vendor`).

**Multi-value syntax**: Filters that support multi-value accept comma-separated lists (e.g., `lang=go,python`, `ext=ts,tsx`). This works like an `IN` clause — the document matches if its value is any of the listed values.

## How to Extend the Engine

The engine was designed to be easily extended. Here are guides for the most common modifications.

### Scenario 1: Adding a New Metadata Filter (e.g., path prefix)

Let's add the ability to filter on a path prefix, so users can write `pathprefix="/src"`.

**Step 1: Implement the Filter Handler**

In `executor.go`, create a new function that matches the `FilterHandler` signature. Each handler receives a single document and returns whether it matches.

```go
// in executor.go

// handlePathPrefixFilter filters documents by their path prefix.
func handlePathPrefixFilter(op string, val interface{}, doc *Document) bool {
	prefix, ok := val.(string)
	if !ok {
		return false
	}
	prefix = strings.ToLower(prefix)

	switch op {
	case "=":
		return strings.HasPrefix(strings.ToLower(doc.Path), prefix)
	case "!=":
		return !strings.HasPrefix(strings.ToLower(doc.Path), prefix)
	}
	return false
}
```

**Step 2: Register the New Handler**

In `executor.go`, inside the `registerFilterHandlers` function, map the field name to your new handler function.

```go
// in executor.go

func (se *SearchEngine) registerFilterHandlers() {
	se.filterHandlers = make(map[string]FilterHandler)
	se.filterHandlers["complexity"] = handleComplexityFilter
	se.filterHandlers["lang"] = handleLanguageFilter
	se.filterHandlers["language"] = handleLanguageFilter
	se.filterHandlers["file"] = handleFilenameFilter
	se.filterHandlers["filename"] = handleFilenameFilter
	se.filterHandlers["ext"] = handleExtensionFilter
	se.filterHandlers["extension"] = handleExtensionFilter
	se.filterHandlers["pathprefix"] = handlePathPrefixFilter // <-- ADD THIS LINE
}
```

**That's it!** The parser is already generic enough to understand `pathprefix="/src"`. Now the executor knows how to handle it. Add tests in `search_test.go` to verify.

### Scenario 2: Adding a New Semantic Alias (e.g., `complexity=low`)

Currently only `complexity=high` is implemented (rewritten to `complexity>=8`). Let's add `complexity=low`.

**Step 1: Implement the Transformation Logic**

In `transformer.go`, find the `transformFilterNode` function and add a new case to the `if` block.

```go
// in transformer.go

func (t *Transformer) transformFilterNode(node *FilterNode) Node {
	if node.Field == "complexity" && node.Operator == "=" {
		if val, ok := node.Value.(string); ok {
			valLower := strings.ToLower(val)

			if valLower == "high" {
				// Existing code: 'high' is defined as 8 or more
				newNode := &FilterNode{
					Field:    "complexity",
					Operator: ">=",
					Value:    8,
				}
				notice := fmt.Sprintf("Notice: '%s=%s' was interpreted as 'complexity >= 8'.", node.Field, val)
				t.notices = append(t.notices, notice)
				return newNode
			}

			// v-- ADD THIS LOGIC --v
			if valLower == "low" {
				newNode := &FilterNode{
					Field:    "complexity",
					Operator: "<=",
					Value:    3, // 'low' is defined as 3 or less
				}
				notice := fmt.Sprintf("Notice: '%s=%s' was interpreted as 'complexity <= 3'.", node.Field, val)
				t.notices = append(t.notices, notice)
				return newNode
			}
			// ^-- END OF NEW LOGIC --^
		}
	}
	return node
}
```

### Scenario 3: Implementing Fuzzy Matching

The `FuzzyNode` AST type is already defined but the executor currently has a placeholder that returns no results. Let's implement fuzzy matching using Levenshtein distance.

**Step 1: Add a Distance Function**

In `executor.go`, add a Levenshtein distance helper:

```go
// in executor.go

func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 1; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+cost))
		}
	}
	return d[la][lb]
}
```

**Step 2: Replace the Placeholder in `evaluate`**

In `executor.go`, find the `FuzzyNode` case (which currently returns `[]*Document{}`) and replace it:

```go
	case *FuzzyNode:
		var results []*Document
		content := strings.ToLower(string(doc.Content))
		term := strings.ToLower(n.Value)
		termLen := len(term)
		// Slide a window over the content to find near-matches
		for i := 0; i <= len(content)-termLen; i++ {
			window := content[i : i+termLen]
			if levenshtein(term, window) <= n.Distance {
				results = append(results, doc)
				break
			}
		}
```

## API Reference

### Types

```go
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
type SearchResult struct {
    Documents        []*Document   // Matching documents
    Notices          []string      // Parser warnings and transformation notices
    TermsToHighlight []string      // Positive search terms for highlighting
}

// FilterHandler is a function that checks whether a single document matches a filter.
type FilterHandler func(op string, val interface{}, doc *Document) bool
```

### Functions

```go
// NewSearchEngine creates a new engine with built-in filters registered.
func NewSearchEngine(docs []*Document) *SearchEngine

// Search runs a query against the engine's documents.
// caseSensitive controls whether keyword/phrase matching is case-sensitive.
func (se *SearchEngine) Search(query string, caseSensitive bool) (*SearchResult, error)

// EvaluateFile evaluates a parsed AST against a single file's content.
// Returns whether the file matches and a map of term → match locations.
func EvaluateFile(node Node, content []byte, filename string, caseSensitive bool) (bool, map[string][][]int)

// ExtractTerms traverses the AST and returns terms for highlighting,
// excluding terms inside NOT subtrees.
func ExtractTerms(node Node) []string
```

## Project Structure

```
.
├── ast.go              # Defines the Abstract Syntax Tree node types (And, Or, Not, Keyword, Phrase, Regex, Filter, Fuzzy)
├── document.go         # Defines Document and SearchResult structs
├── executor.go         # The main engine; executes the AST against data, contains filter handlers
├── extractor.go        # Walks AST to extract positive search terms for highlighting
├── lexer.go            # The tokenizer; turns strings into tokens
├── parser.go           # The parser; turns tokens into an AST, handles self-healing
├── planner.go          # The query optimizer; reorders AND clauses by cost
├── README.md           # This file
├── search_fuzz_test.go # Fuzz tests for the search engine
├── search_test.go      # Unit tests for all components
└── transformer.go      # Transforms the AST for semantic aliases
```

## How Multi-Value Filtering Works

The engine supports multi-value `IN`-style filtering with comma syntax (e.g., `lang=go,python`). This feature spans three pipeline stages:

### Lexer

The lexer recognizes `,` as a `COMMA` token (see `lexer.go`). This allows comma-separated values to be tokenized individually.

### Parser

In `parser.go`, the `parseFilterExpression` function collects values in a loop. After consuming the operator, it reads values separated by `COMMA` tokens:

- If only one value is found, `FilterNode.Value` is stored directly (e.g., `"go"` or `5`).
- If multiple values are found, `FilterNode.Value` is stored as a `[]interface{}` slice (e.g., `["go", "python"]`).

### Executor

Filter handlers use a type switch on the `val` parameter to handle both cases:

```go
func handleLanguageFilter(op string, val interface{}, doc *Document) bool {
    isEquality := (op == "=")

    switch v := val.(type) {
    case string: // Single value: lang=go
        lang := strings.ToLower(v)
        return (strings.ToLower(doc.Language) == lang) == isEquality
    case []interface{}: // Multiple values: lang=go,python
        valueSet := make(map[string]bool)
        for _, item := range v {
            if strItem, ok := item.(string); ok {
                valueSet[strings.ToLower(strItem)] = true
            }
        }
        _, exists := valueSet[strings.ToLower(doc.Language)]
        return exists == isEquality
    }
    return false
}
```

With `=`, the document matches if its value is **any** of the listed values. With `!=`, the document matches if its value is **none** of the listed values.
