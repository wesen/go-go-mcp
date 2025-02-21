# MCP Tool Provider Middleware Implementation Plan

## Overview
Implement a flexible middleware pattern for Tool Providers that allows intercepting, modifying, and monitoring tool calls. This system will support composable middleware chains, enabling features like logging, metrics collection, rate limiting, and caching.

## Directory Structure
```
pkg/
  middleware/
    tool/
      types.go           # Core middleware types and interfaces
      chain.go           # Middleware chaining utilities
      providers/
        logging.go       # Logging middleware implementation
        metrics.go       # Metrics collection middleware
        ratelimit.go     # Rate limiting middleware
        cache.go         # Caching middleware
      factory.go         # Middleware factory and configuration
```

## Core Concepts and Implementation

### Middleware Interface
The middleware pattern will be based on the decorator pattern, where each middleware wraps the base ToolProvider:

```go
// types.go
type ToolProviderMiddleware func(ToolProvider) ToolProvider

type MiddlewareChain struct {
    middlewares []ToolProviderMiddleware
}

// Configuration for different middleware types
type MiddlewareConfig struct {
    Logging   *LoggingConfig
    Metrics   *MetricsConfig
    RateLimit *RateLimitConfig
    Cache     *CacheConfig
}
```

### Middleware Chain Builder
```go
// chain.go
type ChainBuilder struct {
    chain []ToolProviderMiddleware
}

func NewChainBuilder() *ChainBuilder
func (b *ChainBuilder) With(middleware ToolProviderMiddleware) *ChainBuilder
func (b *ChainBuilder) Build(base ToolProvider) ToolProvider
```

### Example Middleware Implementations

#### Logging Middleware
```go
// logging.go
type LoggingConfig struct {
    LogLevel string
    Format   string
    Output   string
}

func WithLogging(cfg LoggingConfig) ToolProviderMiddleware {
    return func(next ToolProvider) ToolProvider {
        return &loggingToolProvider{
            next: next,
            cfg:  cfg,
        }
    }
}
```

#### Metrics Middleware
```go
// metrics.go
type MetricsConfig struct {
    Namespace   string
    Subsystem   string
    Labels      []string
    Collectors  []string
}

func WithMetrics(cfg MetricsConfig) ToolProviderMiddleware {
    return func(next ToolProvider) ToolProvider {
        return &metricsToolProvider{
            next: next,
            cfg:  cfg,
        }
    }
}
```

## Implementation Tasks

### Core Middleware Framework
- [x] Define core middleware types and interfaces
- [x] Implement middleware chain builder
- [x] Add middleware configuration structures
- [x] Create middleware factory with configuration support

### Logging Middleware
- [x] Implement structured logging middleware
- [x] Add context propagation
- [x] Support different log formats (JSON, text)
- [x] Add log level filtering
- [x] Implement async logging option

### Metrics Middleware
- [x] Add metrics collection interface
- [x] Track call durations
- [x] Count success/failure rates
- [x] Implement custom metric labels
- [x] Add NoopMetricsCollector implementation

### Rate Limiting Middleware
- [ ] Implement token bucket rate limiter
- [ ] Add concurrent call limiting
- [ ] Support per-tool rate limits
- [ ] Implement backoff strategies
- [ ] Add quota management

### Caching Middleware
- [ ] Implement result caching
- [ ] Add TTL support
- [ ] Support different cache backends
- [ ] Implement cache invalidation
- [ ] Add cache warming strategies

### Integration Support
- [x] Create middleware registration system
- [x] Add middleware ordering control
- [x] Implement middleware context passing
- [x] Add middleware configuration validation
- [x] Create middleware documentation

## Usage Examples

### Basic Usage
```go
provider := NewChainBuilder().
    With(NewLoggingBuilder().
        WithLogLevel("info").
        WithFormat("json").
        Build()).
    With(NewMetricsBuilder().
        WithCollector(myCollector).
        Build()).
    Build(baseProvider)
```

### Configuration-based Setup
```go
config := MiddlewareConfig{
    Logging: &LoggingConfig{
        LogLevel: "debug",
        Format:   "json",
    },
    Metrics: &MetricsConfig{
        Namespace: "mcp",
        Labels:    []string{"tool", "version"},
    },
}

provider := NewProviderFromConfig(baseProvider, config)
```

## Notes
- Ensure thread safety
- Document middleware ordering requirements
- Consider resource cleanup

## Chain Builder Deep Dive

### Overview
The Chain Builder is a crucial component that implements the Builder pattern to construct middleware chains in a fluent, readable manner. It handles the complexity of middleware ordering and composition while providing a simple API for users.

