# You.com Web Search Integration

This branch adds You.com web search capability to the codespelunker MCP server, complementing the existing code search functionality.

## New Features

### MCP Web Search Tool

A new `web_search` tool has been added to the MCP server that provides web search capabilities using the You.com Search API.

#### Usage

The tool is automatically available when running `cs` in MCP mode:

```bash
cs --mcp --dir /path/to/codebase
```

#### Tool Parameters

- `query` (required): The web search query. Can be keywords, phrases, or questions.
- `max_results` (optional): Maximum number of results to return (1-20). Defaults to 10.

#### Example Usage in Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "codespelunker": {
      "command": "/path/to/cs",
      "args": ["--mcp", "--dir", "/path/to/codebase"],
      "env": {
        "YOU_COM_API_KEY": "optional_api_key"
      }
    }
  }
}
```

### Authentication Modes

#### Keyless Operation (Default)
- Uses the public You.com Search API endpoint
- 100 free searches per day per IP address
- No API key required

#### Authenticated Operation
- Set the `YOU_COM_API_KEY` environment variable
- Higher rate limits and enhanced features
- Uses the authenticated You.com Search API

### Integration Benefits

1. **Complementary Functionality**: Web search complements code search - use `search` for local code and `web_search` for external information
2. **Zero Setup Required**: Works out of the box with keyless operation
3. **Optional Enhancement**: Adding an API key provides higher limits
4. **Consistent Interface**: Follows the same MCP patterns as existing tools
5. **Error Handling**: Graceful fallback and informative error messages

### Example Queries

```bash
# In Claude with MCP enabled
"Use web_search to find Go error handling best practices"
"Search the web for JWT authentication tutorials"
"Find recent articles about GraphQL performance optimization"
```

### Response Format

```json
{
  "query": "golang error handling",
  "results_returned": 5,
  "results": [
    {
      "title": "Error Handling in Go",
      "url": "https://example.com/go-errors",
      "snippet": "Go has a built-in error type...",
      "source": "example.com",
      "type": "web"
    }
  ]
}
```

## Implementation Details

### Files Modified

- `mcp.go`: Added web search tool registration
- `youcom.go`: New file containing You.com API client and MCP handler

### Key Features

- Automatic API endpoint selection based on authentication
- Proper error handling and user feedback
- Standard User-Agent for integration tracking
- Rate limiting awareness
- Structured JSON responses

### Security Considerations

- API keys are optional and read from environment variables
- No sensitive data is logged or exposed
- Requests include proper User-Agent identification

## Testing

To test the integration:

1. Build the updated cs binary
2. Set up MCP server mode: `cs --mcp --dir .`
3. Configure Claude Desktop or another MCP client
4. Try web search queries through the MCP interface

## Compatibility

- Maintains full backward compatibility with existing functionality
- No changes to existing `search` or `get_file` tools
- Web search is purely additive functionality