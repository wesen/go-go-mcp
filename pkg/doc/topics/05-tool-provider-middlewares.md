---
Title: Tool Provider Middlewares in MCP
Slug: tool-provider-middlewares
Short: Learn how to use and create middleware for MCP tool providers to add logging, metrics, and other cross-cutting concerns
Topics:
  - middleware
  - tools
  - logging
  - metrics
Commands:
  - start
  - tools
Flags:
  - profile
  - transport
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This guide will walk you through using and creating middleware for MCP tool providers.

## Table of Contents

1. [Understanding Middleware](#understanding-middleware)
2. [Using Built-in Middleware](#using-built-in-middleware)
3. [Creating Custom Middleware](#creating-custom-middleware)
4. [Best Practices](#best-practices)
5. [Advanced Topics](#advanced-topics)

## Understanding Middleware

MCP uses a decorator pattern for middleware, where each middleware wraps a tool provider and can intercept, modify, or monitor tool calls. The middleware chain is built using a fluent builder pattern for easy configuration.

### Core Concepts

1. **Middleware Function Type**
   ```go
   type ToolProviderMiddleware func(pkg.ToolProvider) pkg.ToolProvider
   ```
   This is the core type that defines a middleware. It takes a tool provider and returns a wrapped version of it.

2. **Chain Builder**
   ```go
   builder := NewChainBuilder().
       With(middleware1).
       With(middleware2).
       Build(baseProvider)
   ```
   The chain builder helps construct middleware chains in a readable and maintainable way.

3. **Middleware Order**
   Middleware is applied in reverse order of addition. The last middleware added is the outermost wrapper. For example:
   ```go
   builder.With(logging).With(metrics)
   ```
   Results in: `metrics -> logging -> base provider`

## Using Built-in Middleware

MCP comes with several built-in middleware implementations that you can use immediately.

### Logging Middleware

The logging middleware provides structured logging of tool calls using slog. Here's how to use it:

```go
provider := NewChainBuilder().
    With(NewLoggingBuilder().
        WithLogLevel("debug").
        WithFormat("json").
        WithOutput("stdout").
        Build()).
    Build(baseProvider)
```

This will log:
- Tool call durations
- Arguments and results
- Error information
- Context values

### Metrics Middleware

The metrics middleware collects metrics about tool usage through a pluggable metrics collector interface:

```go
// Create a custom metrics collector
type PrometheusCollector struct {
    // Your Prometheus metrics here
}

func (p *PrometheusCollector) ObserveToolCall(operation string, tool string, duration time.Duration, err error) {
    // Record metrics using Prometheus
}

// Use it in your middleware chain
provider := NewChainBuilder().
    With(NewMetricsBuilder().
        WithCollector(myPrometheusCollector).
        Build()).
    Build(baseProvider)
```

The metrics middleware collects:
- Call counts
- Duration histograms
- Error counts
- Custom metrics through your collector

## Creating Custom Middleware

You can create your own middleware to add custom functionality. Here's a step-by-step guide:

1. **Create Your Provider Type**
   ```go
   type myCustomProvider struct {
       next pkg.ToolProvider
       // Add your custom fields here
   }
   ```

2. **Implement the ToolProvider Interface**
   ```go
   func (m *myCustomProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
       // Add your logic before the call
       tools, nextCursor, err := m.next.ListTools(ctx, cursor)
       // Add your logic after the call
       return tools, nextCursor, err
   }

   func (m *myCustomProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
       // Add your logic before the call
       result, err := m.next.CallTool(ctx, name, arguments)
       // Add your logic after the call
       return result, err
   }
   ```

3. **Create a Builder (Optional but Recommended)**
   ```go
   type MyCustomBuilder struct {
       // Builder configuration
       config MyConfig
   }

   func NewMyCustomBuilder() *MyCustomBuilder {
       return &MyCustomBuilder{
           config: DefaultConfig,
       }
   }

   func (b *MyCustomBuilder) WithOption(opt string) *MyCustomBuilder {
       b.config.Option = opt
       return b
   }

   func (b *MyCustomBuilder) Build() middlewares.ToolProviderMiddleware {
       return func(next pkg.ToolProvider) pkg.ToolProvider {
           return &myCustomProvider{
               next: next,
               // Initialize with builder config
           }
       }
   }
   ```

4. **Use Your Middleware**
   ```go
   provider := NewChainBuilder().
       With(NewMyCustomBuilder().
           WithOption("value").
           Build()).
       Build(baseProvider)
   ```

## Best Practices

1. **Use the Builder Pattern**
   - Makes middleware configuration clear and type-safe

2. **Handle Context Properly**
   - Always pass context through to the next provider
   - Use context for timeouts and cancellation
   - Add context values when appropriate

3. **Error Handling**
   - Don't swallow errors from the next provider
   - Add context to errors when wrapping them
   - Log errors at appropriate levels using zerolog

4. **Resource Management**
   - Clean up resources when context is cancelled
   - Use proper synchronization for shared state

## Advanced Topics

### Conditional Middleware

You can create middleware that only activates under certain conditions:

```go
func WithConditional(condition bool, middleware ToolProviderMiddleware) ToolProviderMiddleware {
    return func(next pkg.ToolProvider) pkg.ToolProvider {
        if condition {
            return middleware(next)
        }
        return next
    }
}

// Usage
builder.With(WithConditional(debug, NewLoggingBuilder().Build()))
```

### Middleware with Cleanup

For middleware that needs cleanup:

```go
type cleanupProvider struct {
    next pkg.ToolProvider
    cleanup func()
}

func (c *cleanupProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
    // Use defer to ensure cleanup runs
    defer c.cleanup()
    return c.next.CallTool(ctx, name, arguments)
}
```

### Context-Aware Middleware

Middleware can use context to share data:

```go
type contextKey string

func WithRequestID(next pkg.ToolProvider) pkg.ToolProvider {
    return pkg.ToolProviderFunc(func(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
        requestID := uuid.New().String()
        ctx = context.WithValue(ctx, contextKey("request-id"), requestID)
        return next.CallTool(ctx, name, arguments)
    })
}
```

## Examples

Here are some real-world examples of using middleware:

### Development Setup
```go
provider := NewChainBuilder().
    With(NewLoggingBuilder().
        WithLogLevel("debug").
        Build()).
    With(NewMetricsBuilder().
        WithCollector(NewNoopMetricsCollector()).
        Build()).
    Build(baseProvider)
```

### Production Setup
```go
provider := NewChainBuilder().
    With(NewLoggingBuilder().
        WithLogLevel("info").
        WithFormat("json").
        Build()).
    With(NewMetricsBuilder().
        WithCollector(promCollector).
        Build()).
    Build(baseProvider)
```

### Testing Setup
```go
provider := NewChainBuilder().
    With(NewLoggingBuilder().
        WithLogLevel("debug").
        Build()).
    With(NewTestingMiddleware().
        WithAssertions(assertions).
        Build()).
    Build(baseProvider)
```
