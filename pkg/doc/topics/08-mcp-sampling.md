---
Title: MCP Sampling Support
Slug: mcp-sampling
Short: Learn how to use Anthropic MCP sampling to enable LLM interactions in your MCP servers.
Topics:
  - sampling
  - llm
  - anthropic
  - ai
  - protocols
Commands:
  - server
  - client
Flags:
  - profile
  - transport
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This guide covers the Anthropic Model Context Protocol (MCP) sampling functionality implemented in go-go-mcp. Sampling allows MCP servers to request LLM completions from clients, enabling sophisticated agentic behaviors while maintaining security and user control.

## Table of Contents

1. [Introduction to MCP Sampling](#introduction-to-mcp-sampling)
2. [Sampling Architecture](#sampling-architecture)
3. [Setting Up Sampling](#setting-up-sampling)
4. [Creating Sampling Requests](#creating-sampling-requests)
5. [Implementing Sampling Providers](#implementing-sampling-providers)
6. [Real-World Examples](#real-world-examples)
7. [Security Considerations](#security-considerations)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Introduction to MCP Sampling

MCP Sampling is a powerful feature that allows MCP servers to request LLM completions from clients. This enables servers to:

- Analyze and process data using AI models
- Generate summaries and insights
- Provide intelligent responses to complex queries
- Create agentic workflows that combine tool execution with AI reasoning

### Key Features

The go-go-mcp sampling implementation provides:

- **Full MCP Specification Compliance**: Complete support for the Anthropic MCP sampling specification
- **Rich Message Support**: Text and image message handling
- **Model Preferences**: Request specific models with hints and priorities
- **Context Control**: Include context from servers or keep it isolated
- **Parameter Control**: Temperature, max tokens, stop sequences, and metadata
- **Human-in-the-Loop**: Clients maintain control over sampling requests

### When to Use Sampling

Consider using sampling when your MCP server needs to:

- Analyze user input or data with AI models
- Generate natural language responses
- Summarize large amounts of information
- Make AI-powered decisions in workflows
- Create conversational experiences

## Sampling Architecture

The sampling system follows this architecture:

```
┌─────────────┐    sampling/createMessage    ┌─────────────┐
│             │ ────────────────────────────► │             │
│ MCP Server  │                               │ MCP Client  │
│             │ ◄──────────────────────────── │             │
└─────────────┘    CreateMessageResponse      └─────────────┘
       │                                             │
       ▼                                             ▼
┌─────────────┐                               ┌─────────────┐
│ Sampling    │                               │    LLM      │
│ Provider    │                               │  Provider   │
└─────────────┘                               └─────────────┘
```

**Components:**

- **MCP Server**: Your server that needs AI capabilities
- **Sampling Provider**: Interface for making sampling requests
- **MCP Client**: Client that handles sampling requests
- **LLM Provider**: The actual AI model (Claude, GPT, etc.)

### Request Flow

1. Server creates a sampling request using `SamplingRequestBuilder`
2. Request is sent to client via `sampling/createMessage`
3. Client reviews and may modify the request
4. Client forwards to LLM provider
5. Client reviews the completion
6. Response is returned to server

This design ensures users maintain control over what data is sent to AI models and what responses are returned.

## Setting Up Sampling

### Installing Dependencies

The sampling functionality is built into go-go-mcp. Import the necessary packages:

```go
import (
    "github.com/go-go-golems/go-go-mcp/pkg/protocol"
    "github.com/go-go-golems/go-go-mcp/pkg/server"
)
```

### Basic Server Setup

Create a server with sampling support:

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-go-golems/go-go-mcp/pkg/server"
    "github.com/go-go-golems/go-go-mcp/pkg/transport/stdio"
    "github.com/rs/zerolog"
)

func main() {
    logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
    
    // Create transport
    transport, err := stdio.NewStdioTransport(
        transport.WithLogger(logger),
    )
    if err != nil {
        log.Fatal("Failed to create transport:", err)
    }
    
    // Create sampling provider (implementation depends on your setup)
    samplingProvider := &YourSamplingProvider{}
    
    // Create server with sampling
    server := NewServerWithSampling(logger, transport, samplingProvider)
    
    // Start server
    ctx := context.Background()
    if err := server.Start(ctx); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### Implementing the Sampling Provider Interface

The `SamplingProvider` interface enables servers to request sampling:

```go
type SamplingProvider interface {
    RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error)
}
```

For development and testing, you can use a mock provider:

```go
type MockSamplingProvider struct{}

func (m *MockSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    return &protocol.CreateMessageResponse{
        Role: "assistant",
        Content: protocol.MessageContent{
            Type: "text",
            Text: "This is a mock response for development.",
        },
        Model:      "mock-model",
        StopReason: "endTurn",
    }, nil
}
```

## Creating Sampling Requests

### Using the SamplingRequestBuilder

The `SamplingRequestBuilder` provides a fluent API for constructing sampling requests:

```go
// Create a basic text analysis request
request := server.NewSamplingRequest().
    AddTextMessage("user", "Please analyze the sentiment of this text: I love this product!").
    WithSystemPrompt("You are a helpful sentiment analysis assistant.").
    WithMaxTokens(100).
    WithTemperature(0.3).
    WithModelHint("claude-3").
    Build()

// Make the sampling request
response, err := samplingServer.RequestSampling(ctx, request)
if err != nil {
    return fmt.Errorf("sampling failed: %w", err)
}

fmt.Println("Analysis:", response.Content.Text)
```

### Message Types

**Text Messages:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Hello, how are you?").
    Build()
```

**Image Messages:**
```go
request := server.NewSamplingRequest().
    AddImageMessage("user", "base64encodeddata", "image/png").
    AddTextMessage("user", "What do you see in this image?").
    Build()
```

**Multi-turn Conversations:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Hello").
    AddTextMessage("assistant", "Hi there! How can I help?").
    AddTextMessage("user", "I need help with my code").
    WithSystemPrompt("You are a helpful programming assistant").
    Build()
```

### Model Preferences

Control which AI model is used:

```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Explain quantum computing").
    WithModelHint("claude-3-sonnet").  // Prefer Claude 3 Sonnet
    WithModelHint("gpt-4").            // Fallback to GPT-4
    Build()
```

Set model selection priorities:

```go
prefs := &protocol.ModelPreferences{
    Hints: []protocol.ModelHint{
        {Name: "claude-3-sonnet"},
    },
    CostPriority:         0.3,  // Moderate cost concern
    SpeedPriority:        0.8,  // High speed priority
    IntelligencePriority: 0.9,  // High intelligence priority
}

request := server.NewSamplingRequest().
    AddTextMessage("user", "Complex analysis task").
    WithModelPreferences(prefs).
    Build()
```

### Sampling Parameters

**Temperature Control:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Write a creative story").
    WithTemperature(0.9).  // High creativity
    Build()

request := server.NewSamplingRequest().
    AddTextMessage("user", "Analyze this data").
    WithTemperature(0.1).  // Low variability
    Build()
```

**Token Limits:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Summarize this article").
    WithMaxTokens(200).  // Concise summary
    Build()
```

**Stop Sequences:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Generate a list").
    WithStopSequences([]string{"---", "[END]"}).
    Build()
```

**Context Inclusion:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "What tools are available?").
    WithIncludeContext("thisServer").  // Include this server's context
    Build()

request := server.NewSamplingRequest().
    AddTextMessage("user", "Analyze this private data").
    WithIncludeContext("none").  // No additional context
    Build()
```

**Custom Metadata:**
```go
metadata := map[string]interface{}{
    "task_type": "analysis",
    "priority":  "high",
    "session_id": "abc123",
}

request := server.NewSamplingRequest().
    AddTextMessage("user", "Important analysis task").
    WithMetadata(metadata).
    Build()
```

## Implementing Sampling Providers

### Production Implementation

In production, the sampling provider typically connects to your MCP client:

```go
type ClientSamplingProvider struct {
    client *mcp.Client
}

func (c *ClientSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    // Forward the request to the MCP client
    response, err := c.client.CreateMessage(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("client sampling failed: %w", err)
    }
    
    return response, nil
}
```

### HTTP Sampling Provider

For HTTP-based sampling services:

```go
type HTTPSamplingProvider struct {
    baseURL string
    client  *http.Client
    apiKey  string
}

func (h *HTTPSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    // Convert request to HTTP API format
    payload, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/sampling/createMessage", bytes.NewBuffer(payload))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+h.apiKey)
    
    // Send request
    resp, err := h.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Parse response
    var response protocol.CreateMessageResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &response, nil
}
```

## Real-World Examples

### Text Analysis Tool

```go
type TextAnalysisTool struct {
    sampling *server.SamplingServer
}

func (t *TextAnalysisTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    text, ok := args["text"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'text' parameter")
    }
    
    analysisType, ok := args["analysis_type"].(string)
    if !ok {
        analysisType = "sentiment"
    }
    
    // Build sampling request
    request := server.NewSamplingRequest().
        AddTextMessage("user", fmt.Sprintf("Analyze this text for %s: %s", analysisType, text)).
        WithSystemPrompt("You are an expert text analyst. Provide clear, actionable insights.").
        WithMaxTokens(300).
        WithTemperature(0.2).
        WithModelHint("claude-3").
        WithIncludeContext("none").
        Build()
    
    // Request sampling
    response, err := t.sampling.RequestSampling(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }
    
    return map[string]interface{}{
        "analysis":    response.Content.Text,
        "model_used":  response.Model,
        "stop_reason": response.StopReason,
        "analysis_type": analysisType,
    }, nil
}
```

### Smart Summarizer

```go
type SmartSummarizerTool struct {
    sampling *server.SamplingServer
}

func (t *SmartSummarizerTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    content, ok := args["content"].(string)
    if !ok {
        return nil, fmt.Errorf("missing content")
    }
    
    length, _ := args["length"].(string)
    if length == "" {
        length = "medium"
    }
    
    // Dynamic prompt based on content length
    var prompt string
    var maxTokens int
    
    contentWords := len(strings.Fields(content))
    
    switch {
    case contentWords < 100:
        prompt = "Provide a brief summary highlighting the key point:"
        maxTokens = 50
    case contentWords < 1000:
        prompt = "Summarize the main points and conclusions:"
        maxTokens = 200
    default:
        prompt = "Provide a comprehensive summary covering all major themes and insights:"
        maxTokens = 500
    }
    
    // Adjust for requested length
    switch length {
    case "short":
        maxTokens = min(maxTokens, 100)
        prompt = "Provide a concise summary:"
    case "long":
        maxTokens = maxTokens * 2
        prompt = "Provide a detailed summary with analysis:"
    }
    
    request := server.NewSamplingRequest().
        AddTextMessage("user", fmt.Sprintf("%s\n\n%s", prompt, content)).
        WithSystemPrompt("You are an expert summarizer. Focus on the most important information.").
        WithMaxTokens(maxTokens).
        WithTemperature(0.3).
        WithModelHint("claude-3-sonnet").
        Build()
    
    response, err := t.sampling.RequestSampling(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("summarization failed: %w", err)
    }
    
    return map[string]interface{}{
        "summary":       response.Content.Text,
        "original_words": contentWords,
        "summary_words":  len(strings.Fields(response.Content.Text)),
        "compression_ratio": float64(len(strings.Fields(response.Content.Text))) / float64(contentWords),
        "model_used":    response.Model,
    }, nil
}
```

### Conversational Assistant

```go
type ConversationalTool struct {
    sampling *server.SamplingServer
    sessions map[string][]protocol.Message
}

func (t *ConversationalTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    sessionID, _ := args["session_id"].(string)
    if sessionID == "" {
        sessionID = "default"
    }
    
    userMessage, ok := args["message"].(string)
    if !ok {
        return nil, fmt.Errorf("missing message")
    }
    
    // Initialize session if needed
    if t.sessions == nil {
        t.sessions = make(map[string][]protocol.Message)
    }
    
    // Add user message to conversation
    t.sessions[sessionID] = append(t.sessions[sessionID], protocol.Message{
        Role: "user",
        Content: protocol.MessageContent{
            Type: "text",
            Text: userMessage,
        },
    })
    
    // Build request with conversation history
    builder := server.NewSamplingRequest().
        WithSystemPrompt("You are a helpful, conversational assistant. Maintain context and provide thoughtful responses.").
        WithMaxTokens(400).
        WithTemperature(0.7).
        WithModelHint("claude-3").
        WithIncludeContext("thisServer")
    
    // Add conversation history
    for _, msg := range t.sessions[sessionID] {
        builder.AddMessage(msg.Role, msg.Content)
    }
    
    request := builder.Build()
    
    response, err := t.sampling.RequestSampling(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("conversation failed: %w", err)
    }
    
    // Store assistant response
    t.sessions[sessionID] = append(t.sessions[sessionID], protocol.Message{
        Role: "assistant",
        Content: response.Content,
        Model: response.Model,
    })
    
    return map[string]interface{}{
        "response":    response.Content.Text,
        "session_id":  sessionID,
        "turn_count":  len(t.sessions[sessionID]),
        "model_used":  response.Model,
    }, nil
}
```

### Image Analysis Tool

```go
type ImageAnalysisTool struct {
    sampling *server.SamplingServer
}

func (t *ImageAnalysisTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    imageData, ok := args["image_data"].(string)
    if !ok {
        return nil, fmt.Errorf("missing image_data")
    }
    
    mimeType, ok := args["mime_type"].(string)
    if !ok {
        mimeType = "image/jpeg"
    }
    
    question, ok := args["question"].(string)
    if !ok {
        question = "Describe what you see in this image."
    }
    
    request := server.NewSamplingRequest().
        AddImageMessage("user", imageData, mimeType).
        AddTextMessage("user", question).
        WithSystemPrompt("You are an expert image analyst. Provide detailed, accurate descriptions.").
        WithMaxTokens(500).
        WithTemperature(0.4).
        WithModelHint("claude-3").  // Ensure vision-capable model
        Build()
    
    response, err := t.sampling.RequestSampling(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("image analysis failed: %w", err)
    }
    
    return map[string]interface{}{
        "analysis":   response.Content.Text,
        "question":   question,
        "mime_type":  mimeType,
        "model_used": response.Model,
    }, nil
}
```

## Security Considerations

### Data Privacy

**Context Isolation:**
```go
// Don't include sensitive server context
request := server.NewSamplingRequest().
    AddTextMessage("user", "Analyze this sensitive data").
    WithIncludeContext("none").  // Isolate from other context
    Build()
```

**Input Sanitization:**
```go
func sanitizeInput(text string) string {
    // Remove potential sensitive information
    re := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
    text = re.ReplaceAllString(text, "[EMAIL]")
    
    // Remove credit card numbers
    re = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
    text = re.ReplaceAllString(text, "[CARD]")
    
    return text
}
```

### Rate Limiting

```go
type RateLimitedSamplingProvider struct {
    provider server.SamplingProvider
    limiter  *rate.Limiter
}

func (r *RateLimitedSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    if !r.limiter.Allow() {
        return nil, fmt.Errorf("rate limit exceeded")
    }
    
    return r.provider.RequestSampling(ctx, request)
}
```

### Cost Control

```go
type CostControlledSamplingProvider struct {
    provider   server.SamplingProvider
    maxTokens  int
    dailyLimit int
    usage      map[string]int
}

func (c *CostControlledSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    // Enforce token limits
    if request.MaxTokens > c.maxTokens {
        request.MaxTokens = c.maxTokens
    }
    
    // Check daily usage
    today := time.Now().Format("2006-01-02")
    if c.usage[today] >= c.dailyLimit {
        return nil, fmt.Errorf("daily usage limit exceeded")
    }
    
    response, err := c.provider.RequestSampling(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // Track usage (approximate)
    c.usage[today] += len(strings.Fields(response.Content.Text))
    
    return response, nil
}
```

## Best Practices

### Request Construction

**1. Use Clear, Specific Prompts:**
```go
// Good: Specific and actionable
request := server.NewSamplingRequest().
    AddTextMessage("user", "Extract the main action items from this meeting transcript and format them as a numbered list").
    WithSystemPrompt("You are a professional meeting assistant focused on identifying actionable tasks.")

// Avoid: Vague and ambiguous
request := server.NewSamplingRequest().
    AddTextMessage("user", "Do something with this text")
```

**2. Set Appropriate Token Limits:**
```go
// For summaries
WithMaxTokens(200)

// For detailed analysis
WithMaxTokens(800)

// For simple answers
WithMaxTokens(50)
```

**3. Use Temperature Wisely:**
```go
// Creative tasks
WithTemperature(0.8)

// Analytical tasks
WithTemperature(0.2)

// Balanced responses
WithTemperature(0.5)
```

### Error Handling

**Robust Error Handling:**
```go
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    request := server.NewSamplingRequest().
        AddTextMessage("user", "Analyze this").
        WithMaxTokens(200).
        Build()
    
    response, err := t.sampling.RequestSampling(ctx, request)
    if err != nil {
        // Log the error
        log.Printf("Sampling failed: %v", err)
        
        // Return a fallback response
        return map[string]interface{}{
            "error": "Analysis unavailable",
            "fallback": "Unable to process request at this time",
        }, nil
    }
    
    // Validate response
    if response.Content.Text == "" {
        return map[string]interface{}{
            "error": "Empty response",
            "fallback": "No analysis generated",
        }, nil
    }
    
    return map[string]interface{}{
        "analysis": response.Content.Text,
        "model": response.Model,
    }, nil
}
```

### Performance Optimization

**Timeout Handling:**
```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

response, err := samplingProvider.RequestSampling(ctx, request)
```

**Caching Responses:**
```go
type CachedSamplingProvider struct {
    provider server.SamplingProvider
    cache    map[string]*protocol.CreateMessageResponse
    mutex    sync.RWMutex
}

func (c *CachedSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    // Create cache key from request
    key := c.createCacheKey(request)
    
    // Check cache
    c.mutex.RLock()
    if cached, exists := c.cache[key]; exists {
        c.mutex.RUnlock()
        return cached, nil
    }
    c.mutex.RUnlock()
    
    // Not cached, make request
    response, err := c.provider.RequestSampling(ctx, request)
    if err != nil {
        return nil, err
    }
    
    // Cache the response
    c.mutex.Lock()
    c.cache[key] = response
    c.mutex.Unlock()
    
    return response, nil
}
```

## Troubleshooting

### Common Issues

**1. Sampling Provider Not Available:**
```
Error: no sampling provider available
```

**Solution:** Ensure you've properly injected a sampling provider:
```go
samplingProvider := &YourSamplingProvider{}
samplingServer := server.NewSamplingServer(samplingProvider)
```

**2. Request Too Large:**
```
Error: request exceeds maximum size
```

**Solutions:**
- Reduce message content length
- Lower max tokens
- Split large requests into smaller chunks

**3. Model Not Available:**
```
Error: requested model not supported
```

**Solutions:**
- Use model hints instead of specific model names
- Set fallback model preferences
- Check client model availability

**4. Context Issues:**
```
Error: context inclusion failed
```

**Solutions:**
- Use `WithIncludeContext("none")` for isolated requests
- Verify server context is available
- Check client context handling

### Debugging

**Enable Debug Logging:**
```go
request := server.NewSamplingRequest().
    AddTextMessage("user", "Debug this request").
    WithMetadata(map[string]interface{}{
        "debug": true,
        "trace_id": "12345",
    }).
    Build()
```

**Log Request Details:**
```go
func (p *DebugSamplingProvider) RequestSampling(ctx context.Context, request protocol.CreateMessageRequest) (*protocol.CreateMessageResponse, error) {
    log.Printf("Sampling request: %+v", request)
    
    response, err := p.provider.RequestSampling(ctx, request)
    
    if err != nil {
        log.Printf("Sampling error: %v", err)
    } else {
        log.Printf("Sampling response: %+v", response)
    }
    
    return response, err
}
```

### Testing

**Unit Testing with Mock Provider:**
```go
func TestTextAnalysisTool(t *testing.T) {
    mockProvider := &MockSamplingProvider{
        response: &protocol.CreateMessageResponse{
            Role: "assistant",
            Content: protocol.MessageContent{
                Type: "text",
                Text: "Positive sentiment detected",
            },
            Model: "test-model",
        },
    }
    
    tool := &TextAnalysisTool{
        sampling: server.NewSamplingServer(mockProvider),
    }
    
    result, err := tool.Execute(context.Background(), map[string]interface{}{
        "text": "I love this product!",
        "analysis_type": "sentiment",
    })
    
    assert.NoError(t, err)
    assert.Contains(t, result.(map[string]interface{})["analysis"], "Positive")
}
```

**Integration Testing:**
```go
func TestSamplingIntegration(t *testing.T) {
    // Set up real client and server
    client := setupTestClient()
    provider := &ClientSamplingProvider{client: client}
    
    request := server.NewSamplingRequest().
        AddTextMessage("user", "Hello").
        WithMaxTokens(50).
        Build()
    
    response, err := provider.RequestSampling(context.Background(), request)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content.Text)
}
```

For more details on specific topics, see:
- [Building MCP Tools](06-building-a-go-mcp-tool.md)
- [MCP in Practice](03-mcp-in-practice.md)
- [Configuration Files](01-config-file.md)

## Next Steps

1. **Implement Your First Sampling Tool**: Start with a simple text analysis tool
2. **Add Error Handling**: Implement robust error handling and fallbacks
3. **Optimize Performance**: Add caching and timeout handling
4. **Enhance Security**: Implement rate limiting and input sanitization
5. **Create Advanced Tools**: Build multi-turn conversations and image analysis

Remember to:
- Test thoroughly with different sampling providers
- Monitor usage and costs in production
- Keep security and privacy in mind
- Document your sampling tools clearly
- Follow MCP best practices for tool development

The sampling functionality opens up powerful possibilities for creating intelligent, AI-powered MCP tools that can understand, analyze, and respond to complex user needs.
