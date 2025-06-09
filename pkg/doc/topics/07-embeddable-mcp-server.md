---
Title: Embeddable MCP Server
Slug: embeddable-mcp-server
Short: Add MCP server capabilities to your Go applications with minimal code changes.
Topics:
  - embeddable
  - server
  - integration
  - api
Commands:
  - mcp
Flags:
  - transport
  - port
  - config
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The embeddable MCP server package provides a simple way for Go applications to add MCP (Model Context Protocol) server capabilities with minimal code changes. Instead of building a standalone MCP server, you can embed MCP functionality directly into your existing applications.

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Tool Registration Methods](#tool-registration-methods)
4. [Enhanced API (v2)](#enhanced-api-v2)
5. [Session Management](#session-management)
6. [Configuration Options](#configuration-options)
7. [Command Structure](#command-structure)
8. [Advanced Features](#advanced-features)
9. [Real-World Examples](#real-world-examples)
10. [Best Practices](#best-practices)
11. [Migration Guide](#migration-guide)

## Introduction

The embeddable package transforms any Cobra-based Go application into an MCP server by adding a standard `mcp` subcommand. This approach allows you to:

- **Leverage Existing Applications**: Add MCP capabilities to existing CLI tools
- **Minimal Integration**: Add server functionality with just a few lines of code
- **Share Business Logic**: Expose your application's functionality to AI models
- **Maintain Consistency**: Use the same codebase for both CLI and MCP interfaces

### When to Use Embeddable

Consider using the embeddable package when:

- You have an existing Cobra-based CLI application
- You want to expose your application's functionality to AI models
- You need to share tools across multiple environments
- You want to avoid maintaining separate MCP server implementations

## Quick Start

Here's how to add MCP server capabilities to your existing application:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My application",
    }

    // Add MCP server capability
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("MyApp MCP Server"),
        embeddable.WithVersion("1.0.0"),
        embeddable.WithServerDescription("Example MCP server"),
        embeddable.WithTool("greet", greetHandler,
            embeddable.WithDescription("Greet a person"),
            embeddable.WithStringArg("name", "Name of the person to greet", true),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func greetHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    name, ok := args["name"].(string)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("name must be a string")), nil
    }

    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Hello, %s!", name)),
    ), nil
}
```

### Usage

Once you've added the embeddable package, your application gains these new commands:

```bash
# Start the MCP server with stdio transport
myapp mcp start

# Start with SSE transport on port 3001
myapp mcp start --transport sse --port 3001

# List available tools
myapp mcp list-tools

# Test a specific tool
myapp mcp test-tool greet --args '{"name":"World"}'
```

## Tool Registration Methods

The embeddable package supports multiple ways to register tools, allowing you to choose the approach that best fits your needs.

### Simple Function-Based Registration

Register tools using simple handler functions:

```go
func calculatorHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    a, ok1 := args["a"].(float64)
    b, ok2 := args["b"].(float64)
    operation, ok3 := args["operation"].(string)
    
    if !ok1 || !ok2 || !ok3 {
        return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid arguments")), nil
    }
    
    var result float64
    switch operation {
    case "add":
        result = a + b
    case "subtract":
        result = a - b
    case "multiply":
        result = a * b
    case "divide":
        if b == 0 {
            return protocol.NewErrorToolResult(protocol.NewTextContent("Cannot divide by zero")), nil
        }
        result = a / b
    default:
        return protocol.NewErrorToolResult(protocol.NewTextContent("Unknown operation")), nil
    }
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Result: %f", result)),
    ), nil
}

// Register the tool
embeddable.WithTool("calculator", calculatorHandler,
    embeddable.WithDescription("Perform basic arithmetic operations"),
    embeddable.WithNumberArg("a", "First number", true),
    embeddable.WithNumberArg("b", "Second number", true),
    embeddable.WithStringArg("operation", "Operation to perform", true),
)
```

### Struct-Based Registration

Register tools from struct methods for better organization:

```go
type DatabaseService struct {
    connectionString string
}

type QueryArgs struct {
    Query string `json:"query" description:"SQL query to execute"`
    Limit int    `json:"limit,omitempty" description:"Maximum number of rows to return"`
}

func (db *DatabaseService) ExecuteQuery(ctx context.Context, args QueryArgs) (*protocol.ToolResult, error) {
    // Implement database query logic
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Executed: %s", args.Query)),
    ), nil
}

func main() {
    config := embeddable.NewServerConfig()
    dbService := &DatabaseService{connectionString: "..."}
    
    err := embeddable.RegisterStructTool(config, "execute_query", dbService, "ExecuteQuery")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use config with AddMCPCommand
}
```

### Reflection-Based Registration

Automatically register functions using reflection:

```go
func addNumbers(ctx context.Context, args AddArgs) (*protocol.ToolResult, error) {
    result := args.A + args.B
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Sum: %d", result)),
    ), nil
}

type AddArgs struct {
    A int `json:"a" description:"First number"`
    B int `json:"b" description:"Second number"`
}

// Register using reflection
err := embeddable.RegisterFunctionTool(config, "add", addNumbers)
if err != nil {
    log.Fatal(err)
}
```

## Enhanced API (v2)

Inspired by [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go), the enhanced API provides more powerful argument handling and tool configuration.

### Enhanced Tool Registration

```go
embeddable.WithEnhancedTool("format_text", formatTextHandler,
    embeddable.WithEnhancedDescription("Format text with various options"),
    embeddable.WithReadOnlyHint(true),
    embeddable.WithIdempotentHint(true),
    embeddable.WithStringProperty("text",
        embeddable.PropertyDescription("Text to format"),
        embeddable.PropertyRequired(),
        embeddable.MinLength(1),
    ),
    embeddable.WithStringProperty("format",
        embeddable.PropertyDescription("Format type"),
        embeddable.StringEnum("uppercase", "lowercase", "title"),
        embeddable.DefaultString("lowercase"),
    ),
    embeddable.WithIntProperty("max_length",
        embeddable.PropertyDescription("Maximum length"),
        embeddable.Minimum(1),
        embeddable.Maximum(1000),
        embeddable.DefaultNumber(100),
    ),
)
```

### Enhanced Argument Access

The enhanced API provides type-safe argument access with automatic validation:

```go
func formatTextHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    // Type-safe argument access with validation
    text, err := args.RequireString("text")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }
    
    // Get arguments with defaults
    format := args.GetString("format", "lowercase")
    maxLength := args.GetInt("max_length", 100)
    enabled := args.GetBool("enabled", true)
    
    // Bind to struct for complex argument sets
    var config FormatConfig
    if err := args.BindArguments(&config); err != nil {
        return nil, err
    }
    
    // Process the text
    result := processText(text, format, maxLength)
    
    return protocol.NewToolResult(
        protocol.WithText(result),
    ), nil
}

