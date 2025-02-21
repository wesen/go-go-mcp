# MCP Database Layer Tutorial

This tutorial explains how to use the MCP database layer for tool call logging. We'll cover everything from basic setup to advanced usage patterns.

## What is GORM and Why Do We Use It?

[GORM](https://gorm.io) is the most widely-used Object-Relational Mapping (ORM) library for Go, and it's the foundation of MCP's database layer. Think of GORM as a smart translator between your Go code and your database - it handles all the complex SQL operations behind the scenes while letting you work with familiar Go structs.

### The Power of GORM

GORM makes database operations feel natural in Go. Instead of writing raw SQL like:
```sql
INSERT INTO tool_calls (tool_name, status, created_at) VALUES ('example', 'running', CURRENT_TIMESTAMP);
```

You can write Go code that's type-safe and intuitive:
```go
toolCall := &models.ToolCall{
    ToolName: "example",
    Status:   "running",
}
db.Create(toolCall) // GORM handles the rest!
```

But GORM is more than just a convenience layer. It provides:

1. **Smart Schema Management**: Through [Auto Migration](https://gorm.io/docs/migration.html), GORM can automatically create and update your database schema based on your Go structs. It's like having a database admin who automatically keeps your tables in sync with your code.

2. **Type Safety**: No more string concatenation for SQL queries or type conversion headaches. GORM understands Go types and handles the mapping to database types automatically.

3. **Hooks and Callbacks**: Want to automatically set timestamps, validate data, or trigger events before/after database operations? GORM's [hooks system](https://gorm.io/docs/hooks.html) has you covered.

4. **Powerful Query Interface**: From simple CRUD to complex joins and transactions, GORM provides an expressive API that makes database queries feel like natural Go operations.

### How MCP Uses GORM

In MCP, we've built a robust database layer around GORM that follows best practices and provides a clean, maintainable architecture:

1. **Model-Driven Design**: Our database schema is defined through Go structs with GORM tags, making it self-documenting and type-safe. For example, our `ToolCall` model clearly shows what data we store and how it's indexed.

2. **Repository Pattern**: Instead of using GORM directly throughout the codebase, we encapsulate all database operations behind a clean repository interface. This makes our code more testable and maintainable.

3. **Automatic Migrations**: During development, GORM's AutoMigrate feature keeps our database schema in sync with our code, making it easy to iterate and evolve our data model.

4. **Connection Management**: GORM handles connection pooling, reconnection, and other low-level database concerns, letting us focus on business logic.

### Real-World Benefits

This architecture has several practical advantages:

1. **Development Speed**: Adding new fields to our models automatically updates the database schema - no manual migrations needed during development.

2. **Type Safety**: The Go compiler catches many potential errors before they hit production, thanks to our type-safe repository interface.

3. **Maintainability**: By abstracting database operations behind a clean interface, we can change our database implementation without affecting the rest of the codebase.

4. **Performance**: GORM's connection pooling and efficient query building help maintain good performance under load.

Let's dive deeper into how this all works in practice...

## Overview

The MCP database layer is built on top of GORM and provides a clean, type-safe way to store and retrieve tool call records. The architecture consists of several layers:

1. **Models**: Define the database schema using Go structs
2. **Repository**: Provides a clean interface for data access
3. **GORM Implementation**: Implements the repository interface using GORM
4. **Database Factory**: Handles database connection and configuration

## Understanding db.go and GORM Integration

The `db.go` file is the core of our database initialization and configuration. Let's break down how it works with GORM:

### Database Configuration

```go
type DatabaseConfig struct {
    // Driver specifies the database type (currently only "sqlite" is supported)
    Driver string `json:"driver" yaml:"driver"`

    // DSN is the data source name (connection string)
    // For SQLite, this is the path to the database file
    DSN string `json:"dsn" yaml:"dsn"`

    // Debug enables GORM debug mode with detailed logging
    Debug bool `json:"debug" yaml:"debug"`

    // Options contains driver-specific options
    Options map[string]interface{} `json:"options" yaml:"options"`
}
```

The `DatabaseConfig` struct provides a flexible way to configure database connections. Currently, it supports SQLite, but the structure is designed to be extensible for other database types in the future.

### Database Initialization Flow

Let's walk through how `NewDatabase` works:

1. **Dialector Selection**:
   ```go
   var dialector g.Dialector
   switch config.Driver {
   case "sqlite":
       dialector = sqlite.Open(config.DSN)
   default:
       return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
   }
   ```
   - A dialector is GORM's way of handling different database types
   - Currently only SQLite is supported, but adding new drivers is straightforward

2. **GORM Configuration**:
   ```go
   gormConfig := &g.Config{}
   if config.Debug {
       gormConfig.Logger = logger.Default.LogMode(logger.Info)
   }
   ```
   - Debug mode enables SQL query logging
   - GORM's logger helps with development and troubleshooting

3. **Database Connection**:
   ```go
   db, err := g.Open(dialector, gormConfig)
   if err != nil {
       return nil, errors.Wrap(err, "failed to open database connection")
   }
   ```
   - Opens the database connection using the configured dialector
   - Returns a GORM DB instance that manages the connection pool

4. **Auto-Migration**:
   ```go
   if err := db.AutoMigrate(&models.ToolCall{}); err != nil {
       return nil, errors.Wrap(err, "failed to migrate database schema")
   }
   ```
   - GORM's AutoMigrate automatically creates or updates database tables
   - It reads the struct definitions and their GORM tags
   - Creates tables, adds missing columns, and updates column types
   - Never deletes existing columns or data

### Understanding AutoMigrate

GORM's AutoMigrate is a powerful feature that handles database schema management. Here's what it does when you call `db.AutoMigrate(&models.ToolCall{})`:

1. **Table Creation**:
   - If the table doesn't exist, creates it based on the struct
   - Table name is derived from struct name (snake_case)

2. **Column Management**:
   - Creates columns for each struct field
   - Uses GORM tags to determine:
     - Column types
     - Constraints
     - Indexes
     - Default values

3. **Safe Schema Updates**:
   - Adds new columns if they exist in the struct but not in the database
   - Updates column types if they've changed in the struct
   - Preserves existing data
   - Never removes columns (even if they're removed from the struct)

Example of how GORM maps our `ToolCall` struct:

```go
type ToolCall struct {
    ID        uint           `gorm:"primarykey"`       // Creates INTEGER PRIMARY KEY
    CreatedAt time.Time                               // DATETIME column
    UpdatedAt time.Time                               // DATETIME column
    DeletedAt gorm.DeletedAt `gorm:"index"`           // Adds index for soft deletes
    
    ToolName  string         `gorm:"index;size:255"`  // VARCHAR(255) with index
    ToolType  string         `gorm:"index"`           // Creates index
    Arguments datatypes.JSON `gorm:"type:json"`       // JSON column type
    // ... other fields
}
```

This results in SQL like:
```sql
CREATE TABLE tool_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    tool_name VARCHAR(255),
    tool_type VARCHAR(255),
    arguments JSON,
    -- ... other columns
    INDEX idx_tool_calls_deleted_at (deleted_at),
    INDEX idx_tool_calls_tool_name (tool_name),
    INDEX idx_tool_calls_tool_type (tool_type)
);
```

### Repository Pattern Integration

After database initialization, we create a repository instance:

```go
func NewRepository(db *g.DB) repository.ToolCallRepository {
    return gorm.NewGormRepository(db)
}
```

The repository pattern provides a clean interface for database operations:
- Abstracts GORM implementation details
- Makes testing easier (can mock the repository)
- Provides type-safe methods for common operations

### Complete Initialization

The `InitializeDatabase` function ties everything together:

```go
func InitializeDatabase(config DatabaseConfig) (repository.ToolCallRepository, error) {
    db, err := NewDatabase(config)
    if err != nil {
        return nil, err
    }
    return NewRepository(db), nil
}
```

This gives you a ready-to-use repository with:
- Configured database connection
- Migrated schema
- Connection pool management
- Query logging (if debug enabled)

### Best Practices

1. **Configuration**:
   - Always provide a DSN appropriate for your environment
   - Enable debug mode during development
   - Use Options map for driver-specific settings

2. **Schema Management**:
   - Let AutoMigrate handle schema updates in development
   - Use manual migrations in production
   - Keep struct tags up to date

3. **Connection Management**:
   - GORM handles the connection pool
   - No need to manually close connections
   - Use context for query timeouts

4. **Error Handling**:
   - Check initialization errors
   - Use wrapped errors for better context
   - Monitor migration success

Example usage with best practices:

```go
func initializeWithBestPractices() (*repository.ToolCallRepository, error) {
    config := DatabaseConfig{
        Driver: "sqlite",
        DSN:    "file:tool_calls.db?cache=shared&mode=rwc",
        Debug:  true,
        Options: map[string]interface{}{
            "pragma": map[string]string{
                "journal_mode": "WAL",
                "busy_timeout": "5000",
            },
        },
    }

    repo, err := InitializeDatabase(config)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize database: %w", err)
    }

    return repo, nil
}
```

## Database Initialization and Migrations

GORM handles database schema management through its AutoMigrate feature, which automatically creates or updates tables based on your model definitions. The MCP database layer encapsulates this in the `NewDatabase` function.

### Basic Initialization

Here's the simplest way to initialize the database:

```go
package main

import (
    "github.com/go-go-golems/go-go-mcp/pkg/db"
)

func main() {
    // Create database configuration
    config := db.DatabaseConfig{
        Driver: "sqlite",            // Currently only SQLite is supported
        DSN:    "tool_calls.db",    // Database file path
        Debug:  true,               // Enable SQL query logging
    }

    // Initialize database and repository in one step
    repo, err := db.InitializeDatabase(config)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to initialize database")
    }

    // The database is now ready to use
    // Tables are automatically created/updated
}
```

### What Happens During Initialization

1. **Connection Setup**:
   ```go
   // Inside pkg/db/db.go
   db, err := gorm.Open(dialector, gormConfig)
   ```

2. **Auto-Migration**:
   ```go
   // This creates/updates tables for all registered models
   if err := db.AutoMigrate(&models.ToolCall{}); err != nil {
       return nil, errors.Wrap(err, "failed to migrate database schema")
   }
   ```

3. **Repository Creation**:
   ```go
   // Creates a repository instance with the initialized database
   repo := gorm.NewGormRepository(db)
   ```

### Understanding Auto-Migration

GORM's AutoMigrate:
- Creates tables if they don't exist
- Adds missing columns
- Updates column types if changed
- Creates indexes defined in model tags
- Never deletes columns or indexes

Example of how model tags affect the schema:
```go
type ToolCall struct {
    ID        uint           `gorm:"primarykey"`       // Creates primary key
    CreatedAt time.Time                               // Automatic timestamps
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`           // Adds index for soft deletes
    
    ToolName  string         `gorm:"index;size:255"`  // Creates index, sets column size
    Arguments datatypes.JSON `gorm:"type:json"`       // Uses JSON column type
    // ... other fields
}
```

### Manual Schema Management

While auto-migration is convenient for development, you might want more control in production:

```go
func initializeWithManualMigration(config db.DatabaseConfig) error {
    // Get raw database connection
    gormDB, err := db.NewDatabase(config)
    if err != nil {
        return err
    }

    // Get underlying SQL database
    sqlDB, err := gormDB.DB()
    if err != nil {
        return err
    }

    // Perform manual migrations if needed
    if err := runMigrations(sqlDB); err != nil {
        return err
    }

    // Create repository
    repo := gorm.NewGormRepository(gormDB)
    return nil
}

func runMigrations(db *sql.DB) error {
    migrations := []string{
        `CREATE TABLE IF NOT EXISTS tool_calls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at DATETIME,
            updated_at DATETIME,
            deleted_at DATETIME,
            tool_name TEXT,
            arguments JSON,
            -- ... other columns
            INDEX idx_tool_name (tool_name)
        )`,
        // Add more migration statements
    }

    for _, migration := range migrations {
        if _, err := db.Exec(migration); err != nil {
            return err
        }
    }
    return nil
}
```

## Quick Start

Now that we understand initialization, here's a complete example:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/go-go-golems/go-go-mcp/pkg/db"
    "github.com/go-go-golems/go-go-mcp/pkg/db/models"
)

func main() {
    // 1. Initialize database
    config := db.DatabaseConfig{
        Driver: "sqlite",
        DSN:    "tool_calls.db",
        Debug:  true,
    }

    repo, err := db.InitializeDatabase(config)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // 2. Create a new tool call record
    call := &models.ToolCall{
        ToolName:  "example-tool",
        ToolType:  "shell",
        Arguments: datatypes.JSON([]byte(`{"arg1": "value1"}`)),
        StartTime: time.Now(),
        Status:    "running",
        SessionID: "session-123",
    }

    // 3. Store the record
    ctx := context.Background()
    if err := repo.Create(ctx, call); err != nil {
        log.Fatalf("Failed to create tool call: %v", err)
    }

    // 4. Query the record
    filter := models.NewToolCallFilter().
        WithToolName("example-tool").
        WithStatus("running")

    calls, err := repo.List(ctx, filter)
    if err != nil {
        log.Fatalf("Failed to query tool calls: %v", err)
    }

    for _, c := range calls {
        log.Printf("Found tool call: %s (status: %s)", c.ToolName, c.Status)
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