### Core Components

#### ChainBuilder Structure
```go
type ChainBuilder struct {
    chain []ToolProviderMiddleware
    // Optional fields for advanced configuration
    errorHandler func(error) error
    recovery     func(interface{}) error
}
```

#### Builder Methods
```go
// Creates a new chain builder instance
func NewChainBuilder() *ChainBuilder {
    return &ChainBuilder{
        chain: make([]ToolProviderMiddleware, 0),
    }
}

// Adds a middleware to the chain
func (b *ChainBuilder) With(middleware ToolProviderMiddleware) *ChainBuilder {
    b.chain = append(b.chain, middleware)
    return b
}

// Optional: Adds error handling middleware
func (b *ChainBuilder) WithErrorHandler(handler func(error) error) *ChainBuilder {
    b.errorHandler = handler
    return b
}

// Constructs the final provider with all middleware applied
func (b *ChainBuilder) Build(base ToolProvider) ToolProvider {
    provider := base
    // Apply middlewares in reverse order
    // This ensures the first middleware in the chain is the outermost wrapper
    for i := len(b.chain) - 1; i >= 0; i-- {
        provider = b.chain[i](provider)
    }
    return provider
}
```

### Middleware Order and Composition

The chain builder handles middleware composition in a specific way:

1. **Order of Application**
   ```go
   // Middlewares are applied in reverse order
   builder.With(loggingMiddleware)     // Applied last  (outermost)
         .With(metricsMiddleware)      // Applied second
         .With(rateLimitMiddleware)    // Applied first (innermost)
   ```

   Visual representation of the final structure:
   ```
   Logging → Metrics → RateLimit → Base Provider
   ```

2. **Request Flow**
   ```
   Request  → Logging → Metrics → RateLimit → Base Provider
   Response ← Logging ← Metrics ← RateLimit ← Base Provider
   ```

### Usage Examples

#### Basic Usage
```go
provider := NewChainBuilder().
    With(WithLogging(LoggingConfig{
        LogLevel: "info",
    })).
    With(WithMetrics(MetricsConfig{
        Namespace: "mcp",
    })).
    Build(baseProvider)
```

#### Advanced Configuration
```go
builder := NewChainBuilder().
    WithErrorHandler(func(err error) error {
        // Custom error handling
        return fmt.Errorf("middleware error: %w", err)
    }).
    With(WithLogging(LoggingConfig{
        LogLevel: "debug",
        Format:   "json",
    })).
    With(WithRetry(RetryConfig{
        MaxAttempts: 3,
        Backoff:     exponentialBackoff,
    }))

// Can be built later
provider := builder.Build(baseProvider)
```

#### Conditional Middleware
```go
builder := NewChainBuilder()

if config.Debug {
    builder.With(WithDebugLogging())
}

if config.Metrics {
    builder.With(WithMetrics(metricsConfig))
}

provider := builder.Build(baseProvider)
```

### Best Practices

1. **Middleware Ordering**
   - Put cross-cutting concerns (logging, metrics) as outer layers
   - Put transformative middleware (rate limiting, caching) as inner layers
   - Consider dependencies between middleware when ordering

2. **Error Handling**
   - Each middleware should properly propagate errors
   - Use error wrapping to maintain context
   - Consider adding recovery middleware for panics

3. **Context Propagation**
   - Ensure context.Context is properly passed through the chain
   - Add relevant values to context when needed
   - Respect context cancellation

4. **Performance Considerations**
   - Minimize allocations in the middleware chain
   - Consider adding bypass mechanisms for certain cases
   - Use middleware only when necessary

### Testing

```go
func TestChainBuilder(t *testing.T) {
    mock := &MockToolProvider{}
    
    provider := NewChainBuilder().
        With(func(next ToolProvider) ToolProvider {
            return &testMiddleware{next: next}
        }).
        Build(mock)
        
    // Test the complete chain
    result, err := provider.CallTool(context.Background(), "test", nil)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Common Pitfalls

1. **Middleware Order Confusion**
   - Remember that middleware is applied in reverse order
   - Document the expected order of operations
   - Use clear naming to indicate middleware purpose

2. **Resource Management**
   - Ensure proper cleanup in each middleware
   - Handle context cancellation appropriately
   - Close any opened resources

3. **State Management**
   - Be careful with middleware that maintains state
   - Consider thread safety in concurrent operations
   - Document any stateful behavior

4. **Performance Impact**
   - Monitor the performance impact of each middleware
   - Provide ways to disable expensive middleware
   - Consider adding performance metrics 