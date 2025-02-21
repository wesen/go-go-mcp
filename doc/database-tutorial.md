# MCP Database Layer Tutorial

This tutorial explains how to use the MCP database layer for tool call logging. We'll cover everything from basic setup to advanced usage patterns.

## Overview

The MCP database layer is built on top of GORM (Go Object Relational Mapper) and provides a clean, type-safe way to store and retrieve tool call records. The architecture consists of several layers:

1. **Models**: Define the database schema using Go structs
2. **Repository**: Provides a clean interface for data access
3. **GORM Implementation**: Implements the repository interface using GORM
4. **Database Factory**: Handles database connection and configuration

## Quick Start

Here's a minimal example to get you started:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/go-go-golems/go-go-mcp/pkg/db"
)

func main() {
    // Initialize the database
    config := db.DatabaseConfig{
        Driver: "sqlite",
        DSN:    "tool_calls.db", // Will be created in current directory
        Debug:  true,            // Enable GORM debug logging
    }

    repo, err := db.InitializeDatabase(config)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Create a new tool call record
    call := &models.ToolCall{
        ToolName:  "example-tool",
        ToolType:  "shell",
        Arguments: datatypes.JSON([]byte(`{"arg1": "value1"}`)),
        StartTime: time.Now(),
        Status:    "running",
        SessionID: "session-123",
    }

    ctx := context.Background()
    if err := repo.Create(ctx, call); err != nil {
        log.Fatalf("Failed to create tool call: %v", err)
    }
}
```

## Understanding GORM

GORM is a powerful ORM library for Go that handles:

1. **Database Connections**: Manages connection pools and configuration
2. **Schema Management**: Auto-migrates database schema based on Go structs
3. **CRUD Operations**: Provides methods for Create, Read, Update, Delete
4. **Query Building**: Offers a fluent interface for building complex queries

### How GORM Maps Our Models

The `ToolCall` struct is mapped to a database table:

```go
type ToolCall struct {
    ID        uint           `gorm:"primarykey"`  // Primary key, auto-incrementing
    CreatedAt time.Time      // Automatically managed by GORM
    UpdatedAt time.Time      // Automatically managed by GORM
    DeletedAt gorm.DeletedAt `gorm:"index"`      // Enables soft deletes

    // Business fields
    ToolName  string         `gorm:"index"`      // Indexed for faster queries
    ToolType  string         `gorm:"index"`
    Arguments datatypes.JSON  `gorm:"type:json"` // Stored as JSON in database
    // ... other fields
}
```

GORM uses struct tags to configure:
- Primary keys: `gorm:"primarykey"`
- Indexes: `gorm:"index"`
- Column types: `gorm:"type:json"`
- Relationships (not used in our case)

## Repository Pattern

We use the repository pattern to abstract database operations:

1. **Interface**: `ToolCallRepository` defines available operations
2. **Implementation**: `GormRepository` implements these using GORM
3. **Query Builder**: Helps construct complex queries

### Using Filters

The repository supports flexible filtering:

```go
// Get all running tools for a specific session
filter := models.NewToolCallFilter().
    WithStatus("running").
    WithSessionID("session-123")

calls, err := repo.List(ctx, filter)
```

### Pagination

Handle large result sets with pagination:

```go
// Get the second page of results, 20 items per page
filter := models.NewToolCallFilter().
    WithPagination(2, 20).
    WithOrder("created_at", "desc")

calls, err := repo.List(ctx, filter)
```

## Advanced Usage

### Getting Statistics

```go
stats, err := repo.GetStats(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("Total calls: %d", stats.TotalCalls)
log.Printf("Success rate: %.2f%%", 
    float64(stats.SuccessfulCalls)/float64(stats.TotalCalls)*100)
log.Printf("Average duration: %v", stats.AverageDuration)
```

### Maintenance Operations

```go
// Clean up old records
threshold := time.Now().AddDate(0, -1, 0) // 1 month ago
if err := repo.Cleanup(ctx, threshold); err != nil {
    log.Printf("Failed to cleanup old records: %v", err)
}

// Optimize database (SQLite specific)
if err := repo.Vacuum(ctx); err != nil {
    log.Printf("Failed to vacuum database: %v", err)
}
```

## Best Practices

1. **Context Usage**
   - Always pass a context to repository methods
   - Use timeouts for long-running queries
   - Cancel operations when appropriate

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

calls, err := repo.List(ctx, filter)
```

2. **Error Handling**
   - Check errors from all repository operations
   - Use wrapped errors for better context
   - Log errors with appropriate detail

3. **Resource Management**
   - The database connection is managed by GORM
   - No need to explicitly close connections
   - Use context cancellation for query control

4. **Performance Tips**
   - Use indexes (already set up in our models)
   - Keep the database maintained with periodic Vacuum
   - Use pagination for large result sets
   - Monitor query performance with Debug mode

## Debugging

Enable debug mode to see GORM's SQL queries:

```go
config := db.DatabaseConfig{
    Driver: "sqlite",
    DSN:    "tool_calls.db",
    Debug:  true, // Shows all SQL queries
}
```

Example debug output:
```
[info] /home/app/tool_calls.db
[info] CREATE TABLE `tool_calls` ...
[info] SELECT * FROM `tool_calls` WHERE status = "running" ...
```

## Common Patterns

### Transaction Example

```go
func UpdateToolStatus(ctx context.Context, repo repository.ToolCallRepository, id uint, status string) error {
    call, err := repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    call.Status = status
    call.EndTime = time.Now()
    call.Duration = call.EndTime.Sub(call.StartTime)

    return repo.Update(ctx, call)
}
```

### Batch Processing

```go
func CleanupOldSessions(ctx context.Context, repo repository.ToolCallRepository, sessionIDs []string) error {
    for _, id := range sessionIDs {
        filter := models.NewToolCallFilter().WithSessionID(id)
        calls, err := repo.List(ctx, filter)
        if err != nil {
            return err
        }

        for _, call := range calls {
            if err := repo.Update(ctx, call); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## Next Steps

1. Explore the codebase:
   - Look at `pkg/db/models` for data structures
   - Check `pkg/db/repository` for available operations
   - Review `pkg/db/repository/gorm` for implementation details

2. Consider contributing:
   - Add support for other databases
   - Implement new query methods
   - Add more statistics and analytics
   - Improve performance monitoring

3. Integration points:
   - Tool middleware for automatic logging
   - Metrics collection
   - Audit logging
   - Performance analysis 