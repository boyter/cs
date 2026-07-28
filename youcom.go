// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// YouComSearchResult represents a search result from You.com Search API
type YouComSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source,omitempty"`
	PublishDate string `json:"publish_date,omitempty"`
}

// YouComSearchResponse represents the response from You.com Search API
type YouComSearchResponse struct {
	Results []YouComSearchResult `json:"web"`
	News    []YouComSearchResult `json:"news,omitempty"`
}

// YouComClient handles requests to You.com Search API
type YouComClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
}

// NewYouComClient creates a new You.com API client
func NewYouComClient() *YouComClient {
	apiKey := os.Getenv("YOU_COM_API_KEY")
	
	// Use keyless endpoint if no API key provided
	baseURL := "https://api.you.com/v1/agents/search"
	if apiKey != "" {
		baseURL = "https://ydc-index.io/v1/search"
	}

	return &YouComClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "youdotcom-integration/boyter-cs",
	}
}

// Search performs a web search using You.com Search API
func (c *YouComClient) Search(ctx context.Context, query string, maxResults int) (*YouComSearchResponse, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 10 // default
	}
	if maxResults > 20 {
		maxResults = 20 // API limit
	}

	var reqBody []byte
	var err error
	
	if c.APIKey != "" {
		// Authenticated API (ydc-index.io)
		payload := map[string]interface{}{
			"query": query,
			"count": maxResults,
		}
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	} else {
		// Keyless API (api.you.com/v1/agents/search)
		payload := map[string]interface{}{
			"query": query,
			"count": maxResults,
		}
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if c.APIKey != "" {
		if strings.Contains(c.BaseURL, "ydc-index.io") {
			req.Header.Set("X-API-Key", c.APIKey)
		} else {
			req.Header.Set("Authorization", "Bearer "+c.APIKey)
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var searchResp YouComSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// mcpWebSearchResult represents a web search result for MCP responses
type mcpWebSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source,omitempty"`
	PublishDate string `json:"publish_date,omitempty"`
	Type        string `json:"type"` // "web" or "news"
}

// mcpWebSearchResponse wraps web search results for MCP
type mcpWebSearchResponse struct {
	Query           string               `json:"query"`
	ResultsReturned int                  `json:"results_returned"`
	Results         []mcpWebSearchResult `json:"results"`
	Message         string               `json:"message,omitempty"`
}

// mcpWebSearchHandler returns an MCP tool handler for You.com web search
func mcpWebSearchHandler(youcom *YouComClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract query parameter
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("missing required parameter: query"), nil
		}
		
		if strings.TrimSpace(query) == "" {
			return mcp.NewToolResultError("query cannot be empty"), nil
		}

		// Extract max_results parameter (optional)
		maxResults := 10 // default
		if v, ok := request.GetArguments()["max_results"]; ok {
			switch typed := v.(type) {
			case float64:
				maxResults = int(typed)
			case int:
				maxResults = typed
			}
		}

		// Validate max_results
		if maxResults < 1 {
			maxResults = 1
		}
		if maxResults > 20 {
			maxResults = 20
		}

		// Perform the search
		searchResp, err := youcom.Search(ctx, query, maxResults)
		if err != nil {
			response := &mcpWebSearchResponse{
				Query:   query,
				Message: fmt.Sprintf("Search failed: %v", err),
				Results: []mcpWebSearchResult{},
			}
			jsonResult, marshalErr := mcp.NewToolResultJSON(response)
			if marshalErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
			}
			return jsonResult, nil
		}

		// Convert results
		var results []mcpWebSearchResult
		
		// Add web results
		for _, result := range searchResp.Results {
			results = append(results, mcpWebSearchResult{
				Title:       result.Title,
				URL:         result.URL,
				Snippet:     result.Snippet,
				Source:      result.Source,
				PublishDate: result.PublishDate,
				Type:        "web",
			})
		}
		
		// Add news results if available
		for _, result := range searchResp.News {
			results = append(results, mcpWebSearchResult{
				Title:       result.Title,
				URL:         result.URL,
				Snippet:     result.Snippet,
				Source:      result.Source,
				PublishDate: result.PublishDate,
				Type:        "news",
			})
		}

		response := &mcpWebSearchResponse{
			Query:           query,
			ResultsReturned: len(results),
			Results:         results,
		}

		jsonResult, err := mcp.NewToolResultJSON(response)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}
		
		return jsonResult, nil
	}
}