Of course. Creating solid documentation is just as important as writing the code itself. Here is a comprehensive `README.md` file that explains the architecture, the flow, and exactly how to extend the system, complete with diagrams, as we discussed.

---

# Go In-Memory Search Engine

This project is a robust, extensible, in-memory search engine written in Go. It supports a rich text-based query language that combines boolean keyword search with structured metadata filtering. The engine is designed to be self-contained with minimal dependencies, featuring a "self-healing" parser that gracefully handles common user syntax errors.

## Features

-   **Rich Query Syntax**:
    -   **Boolean Logic**: `AND`, `OR`, `NOT` operators. `AND` is the default operator between terms (e.g., `cat dog` is the same as `cat AND dog`).
    -   **Grouping**: Use parentheses `()` for controlling order of operations.
    -   **Phrase Search**: Exact phrases using double quotes (e.g., `"lazy fox"`).
    -   **Regex Search**: Pattern matching using `/.../` syntax (e.g., `/[cb]at/`).
-   **Structured Metadata Filtering**:
    -   Filter on metadata fields using operators: `=`, `!=`, `>=`, `<=`.
    -   Example: `complexity>=5 AND author="jane"`.
-   **Semantic Aliases**:
    -   Define user-friendly aliases that are transformed into concrete queries.
    -   Example: `complexity=high` is automatically rewritten to `complexity>=8`.
-   **Robust "Self-Healing" Parser**:
    -   Automatically corrects common syntax errors and informs the user.
    -   Handles dangling operators (`cat AND`).
    -   Handles mismatched parentheses (`(cat OR dog` or `cat)`).
-   **Optimized Execution**:
    -   A simplified query planner reorders clauses to execute the most restrictive filters first, ensuring better performance.

## How to Run

The project is self-contained and can be run directly from the source.

1.  Ensure all the Go files (`main.go`, `document.go`, `ast.go`, `lexer.go`, `parser.go`, `transformer.go`, `planner.go`, `executor.go`) are in the same directory.
2.  Run the main program from your terminal:

    ```bash
    go run .
    ```

3.  This will execute the example queries defined in `main.go` and print the results to the console.

## Architecture and Flow

The search engine processes a query through a multi-stage pipeline. This decoupled architecture makes the system easier to understand, maintain, and extend.

### Architectural Diagram

```
+---------------+      +-------------+      +-------------------+      +-------------+      +------------------+      +------------------+
|  Query String |----->|    Lexer    |----->|      Parser       |----->| Transformer |----->|     Planner     |----->|     Executor     |
| "cat AND > 5" |      | (Tokens)    |      | (Initial AST)     |      | (Final AST) |      | (Optimized AST) |      | (Filtering Data) |
+---------------+      +-------------+      +-------------------+      +-------------+      +------------------+      +------------------+
                                                                                                                           |
                                                                                                                           |
                                                                                                                           V
                                                                                                                +--------------------+
                                                                                                                |   Search Results   |
                                                                                                                +--------------------+
```

### Breakdown of Stages

1.  **Lexer (Tokenizer)** - `lexer.go`
    -   **Responsibility**: Scans the raw query string and breaks it down into a sequence of "tokens" (e.g., `KEYWORD`, `AND`, `OPERATOR`, `NUMBER`). It has no understanding of grammar; it only identifies the pieces.

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
    -   **Core Search Logic**: The `evaluate` function contains the `switch` statement that handles each node type (e.g., `AndNode`, `KeywordNode`, `FilterNode`). The actual text matching (`strings.Contains`, `regexp.MatchString`) happens here.

## How to Extend the Engine

The engine was designed to be easily extended. Here are guides for the most common modifications.

### Scenario 1: Adding a New Metadata Filter (e.g., `author="jane"`)

Let's add the ability to filter on a new `Author` string field.

**Step 1: Update the Document Struct**

In `document.go`, add the new field to your data structure.

```go
// in document.go
type Document struct {
    ID         int
    Content    string
    Complexity int
    Author     string // <-- ADD THIS LINE
}
```