type FormatConfig struct {
    Text      string `json:"text"`
    Format    string `json:"format"`
    MaxLength int    `json:"max_length"`
    Enabled   bool   `json:"enabled"`
}
```

### Tool Annotations

Add semantic hints about tool behavior to help AI models understand how to use your tools:

```go
embeddable.WithReadOnlyHint(true),        // Tool doesn't modify environment
embeddable.WithDestructiveHint(false),    // Tool won't cause destructive changes
embeddable.WithIdempotentHint(true),      // Repeated calls have no additional effect
embeddable.WithOpenWorldHint(false),      // Tool doesn't interact with external entities
```

These annotations help AI models make better decisions about when and how to call your tools.

## Session Management

The embeddable API provides automatic session management through Go's context system, allowing tools to maintain state across multiple invocations.

### Accessing Session Data

```go
func counterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Session is automatically available via context
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }

    // Get current counter value
    counterVal, exists := sess.GetData("counter")
    counter := 0
    if exists {
        if c, ok := counterVal.(int); ok {
            counter = c
        }
    }

    // Increment and store
    counter++
    sess.SetData("counter", counter)
    
    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Counter: %d", counter)),
    ), nil
}
```

### Stateful Tool Example

Here's a more complex example showing how to build a stateful calculator:

```go
type Calculator struct {
    Value  float64            `json:"value"`
    Memory map[string]float64 `json:"memory"`
}

func calculatorHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    sess, ok := session.GetSessionFromContext(ctx)
    if !ok {
        return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
    }

    // Get or create calculator state
    var calc Calculator
    if calcData, exists := sess.GetData("calculator"); exists {
        if c, ok := calcData.(Calculator); ok {
            calc = c
        }
    } else {
        calc = Calculator{
            Value:  0,
            Memory: make(map[string]float64),
        }
    }

    // Get operation and operand
    operation, err := args.RequireString("operation")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }

    var result float64
    switch operation {
    case "clear":
        calc.Value = 0
        result = calc.Value
    case "add":
        value := args.GetNumber("value", 0)
        calc.Value += value
        result = calc.Value
    case "store":
        name := args.GetString("name", "default")
        calc.Memory[name] = calc.Value
        result = calc.Value
    case "recall":
        name := args.GetString("name", "default")
        if val, ok := calc.Memory[name]; ok {
            calc.Value = val
            result = calc.Value
        } else {
            return protocol.NewErrorToolResult(
                protocol.NewTextContent(fmt.Sprintf("No value stored in %s", name)),
            ), nil
        }
    }

    // Store updated calculator state
    sess.SetData("calculator", calc)

    return protocol.NewToolResult(
        protocol.WithText(fmt.Sprintf("Calculator value: %f", result)),
        protocol.WithJSON(calc),
    ), nil
}
```

## Configuration Options

The embeddable package provides extensive configuration options to customize your MCP server.

### Server Configuration

```go
err := embeddable.AddMCPCommand(rootCmd,
    // Basic server info
    embeddable.WithName("Advanced Server"),
    embeddable.WithVersion("2.0.0"),
    embeddable.WithServerDescription("Full-featured MCP server"),
    
    // Transport configuration
    embeddable.WithDefaultTransport("sse"),
    embeddable.WithDefaultPort(3001),
    
    // Tool registration
    embeddable.WithTool("tool1", handler1, opts...),
    embeddable.WithEnhancedTool("tool2", handler2, enhancedOpts...),
    
    // Advanced features
    embeddable.WithToolRegistry(customRegistry),
    embeddable.WithSessionStore(customStore),
    embeddable.WithMiddleware(loggingMiddleware, authMiddleware),
    embeddable.WithHooks(lifecycleHooks),
)
```

### Middleware Support

Add middleware to process tool calls:

```go
func loggingMiddleware(next embeddable.ToolHandler) embeddable.ToolHandler {
    return func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
        log.Printf("Calling tool with args: %v", args)
        start := time.Now()
        
        result, err := next(ctx, args)
        
        log.Printf("Tool completed in %v, error: %v", time.Since(start), err)
        return result, err
    }
}

func authMiddleware(next embeddable.ToolHandler) embeddable.ToolHandler {
    return func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
        // Check authentication
        if !isAuthenticated(ctx) {
            return protocol.NewErrorToolResult(
                protocol.NewTextContent("Authentication required"),
            ), nil
        }
        return next(ctx, args)
    }
}

// Use middleware
embeddable.AddMCPCommand(rootCmd,
    embeddable.WithMiddleware(loggingMiddleware, authMiddleware),
    // ... other options
)
```

### Custom Tool Registry

For advanced scenarios, you can provide a custom tool registry:

```go
registry := tool_registry.NewRegistry()

// Register tools manually
err := registerCustomTools(registry)
if err != nil {
    log.Fatal(err)
}

err = embeddable.AddMCPCommand(rootCmd,
    embeddable.WithToolRegistry(registry),
    // ... other options
)
```

## Command Structure

The embeddable package adds a comprehensive `mcp` subcommand to your application:

```
myapp mcp
├── start              # Start the MCP server
│   ├── --transport    # Transport type (stdio, sse)
│   ├── --port         # Port for SSE transport
│   └── --config       # Configuration file path
├── list-tools         # List available tools
├── test-tool          # Test a specific tool
│   ├── --args         # Tool arguments as JSON
│   └── --interactive  # Interactive mode
└── config             # Configuration management (future)
```

### Starting the Server

```bash
# Start with stdio transport (default)
myapp mcp start

# Start with SSE transport
myapp mcp start --transport sse --port 3001

# Start with configuration file
myapp mcp start --config config.yaml
```

### Testing Tools

```bash
# List all available tools
myapp mcp list-tools

# Test a tool with JSON arguments
myapp mcp test-tool greet --args '{"name":"World"}'

