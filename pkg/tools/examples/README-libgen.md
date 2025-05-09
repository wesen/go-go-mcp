# LibGen Tools for go-go-mcp

This package provides MCP tools for interacting with the Library Genesis (LibGen) service, which allows searching and downloading books.

## Available Tools

### 1. libgen_search

Search for books on LibGen by query.

**Parameters:**
- `query` (string, required): Search query to find books on LibGen.
- `limit` (integer, optional, default: 10): Maximum number of results to return.
- `mirror` (string, optional, default: "https://libgen.rs"): LibGen mirror to use.
- `debug` (boolean, optional, default: false): Show debug output with URLs and raw data.
- `format` (string, optional, default: "text"): Output format - either "text" or "json".

**Example:**
```json
{
  "name": "libgen_search",
  "arguments": {
    "query": "system design interview",
    "limit": 5,
    "format": "json"
  }
}
```

### 2. libgen_download

Download a book from LibGen by ID.

**Parameters:**
- `id` (string, required): LibGen book ID to download.
- `output_path` (string, optional): Path where the downloaded book should be saved. If not provided, uses the book's MD5 hash with appropriate extension.
- `mirror` (string, optional, default: "https://libgen.rs"): LibGen mirror to use.
- `debug` (boolean, optional, default: false): Show debug output with URLs and raw data.

**Example:**
```json
{
  "name": "libgen_download",
  "arguments": {
    "id": "3043135",
    "output_path": "system-design-interview.pdf"
  }
}
```

## Usage

To enable these tools when starting the MCP server, use the `--internal-servers` flag:

```bash
go-go-mcp server start --internal-servers=libgen
```

Or combine with other tools:

```bash
go-go-mcp server start --internal-servers=libgen,sqlite,fetch
```

## Implementation Details

The tools use a local implementation of the LibGen functionality in the `pkg/tools/examples/libgen` package, adapted from the original CLI in go-go-labs. This approach provides better integration with the MCP framework without requiring external command execution.

Key features:
- Direct HTTP requests to LibGen mirrors
- Parsing and extraction of book information
- Direct download functionality with progress reporting
- Clean Go API for both searching and downloading
- All logs and debug output are captured and returned to the caller, rather than being printed to stdout
- Support for JSON output format for easier integration with other tools

## Example Workflow

1. Search for books in text format (default):
   ```json
   {"name": "libgen_search", "arguments": {"query": "Go programming language"}}
   ```

2. Or search for books in JSON format:
   ```json
   {"name": "libgen_search", "arguments": {"query": "Go programming language", "format": "json"}}
   ```

3. Find the ID of the book you want to download from the search results.

4. Download the book using its ID:
   ```json
   {"name": "libgen_download", "arguments": {"id": "12345678"}}
   ```

## Go Code Example

If you want to use the LibGen library directly in your Go code instead of through the MCP tools, you can import the package and use it like this:

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples/libgen"
)

func main() {
	// Optional: enable debug output
	libgen.Debug = true

	// Search for books
	result, err := libgen.SearchBooks("go programming", 5, "")
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Found %d books\n", len(result.Books))
	fmt.Println(result.Logs) // All output is captured in result.Logs
	
	// Alternatively, convert to JSON
	booksJSON, _ := json.MarshalIndent(result.Books, "", "  ")
	fmt.Println(string(booksJSON))

	// Download a book (if you found one)
	if len(result.Books) > 0 {
		book := result.Books[0]
		fmt.Printf("Downloading book: %s\n", book.Title)
		
		dlResult, err := libgen.DownloadBook(book.ID, "downloaded-book.pdf", "")
		if err != nil {
			fmt.Printf("Error downloading: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Successfully downloaded to %s (%.2f MB)\n", 
			dlResult.FilePath, float64(dlResult.Size)/(1024*1024))
		fmt.Println(dlResult.Logs) // All output is captured in dlResult.Logs
	}
} 