**Step 2: Implement the Filter Handler**

In `executor.go`, create a new function that contains the logic for filtering by author.

```go
// in executor.go

// handleAuthorFilter filters documents by the Author field.
// It only supports '=' and '!=' for strings.
func (se *SearchEngine) handleAuthorFilter(op string, val interface{}, docs []Document) []Document {
	authorName, ok := val.(string)
	if !ok {
		return []Document{} // Value is not a string, return no results.
	}

	var results []Document
	for _, doc := range docs {
		match := false
		switch op {
		case "=":
			if doc.Author == authorName {
				match = true
			}
		case "!=":
			if doc.Author != authorName {
				match = true
			}
		}
		if match {
			results = append(results, doc)
		}
	}
	return results
}
```

**Step 3: Register the New Handler**

In `executor.go`, inside the `registerFilterHandlers` function, map the field name "author" to your new handler function.

```go
// in executor.go

func (se *SearchEngine) registerFilterHandlers() {
	se.filterHandlers = make(map[string]FilterHandler)
	se.filterHandlers["complexity"] = se.handleComplexityFilter
	se.filterHandlers["author"] = se.handleAuthorFilter // <-- ADD THIS LINE
}
```

**That's it!** The parser is already generic enough to understand `author="jane"`. Now the executor knows how to handle it. Remember to add `Author` data to your sample documents in `main.go` to test it.

### Scenario 2: Adding a New Semantic Alias (e.g., `complexity=low`)

**Step 1: Implement the Transformation Logic**

In `transformer.go`, find the `transformFilterNode` function and add a new case to the `if` block.

```go
// in transformer.go

func (t *Transformer) transformFilterNode(node *FilterNode) Node {
	if node.Field == "complexity" && node.Operator == "=" {
		if val, ok := node.Value.(string); ok {
			valLower := strings.ToLower(val)
			
			if valLower == "high" {
				// ... existing code for high ...
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

### Scenario 3: Changing the Core Search Logic

The core logic for how a keyword or phrase matches content is located in the `evaluate` function in `executor.go`.

For example, to change the keyword search from **case-sensitive** to **case-insensitive**:

```go
// in executor.go