# Test a tool interactively
myapp mcp test-tool calculator --interactive
```

## Advanced Features

### Error Handling

The embeddable API provides consistent error handling patterns:

```go
func myHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
    // Validation errors (shown to user)
    if someCondition {
        return protocol.NewErrorToolResult(
            protocol.NewTextContent("Invalid input: expected positive number"),
        ), nil
    }
    
    // System errors (logged, generic error shown to user)
    if err := doSomething(); err != nil {
        return nil, fmt.Errorf("system error: %w", err)
    }
    
    // Success
    return protocol.NewToolResult(
        protocol.WithText("Operation completed successfully"),
    ), nil
}
```

### Custom Session Store

For distributed or persistent sessions, you can provide a custom session store:

```go
type RedisSessionStore struct {
    client *redis.Client
}

func (r *RedisSessionStore) GetSession(id string) (*session.Session, error) {
    // Implement Redis-backed session retrieval
}

func (r *RedisSessionStore) StoreSession(sess *session.Session) error {
    // Implement Redis-backed session storage
}

// Use custom session store
embeddable.AddMCPCommand(rootCmd,
    embeddable.WithSessionStore(&RedisSessionStore{client: redisClient}),
    // ... other options
)
```

## Real-World Examples

### Example 1: File Management Server

Transform a file management CLI into an MCP server:

```go
package main

import (
    "context"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/go-go-golems/go-go-mcp/pkg/embeddable"
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "filemanager",
        Short: "File management utility",
    }

    // Add existing CLI commands
    rootCmd.AddCommand(createListCommand())
    rootCmd.AddCommand(createCopyCommand())

    // Add MCP server capability
    err := embeddable.AddMCPCommand(rootCmd,
        embeddable.WithName("File Manager MCP Server"),
        embeddable.WithVersion("1.0.0"),
        embeddable.WithEnhancedTool("list_files", listFilesHandler,
            embeddable.WithEnhancedDescription("List files in a directory"),
            embeddable.WithReadOnlyHint(true),
            embeddable.WithStringProperty("path",
                embeddable.PropertyDescription("Directory path"),
                embeddable.PropertyRequired(),
            ),
            embeddable.WithStringProperty("pattern",
                embeddable.PropertyDescription("File pattern to match"),
                embeddable.DefaultString("*"),
            ),
        ),
        embeddable.WithEnhancedTool("read_file", readFileHandler,
            embeddable.WithEnhancedDescription("Read file contents"),
            embeddable.WithReadOnlyHint(true),
            embeddable.WithStringProperty("path",
                embeddable.PropertyDescription("File path"),
                embeddable.PropertyRequired(),
            ),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func listFilesHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    path, err := args.RequireString("path")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }

    pattern := args.GetString("pattern", "*")

    var files []map[string]interface{}
    err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if matched, _ := filepath.Match(pattern, d.Name()); matched {
            info, _ := d.Info()
            files = append(files, map[string]interface{}{
                "name":    d.Name(),
                "path":    filePath,
                "size":    info.Size(),
                "isDir":   d.IsDir(),
                "modTime": info.ModTime(),
            })
        }
        return nil
    })

    if err != nil {
        return protocol.NewErrorToolResult(
            protocol.NewTextContent(fmt.Sprintf("Error listing files: %v", err)),
        ), nil
    }

    return protocol.NewToolResult(
        protocol.WithJSON(files),
    ), nil
}

func readFileHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    path, err := args.RequireString("path")
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }

    content, err := os.ReadFile(path)
    if err != nil {
        return protocol.NewErrorToolResult(
            protocol.NewTextContent(fmt.Sprintf("Error reading file: %v", err)),
        ), nil
    }

    return protocol.NewToolResult(
        protocol.WithText(string(content)),
    ), nil
}
```

### Example 2: API Gateway Server

Transform an API client into an MCP server:

```go
type APIClient struct {
    baseURL string
    apiKey  string
}

func (c *APIClient) CallAPI(ctx context.Context, args APIArgs) (*protocol.ToolResult, error) {
    // Implement API call logic
    url := fmt.Sprintf("%s/%s", c.baseURL, args.Endpoint)
    
    // Make HTTP request
    resp, err := http.Get(url)
    if err != nil {
        return protocol.NewErrorToolResult(
            protocol.NewTextContent(fmt.Sprintf("API call failed: %v", err)),
        ), nil
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return protocol.NewErrorToolResult(
            protocol.NewTextContent(fmt.Sprintf("Failed to read response: %v", err)),
        ), nil
    }

    return protocol.NewToolResult(
        protocol.WithText(string(body)),
    ), nil
}

