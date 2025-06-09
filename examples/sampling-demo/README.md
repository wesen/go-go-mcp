# Sampling Demo MCP Server

This demo showcases the Anthropic MCP sampling capabilities implemented in the go-go-mcp framework. The server demonstrates how MCP servers can request LLM completions from clients, enabling sophisticated agentic behaviors.

## Features

This demo implements three tools that demonstrate different aspects of sampling:

### 1. Text Analysis Tool (`analyze_text`)
- Analyzes text for sentiment, tone, readability, or other characteristics
- Demonstrates basic sampling with system prompts and temperature control
- Uses model hints to request specific AI models

### 2. Content Summarizer (`summarize_content`)
- Summarizes content with configurable length (short, medium, long)
- Shows dynamic prompt construction and token limit management
- Demonstrates different sampling parameters for different use cases

### 3. Conversation Tool (`conversation`)
- Enables multi-turn conversations using sampling
- Demonstrates context inclusion and conversation history management
- Shows how to build complex interactions with LLMs

## Sampling Features Demonstrated

- **Message Construction**: Text and image message support
- **Model Preferences**: Model hints and priority settings
- **Context Control**: Including context from servers
- **Sampling Parameters**: Temperature, max tokens, stop sequences
- **System Prompts**: Custom system prompts for different tasks
- **Error Handling**: Robust error handling for sampling failures

## Usage

### Running the Demo Server

```bash
# Build and run the demo
go run examples/sampling-demo/main.go
```

### Example Tool Calls

```bash
# Analyze text sentiment
go-go-mcp client tools call analyze_text --args '{
  "text": "I love this new feature!",
  "analysis_type": "sentiment"
}'

# Summarize content
go-go-mcp client tools call summarize_content --args '{
  "content": "Long article content here...",
  "length": "short"
}'

# Have a conversation
go-go-mcp client tools call conversation --args '{
  "messages": [
    {"role": "user", "text": "Hello, how are you?"},
    {"role": "assistant", "text": "I'm doing well, thank you!"},
    {"role": "user", "text": "What can you help me with?"}
  ]
}'
```

## Sampling Request Examples

### Basic Text Analysis
```json
{
  "messages": [
    {
      "role": "user",
      "content": {
        "type": "text",
        "text": "Please analyze the following text for sentiment: I love this product!"
      }
    }
  ],
  "systemPrompt": "You are a helpful text analysis assistant.",
  "maxTokens": 500,
  "temperature": 0.3,
  "modelPreferences": {
    "hints": [{"name": "claude-3"}]
  },
  "includeContext": "none"
}
```

### Multi-turn Conversation
```json
{
  "messages": [
    {
      "role": "user",
      "content": {"type": "text", "text": "Hello!"}
    },
    {
      "role": "assistant", 
      "content": {"type": "text", "text": "Hi there! How can I help?"}
    },
    {
      "role": "user",
      "content": {"type": "text", "text": "Tell me about MCP sampling"}
    }
  ],
  "systemPrompt": "You are a helpful assistant.",
  "maxTokens": 500,
  "temperature": 0.7,
  "includeContext": "thisServer"
}
```

## Architecture

The demo uses the following components:

- **SamplingServer**: Core sampling functionality for servers
- **SamplingRequestBuilder**: Fluent API for building sampling requests
- **Tool Implementations**: Example tools that use sampling
- **Transport Layer**: Handles communication with clients

## Security Considerations

This demo includes best practices for sampling:

- Input validation for all parameters
- Reasonable token limits to prevent excessive usage
- Temperature controls for consistent outputs
- Context isolation options
- Error handling and fallback behaviors

## Configuration

The demo can be configured via the `demo.yaml` file, which defines:

- Tool schemas and parameters
- Default sampling settings
- Security constraints
- Resource and prompt definitions

## Client Requirements

To use this demo, the MCP client must support:

- Sampling capability (`sampling/createMessage`)
- JSON-RPC 2.0 protocol
- Transport layer (stdio/SSE)

## Future Enhancements

Potential improvements for the demo:

- Image analysis tools
- Streaming response support
- Advanced context management
- Rate limiting and usage tracking
- Custom model preference handling