func (se *SearchEngine) evaluate(node Node, docs []Document) []Document {
	// ... other cases
	case *KeywordNode:
		var results []Document
		// Convert search term to lower once
		lowerValue := strings.ToLower(n.Value) 
		for _, doc := range docs {
            // Convert content to lower for comparison
			if strings.Contains(strings.ToLower(doc.Content), lowerValue) { // <-- MODIFIED LINE
				results = append(results, doc)
			}
		}
		return results
	// ... other cases
}
```

## Project Structure

```
.
├── ast.go              # Defines the Abstract Syntax Tree structures
├── document.go         # Defines Document and SearchResult structs
├── executor.go         # The main engine; executes the AST against data
├── lexer.go            # The tokenizer; turns strings into tokens
├── main.go             # Example usage and entry point
├── parser.go           # The parser; turns tokens into an AST, handles errors
├── planner.go          # The query optimizer; reorders the AST
├── README.md           # This file
├── search_test.go      # Unit tests for all components
└── transformer.go      # Transforms the AST for semantic aliases
```



## Extending the Engine: A Practical Example

To showcase the power and extensibility of the architecture, this guide provides a step-by-step walkthrough for adding a new, advanced feature: a multi-value `IN` clause for filters.

The goal is to enable queries like `category=doc,pdf,csv`, which should match any document whose category is one of "doc", "pdf", or "csv".

This requires touching three layers of the pipeline: the **Lexer**, the **Parser**, and the **Executor**.

### Step 1: Update the Document Struct

First, we need a field to filter on. We'll add a `Category` field to our main `Document` struct.

**File: `document.go`**

```go
type Document struct {
    ID         int
    Content    string
    Complexity int
    Category   string // <-- Add this field
}
```

### Step 2: Teach the Lexer about Commas (`,`)

The lexer must be able to recognize the comma as a distinct token that separates values.

**File: `lexer.go`**

1.  Add a new `COMMA` token type to the `const` block.

    ```go
    const (
        // ... existing token types
        STRING_ALIAS
        COMMA // <-- Add this line
    )
    ```

2.  In the main `scan()` function, add a `case` to handle the comma character.

    ```go
    func (l *Lexer) scan() Token {
        // ... existing code
        switch ch {
        // ... existing cases
        case ')':
            return Token{Type: RPAREN, Literal: string(ch)}
        case ',': // <-- Add this case
            return Token{Type: COMMA, Literal: string(ch)}
        case '=', '!', '>', '<':
        // ... existing code
        }
        // ...
    }
    ```

### Step 3: Teach the Parser the List Syntax

This is the most significant change. We will upgrade the `parseFilterExpression` function to understand that a value can be a single item or a comma-separated list of items.

**File: `parser.go`**

```go
func (p *Parser) parseFilterExpression() Node {
	node := &FilterNode{Field: p.tok.Literal}
	p.nextToken() // Consume field
	node.Operator = p.tok.Literal
	p.nextToken() // Consume operator

	// Check if the next token is a valid value type.
	if p.tok.Type != NUMBER && p.tok.Type != STRING_ALIAS && p.tok.Type != KEYWORD && p.tok.Type != IDENTIFIER {
		return nil // Error: expected a value.
	}
	
	// Collect one or more values.
	var values []interface{}
	for {
		// Add the current value to our list.
		switch p.tok.Type {
		case NUMBER:
			val, _ := strconv.Atoi(p.tok.Literal)
			values = append(values, val)
		case STRING_ALIAS, KEYWORD, IDENTIFIER:
			values = append(values, p.tok.Literal)
		}
		p.nextToken() // Consume the value token

		// If the next token is not a comma, the list is finished.
		if p.tok.Type != COMMA {
			break
		}
		p.nextToken() // Consume the comma and loop again.
	}

	// If we only found one value, store it directly for simple filters (e.g., complexity=5).
	// Otherwise, store the entire slice for multi-value filters.
	if len(values) == 1 {
		node.Value = values[0]
	} else {
		node.Value = values
	}
	
	return node
}
```

### Step 4: Implement and Register the Execution Logic

Finally, we implement the handler in the executor. This function must be able to process both a single value (a `string`) and a list of values (a `[]interface{}`). For efficiency, we convert the list into a map for fast lookups.

**File: `executor.go`**

1.  Create the new `handleCategoryFilter` function.

    ```go
    func (se *SearchEngine) handleCategoryFilter(op string, val interface{}, docs []Document) []Document {
        var results []Document
        isEquality := (op == "=") // Also handles '!=' by inverting the result

        switch v := val.(type) {
        case string: // Handles single value: category=doc
            for _, doc := range docs {
                if (doc.Category == v) == isEquality {
                    results = append(results, doc)
                }
            }
        case []interface{}: // Handles multiple values: category=doc,pdf
            valueSet := make(map[string]bool)
            for _, item := range v {
                if strItem, ok := item.(string); ok {
                    valueSet[strItem] = true
                }
            }
            for _, doc := range docs {
                _, exists := valueSet[doc.Category]
                if exists == isEquality {
                    results = append(results, doc)
                }
            }
        }
        return results
    }
    ```

2.  Register the new handler in the `registerFilterHandlers` function.

    ```go
    func (se *SearchEngine) registerFilterHandlers() {
        se.filterHandlers = make(map[string]FilterHandler)
        se.filterHandlers["complexity"] = se.handleComplexityFilter
        se.filterHandlers["category"] = se.handleCategoryFilter // <-- Add this line
    }
    ```

With these changes, the engine is now fully equipped to handle multi-value `IN` and `NOT IN` queries on the `category` field, demonstrating the clean separation of concerns and extensibility of the architecture.