type APIArgs struct {
    Endpoint string            `json:"endpoint" description:"API endpoint to call"`
    Method   string            `json:"method" description:"HTTP method"`
    Headers  map[string]string `json:"headers,omitempty" description:"HTTP headers"`
    Body     string            `json:"body,omitempty" description:"Request body"`
}

func main() {
    client := &APIClient{
        baseURL: "https://api.example.com",
        apiKey:  os.Getenv("API_KEY"),
    }

    config := embeddable.NewServerConfig()
    err := embeddable.RegisterStructTool(config, "api_call", client, "CallAPI")
    if err != nil {
        log.Fatal(err)
    }

    rootCmd := &cobra.Command{Use: "apiclient"}
    err = embeddable.AddMCPCommandWithConfig(rootCmd, config)
    if err != nil {
        log.Fatal(err)
    }

    rootCmd.Execute()
}
```

## Best Practices

### 1. Tool Naming and Documentation

Use clear, descriptive names and comprehensive documentation:

```go
embeddable.WithEnhancedTool("calculate_distance", calculateDistanceHandler,
    embeddable.WithEnhancedDescription("Calculate distance between two geographic points using the Haversine formula"),
    embeddable.WithReadOnlyHint(true),
    embeddable.WithIdempotentHint(true),
)
```

### 2. Input Validation

Always validate inputs before processing:

```go
func validateCoordinates(lat, lon float64) error {
    if lat < -90 || lat > 90 {
        return fmt.Errorf("latitude must be between -90 and 90")
    }
    if lon < -180 || lon > 180 {
        return fmt.Errorf("longitude must be between -180 and 180")
    }
    return nil
}
```

### 3. Error Handling

Provide meaningful error messages:

```go
if err := validateInput(input); err != nil {
    return protocol.NewErrorToolResult(
        protocol.NewTextContent(fmt.Sprintf("Invalid input: %v", err)),
    ), nil
}
```

### 4. Resource Management

Clean up resources properly:

```go
func fileProcessorHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
    file, err := os.Open(path)
    if err != nil {
        return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
    }
    defer file.Close() // Always clean up

    // Process file...
}
```

### 5. Session State Management

Use sessions for stateful operations:

```go
func initializeState(sess *session.Session, key string, defaultValue interface{}) {
    if _, exists := sess.GetData(key); !exists {
        sess.SetData(key, defaultValue)
    }
}
```

## Migration Guide

### From Manual MCP Server

If you have an existing manual MCP server implementation, you can migrate to the embeddable API:

1. **Extract Tool Logic**: Move your tool handler functions to separate functions
2. **Update Handler Signatures**: Adapt to the embeddable handler signature
3. **Replace Registration**: Use embeddable registration methods
4. **Update Configuration**: Use embeddable configuration options

**Before (Manual)**:
```go
// Manual server setup with custom registration
server := server.NewServer()
registry := tool_registry.NewRegistry()
registry.RegisterToolWithHandler(tool, handler)
server.SetToolProvider(registry)
```

**After (Embeddable)**:
```go
// Embeddable setup
embeddable.AddMCPCommand(rootCmd,
    embeddable.WithTool("tool_name", handler, opts...),
)
```

### From Standalone Application

To convert a standalone MCP server to use embeddable:

1. **Create Cobra Root Command**: Wrap your application in a Cobra command
2. **Move Tool Registration**: Use embeddable registration methods
3. **Update Main Function**: Use embeddable.AddMCPCommand instead of manual server setup
4. **Preserve Configuration**: Migrate configuration to embeddable options

### Gradual Migration

The embeddable API can coexist with existing implementations:

```go
// Combine embeddable tools with existing registry
existingRegistry := createExistingRegistry()

err := embeddable.AddMCPCommand(rootCmd,
    embeddable.WithToolRegistry(existingRegistry),
    embeddable.WithTool("new_tool", newHandler), // Add new tools
)
```

For more detailed information about specific MCP concepts, see:
- [Building Go MCP Tools](06-building-a-go-mcp-tool.md)
- [MCP in Practice](03-mcp-in-practice.md)
- [Configuration Files](01-config-file.md)
