## Summary

This PR adds You.com web search capability to the codespelunker MCP server, complementing the existing code search functionality with web search capabilities.

## What Changed

- **New web_search MCP tool**: Searches the web using You.com Search API
- **Dual operation modes**: 
  - Keyless operation (100 free searches/day per IP, no setup required)
  - Authenticated operation (with YOU_COM_API_KEY env var for higher limits)
- **Clean integration**: Adds new functionality without changing existing behavior
- **Proper error handling**: Graceful fallbacks and informative error messages
- **Standard tracking**: Uses youdotcom-integration/boyter-cs User-Agent

## Why This is Useful for cs

The integration addresses a natural use case gap: while cs excels at searching local codebases, developers often need to search for external documentation, tutorials, API references, and best practices that aren't in their local code. The new web_search tool provides this capability directly within the MCP interface.

**Use cases:**
- Find Go error handling best practices
- Search for JWT authentication tutorials
- Look up GraphQL performance optimization articles
- Find documentation for APIs being integrated

## Integration Details

### MCP Tool Schema
web_search(query: string, max_results?: number)

### Response Format
Returns JSON with query, results_returned count, and array of results containing title, url, snippet, and type fields.

### Setup Example for Claude Desktop
Add YOU_COM_API_KEY environment variable for authenticated access or use keyless mode (100 searches/day).

## Validation Performed

✅ MCP interface compatibility: Follows existing cs MCP patterns  
✅ Error handling: Graceful failures with informative messages  
✅ Backward compatibility: Zero impact on existing search and get_file tools  
✅ Authentication handling: Supports both keyless and authenticated modes  
✅ Rate limiting awareness: Respects API constraints  

## Files Modified

- mcp.go: Added web search tool registration  
- youcom.go: New You.com API client and MCP handler  
- YOU_DOT_COM_INTEGRATION.md: Integration documentation  

The implementation follows cs existing patterns and maintains the project focus on being a powerful, lightweight code search tool while adding complementary web search when needed.