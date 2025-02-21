---
Title: Using the Database Logger Middleware
Slug: using-db-logger
Short: Learn how to use and configure the database logging middleware for MCP tool calls
Topics:
  - middleware
  - logging
  - database
  - tools
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

This guide will walk you through setting up and using the database logging middleware for MCP tool calls.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Configuration Options](#configuration-options)
4. [Advanced Usage](#advanced-usage)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

## Overview

The database logging middleware provides a robust way to log tool calls to a SQLite database. It supports:

- Asynchronous logging with buffering for better performance
- Comprehensive tool call recording (arguments, results, timing)
- Session and profile tracking
- Error handling and recovery
- Batch processing for efficient database writes

## Quick Start

Here's how to quickly set up the database logger:

```go
import (
    "github.com/go-go-golems/go-go-mcp/pkg/db"
    "github.com/go-go-golems/go-go-mcp/pkg/logging/middleware"
)

// 1. Create database configuration
dbConfig := db.DatabaseConfig{
    Driver: "sqlite",
    DSN:    "tool_calls.db",
    Debug:  true,
}

// 2. Initialize database and repository
repo, err := db.InitializeDatabase(dbConfig)
if err != nil {
    log.Fatal().Err(err).Msg("Failed to initialize database")
}

// 3. Create logger configuration
loggerConfig := middleware.LoggerConfig{
    BufferSize:     100,
    FlushInterval:  5 * time.Second,
    ErrorHandler:   func(err error) {
        log.Error().Err(err).Msg("Logging error")
    },
}

// 4. Create and use the middleware
provider := middleware.NewDBLoggerProvider(baseProvider, repo, loggerConfig)
```

## Configuration Options

### Database Configuration

The `DatabaseConfig` struct supports the following options:

```go
type DatabaseConfig struct {
    // Driver specifies the database type (currently only "sqlite")
    Driver string

    // DSN is the database connection string
    // For SQLite, this is the path to the database file
    DSN string

    // Debug enables GORM debug mode with detailed logging
    Debug bool

    // Options contains driver-specific options
    Options map[string]interface{}
}
```

### Logger Configuration

The `LoggerConfig` struct allows you to customize the logging behavior:

```go
type LoggerConfig struct {
    // BufferSize is the size of the channel buffer for async logging
    BufferSize int

    // FlushInterval is how often to flush pending records
    FlushInterval time.Duration

    // ErrorHandler is called when an error occurs during logging
    ErrorHandler func(error)
}
```

## Advanced Usage

### Custom Error Handling

You can provide custom error handling logic:

```go
loggerConfig := middleware.LoggerConfig{
    ErrorHandler: func(err error) {
        // Send errors to your monitoring system
        metrics.RecordError("tool_logger", err)
        // Log with additional context
        log.Error().
            Err(err).
            Str("component", "tool_logger").
            Msg("Failed to log tool call")
    },
}
```

### Context Values

The middleware automatically extracts session and profile IDs from the context:

```go
ctx := context.WithValue(context.Background(), "session_id", "abc-123")
ctx = context.WithValue(ctx, "profile_id", "user-456")

// These values will be automatically logged
result, err := provider.CallTool(ctx, "my-tool", args)
```

### Manual Flushing

You can manually flush logs when needed:

```go
logger := provider.(middleware.LoggerMiddleware)
if err := logger.Flush(); err != nil {
    log.Error().Err(err).Msg("Failed to flush logs")
}
```

### Graceful Shutdown

Ensure proper cleanup when shutting down:

```go
logger := provider.(middleware.LoggerMiddleware)
if err := logger.Close(); err != nil {
    log.Error().Err(err).Msg("Failed to close logger")
}
```

## Best Practices

1. **Buffer Size**
   - Set buffer size based on expected call volume
   - For high-throughput systems, use larger buffers (500-1000)
   - For low-latency systems, use smaller buffers (50-100)

2. **Flush Interval**
   - Balance between write frequency and database load
   - Recommended range: 1-10 seconds
   - Consider your durability requirements

3. **Error Handling**
   - Always provide an error handler
   - Log errors with sufficient context
   - Consider error aggregation for monitoring

4. **Database Maintenance**
   - Implement periodic cleanup of old records
   - Run VACUUM periodically to reclaim space
   - Monitor database size and growth

Example maintenance code:

```go
// Cleanup old records
before := time.Now().AddDate(0, -1, 0) // 1 month ago
if err := repo.Cleanup(ctx, before); err != nil {
    log.Error().Err(err).Msg("Failed to cleanup old records")
}

// Run VACUUM
if err := repo.Vacuum(ctx); err != nil {
    log.Error().Err(err).Msg("Failed to vacuum database")
}
```

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Reduce buffer size
   - Increase flush frequency
   - Check for blocked database writes

2. **Slow Performance**
   - Enable GORM debug mode to identify slow queries
   - Check disk I/O
   - Consider batch size adjustments

3. **Lost Records**
   - Verify error handler is catching issues
   - Check disk space
   - Review database permissions

### Debugging

Enable GORM debug mode for detailed logging:

```go
dbConfig := db.DatabaseConfig{
    Driver: "sqlite",
    DSN:    "tool_calls.db",
    Debug:  true,
}
```

Monitor the logger with metrics:

```go
loggerConfig := middleware.LoggerConfig{
    ErrorHandler: func(err error) {
        metrics.IncrementCounter("tool_logger_errors")
        log.Error().Err(err).Msg("Logging error")
    },
}
```

### Health Checks

Implement periodic health checks:

```go
func checkLoggerHealth(repo repository.ToolCallRepository) error {
    ctx := context.Background()
    
    // Check if we can write
    testCall := &models.ToolCall{
        ToolName:  "health-check",
        StartTime: time.Now(),
        Status:    "success",
    }
    if err := repo.Create(ctx, testCall); err != nil {
        return fmt.Errorf("write check failed: %w", err)
    }
    
    // Check if we can read
    stats, err := repo.GetStats(ctx)
    if err != nil {
        return fmt.Errorf("read check failed: %w", err)
    }
    
    // Check database stats
    if stats.TotalCalls == 0 {
        return fmt.Errorf("unexpected empty database")
    }
    
    return nil
}
```

## Next Steps

- Explore the query builder for complex data analysis
- Implement custom metrics collection
- Set up automated database maintenance
- Consider implementing a log rotation strategy

For more information, see the [Tool Provider Middlewares](tool-provider-middlewares.md) guide for general middleware concepts and patterns. 