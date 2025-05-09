package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/examples/libgen"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// RegisterLibGenTools registers tools for interacting with LibGen:
// libgen_search and libgen_download.
func RegisterLibGenTools(registry *tool_registry.Registry) error {
	if err := registerLibGenSearchTool(registry); err != nil {
		return errors.Wrap(err, "failed to register libgen_search tool")
	}
	if err := registerLibGenDownloadTool(registry); err != nil {
		return errors.Wrap(err, "failed to register libgen_download tool")
	}
	return nil
}

// registerLibGenSearchTool registers the tool to search for books on LibGen.
func registerLibGenSearchTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query to find books on LibGen."
			},
			"limit": {
				"type": "integer",
				"description": "Maximum number of results to return.",
				"default": 10
			},
			"mirror": {
				"type": "string",
				"description": "LibGen mirror to use. Defaults to https://libgen.rs if not provided."
			},
			"debug": {
				"type": "boolean",
				"description": "Enable debug output with URLs and raw data.",
				"default": false
			},
			"format": {
				"type": "string",
				"enum": ["text", "json"],
				"description": "Format of the output (text or json).",
				"default": "text"
			},
			"scientific": {
				"type": "boolean",
				"description": "Search for scientific papers instead of books.",
				"default": false
			}
		},
		"required": ["query"]
	}`

	tool, err := tools.NewToolImpl(
		"libgen_search",
		"Search for books on LibGen by query. Returns ID, title, author, year, and MD5 for each result.",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			query, ok := arguments["query"].(string)
			if !ok || query == "" {
				return protocol.NewToolResult(protocol.WithError("missing or empty required argument 'query'")), nil
			}

			limit := 10 // default
			if limitArg, ok := arguments["limit"].(float64); ok {
				limit = int(limitArg)
			}

			mirror := libgen.DefaultMirror
			if mirrorArg, ok := arguments["mirror"].(string); ok && mirrorArg != "" {
				mirror = mirrorArg
			}

			// Set debug mode if requested
			if debugArg, ok := arguments["debug"].(bool); ok && debugArg {
				libgen.Debug = true
			} else {
				libgen.Debug = false
			}

			// Output format (text or json)
			format := "text"
			if formatArg, ok := arguments["format"].(string); ok && (formatArg == "text" || formatArg == "json") {
				format = formatArg
			}

			scientific := false
			if scientificArg, ok := arguments["scientific"].(bool); ok && scientificArg {
				scientific = scientificArg
			}

			log.Debug().
				Str("query", query).
				Int("limit", limit).
				Str("mirror", mirror).
				Bool("debug", libgen.Debug).
				Str("format", format).
				Bool("scientific", scientific).
				Msg("Executing LibGen search")

			// Use our local libgen implementation
			result, err := libgen.SearchBooks(query, limit, mirror, scientific)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error searching LibGen: %v", err)),
				), nil
			}

			if len(result.Books) == 0 {
				return protocol.NewToolResult(
					protocol.WithText(fmt.Sprintf("No results found for query: %s", query)),
				), nil
			}

			// Return JSON if requested
			if format == "json" {
				jsonBytes, err := json.MarshalIndent(result.Books, "", "  ")
				if err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error converting results to JSON: %v", err)),
					), nil
				}
				return protocol.NewToolResult(
					protocol.WithText(string(jsonBytes)),
				), nil
			}

			// Default: return text format with logs
			return protocol.NewToolResult(
				protocol.WithText(result.Logs),
			), nil
		})

	return nil
}

// registerLibGenDownloadTool registers the tool to download books from LibGen.
func registerLibGenDownloadTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"description": "LibGen book ID to download."
			},
			"output_path": {
				"type": "string",
				"description": "Path where the downloaded book should be saved. If not provided, uses the book's MD5 hash with appropriate extension."
			},
			"mirror": {
				"type": "string",
				"description": "LibGen mirror to use. Defaults to https://libgen.rs if not provided."
			},
			"debug": {
				"type": "boolean",
				"description": "Enable debug output with URLs and raw data.",
				"default": false
			},
			"scientific": {
				"type": "boolean",
				"description": "Download a scientific paper instead of a book.",
				"default": false
			}
		},
		"required": ["id"]
	}`

	tool, err := tools.NewToolImpl(
		"libgen_download",
		"Download a book from LibGen by ID and save it to the specified path.",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			id, ok := arguments["id"].(string)
			if !ok || id == "" {
				return protocol.NewToolResult(protocol.WithError("missing or empty required argument 'id'")), nil
			}

			var outputPath string
			if outArg, ok := arguments["output_path"].(string); ok && outArg != "" {
				outputPath = outArg
			}

			mirror := libgen.DefaultMirror
			if mirrorArg, ok := arguments["mirror"].(string); ok && mirrorArg != "" {
				mirror = mirrorArg
			}

			// Set debug mode if requested
			if debugArg, ok := arguments["debug"].(bool); ok && debugArg {
				libgen.Debug = true
			} else {
				libgen.Debug = false
			}

			scientific := false
			if scientificArg, ok := arguments["scientific"].(bool); ok && scientificArg {
				scientific = scientificArg
			}

			log.Debug().
				Str("id", id).
				Str("outputPath", outputPath).
				Str("mirror", mirror).
				Bool("debug", libgen.Debug).
				Bool("scientific", scientific).
				Msg("Executing LibGen download")

			// Make sure output path is absolute if provided
			if outputPath != "" && !filepath.IsAbs(outputPath) {
				// Get current working directory and join with relative path
				absPath, err := filepath.Abs(outputPath)
				if err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error resolving absolute path: %v", err)),
					), nil
				}
				outputPath = absPath
			}

			// Use our local libgen implementation
			result, err := libgen.DownloadBook(id, outputPath, mirror, scientific)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error downloading book: %v", err)),
				), nil
			}

			// Format a successful response with logs
			response := fmt.Sprintf("Successfully downloaded book to: %s (%.2f MB)\n\n%s",
				result.FilePath, float64(result.Size)/(1024*1024), result.Logs)

			return protocol.NewToolResult(
				protocol.WithText(response),
			), nil
		})

	return nil
}
