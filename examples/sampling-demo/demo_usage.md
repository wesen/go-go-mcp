# Sampling Demo Usage Guide

This document demonstrates how to use the MCP sampling demo server with properly registered tools.

## Quick Start

### 1. Build and Run the Server

```bash
# Build the demo
go build -o sampling-demo .

# Run the server (stdio transport)
./sampling-demo
```

The server will start and listen for MCP protocol messages on stdin/stdout.

### 2. Available Tools

The demo server registers three tools that demonstrate MCP sampling capabilities:

#### `analyze_text`
- **Purpose**: Analyze text for sentiment, tone, readability, etc.
- **Parameters**:
  - `text` (required): The text to analyze
  - `analysis_type` (optional): Type of analysis ("sentiment", "tone", "readability", etc.)

#### `summarize_content`
- **Purpose**: Summarize content with configurable length
- **Parameters**:
  - `content` (required): The content to summarize
  - `length` (optional): Summary length ("short", "medium", "long")

#### `conversation`
- **Purpose**: Engage in multi-turn conversations
- **Parameters**:
  - `messages` (required): Array of conversation messages with `role` and `text` fields

## MCP Protocol Usage

### List Available Tools

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "analyze_text",
        "description": "Analyze text using LLM sampling for sentiment, tone, or other analysis",
        "inputSchema": {
          "type": "object",
          "properties": {
            "text": {
              "type": "string",
              "description": "The text to analyze"
            },
            "analysis_type": {
              "type": "string",
              "description": "Type of analysis: sentiment, tone, readability, etc.",
              "default": "sentiment"
            }
          },
          "required": ["text"]
        }
      },
      // ... other tools
    ]
  }
}
```

### Call the Text Analysis Tool

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "analyze_text",
    "arguments": {
      "text": "I absolutely love this new feature! It works perfectly and saves me so much time.",
      "analysis_type": "sentiment"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"analysis\":\"This is a mock response from the sampling provider. In a real implementation, this would come from the MCP client's LLM.\",\"model_used\":\"mock-model\",\"stop_reason\":\"endTurn\"}"
      }
    ],
    "isError": false
  }
}
```

### Call the Summarizer Tool

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "summarize_content",
    "arguments": {
      "content": "Artificial Intelligence (AI) has revolutionized numerous industries by enabling machines to perform tasks that typically require human intelligence. Machine learning, a subset of AI, allows systems to automatically learn and improve from experience without being explicitly programmed. Natural language processing enables computers to understand, interpret, and generate human language. Computer vision allows machines to interpret and make decisions based on visual data. These technologies have applications in healthcare, finance, transportation, and many other sectors. However, the rapid advancement of AI also raises important ethical considerations regarding privacy, job displacement, and the need for responsible AI development.",
      "length": "short"
    }
  }
}
```

### Call the Conversation Tool

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "conversation",
    "arguments": {
      "messages": [
        {
          "role": "user",
          "text": "Hello! I'm working on a Python project and need some help."
        },
        {
          "role": "assistant",
          "text": "Hi there! I'd be happy to help with your Python project. What specific issue are you facing?"
        },
        {
          "role": "user",
          "text": "I'm trying to understand how to use decorators effectively."
        }
      ]
    }
  }
}
```

## Testing with the MCP Inspector

You can use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
# Install the MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Run the inspector with your server
mcp-inspector ./sampling-demo
```

This will open a web interface where you can:
- List all available tools
- View tool schemas
- Call tools with custom parameters
- See real-time responses

## Integration with Real MCP Clients

### Claude Desktop

To use with Claude Desktop, add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "sampling-demo": {
      "command": "/path/to/sampling-demo",
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### Custom Client

For custom clients, implement the `SamplingProvider` interface to connect to real LLM services:

```go
type RealSamplingProvider struct {
    apiKey string
    baseURL string
}

func (r *RealSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    // Implement actual LLM API calls here
    // This would connect to OpenAI, Anthropic, or other LLM providers
}
```

## Tool Development

### Adding New Tools

1. **Create the tool struct**:
```go
type MyCustomTool struct {
    *tools.ToolImpl
    sampling *server.SamplingServer
}
```

2. **Implement the constructor**:
```go
func NewMyCustomTool(sampling *server.SamplingServer) (*MyCustomTool, error) {
    schema := map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type": "string",
                "description": "Input parameter",
            },
        },
        "required": []string{"input"},
    }
    
    impl, err := tools.NewToolImpl(
        "my_custom_tool",
        "Description of my custom tool",
        schema,
    )
    if err != nil {
        return nil, err
    }
    
    return &MyCustomTool{
        ToolImpl: impl,
        sampling: sampling,
    }, nil
}
```

3. **Implement the Call method**:
```go
func (t *MyCustomTool) Call(ctx context.Context, arguments map[string]interface{}) (*protocol.ToolResult, error) {
    // Your tool logic here
    result, err := t.Execute(ctx, arguments)
    // Handle errors and format response
    return &protocol.ToolResult{
        Content: []protocol.ToolContent{{
            Type: "text",
            Text: string(resultJSON),
        }},
        IsError: false,
    }, nil
}
```

4. **Register the tool**:
```go
func (s *SamplingDemoServer) registerTools(registry *tool_registry.Registry) {
    // ... existing tools
    
    customTool, err := NewMyCustomTool(s.sampling)
    if err != nil {
        log.Printf("Failed to create custom tool: %v", err)
        return
    }
    registry.RegisterTool(customTool)
}
```

## Best Practices

1. **Error Handling**: Always handle errors gracefully and return meaningful error messages
2. **Schema Validation**: Define clear JSON schemas for tool parameters
3. **Documentation**: Provide clear descriptions for tools and parameters
4. **Testing**: Test tools with various input combinations
5. **Security**: Validate and sanitize all inputs
6. **Performance**: Implement reasonable timeouts and limits

## Next Steps

- Connect to a real LLM provider instead of using mock responses
- Add more sophisticated tools for your specific use case
- Implement authentication and authorization
- Add logging and monitoring
- Deploy to production with proper error handling

For more detailed information, see the [MCP Sampling Documentation](../../pkg/doc/topics/08-mcp-sampling.md).
