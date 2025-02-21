# MCP Tool Call SQLite Logger Implementation Plan

## Overview
Implement a SQLite-based logging system for MCP tool calls with filtering capabilities and a repository layer for web UI integration.

## Directory Structure
```
pkg/
  db/
    models/
      tool_call.go       # Core model definitions
      filters.go         # Filter types and utilities
    repository/
      tool_call.go       # GORM repository implementation
      queries.go         # Complex query builders
    migrations/
      001_create_tool_calls.go
    db.go               # Database connection and setup
  logging/
    middleware.go       # Tool call logging middleware
    recorder.go        # Async log recording utilities
```

## Implementation Tasks

### Database Models and Setup
- [ ] Create `pkg/db/models/tool_call.go`:
  ```go
  type ToolCall struct {
      gorm.Model
      ToolName    string
      ToolType    string
      Arguments   datatypes.JSON  // Using GORM JSON type
      Output      string         
      StartTime   time.Time
      EndTime     time.Time
      Duration    time.Duration
      SessionID   string         `gorm:"index"`
      ProfileID   string         `gorm:"index"`
      Status      string         // pending, running, completed, error
      Error       string         // Stores error message if any
  }
  ```

- [ ] Create `pkg/db/models/filters.go`:
  ```go
  type ToolCallFilter struct {
      ToolName    *string
      ToolType    *string
      SessionID   *string
      ProfileID   *string
      Status      *string
      TimeRange   *TimeRange
      Pagination  *Pagination
  }

  type TimeRange struct {
      Start *time.Time
      End   *time.Time
  }

  type Pagination struct {
      Page     int
      PageSize int
      OrderBy  string
      Order    string
  }
  ```

### Repository Layer
- [ ] Create `pkg/db/repository/tool_call.go`:
  ```go
  type ToolCallRepository interface {
      Create(ctx context.Context, call *models.ToolCall) error
      Update(ctx context.Context, call *models.ToolCall) error
      GetByID(ctx context.Context, id uint) (*models.ToolCall, error)
      List(ctx context.Context, filter *models.ToolCallFilter) ([]*models.ToolCall, error)
      GetRunning(ctx context.Context) ([]*models.ToolCall, error)
      GetStats(ctx context.Context) (*ToolCallStats, error)
  }
  ```

- [ ] Create `pkg/db/repository/queries.go`:
  ```go
  // Implement query builders for complex filters
  type QueryBuilder interface {
      WithToolName(name string) QueryBuilder
      WithTimeRange(start, end time.Time) QueryBuilder
      WithStatus(status string) QueryBuilder
      WithPagination(page, pageSize int) QueryBuilder
      Build() *gorm.DB
  }
  ```

### Logging Middleware
- [ ] Create `pkg/logging/middleware.go`:
  ```go
  type ToolCallLogger interface {
      Start(ctx context.Context, toolName, toolType string, args interface{}) (context.Context, error)
      Complete(ctx context.Context, output string) error
      Error(ctx context.Context, err error) error
  }
  ```

- [ ] Create `pkg/logging/recorder.go`:
  ```go
  // Async log recording to avoid blocking tool execution
  type LogRecorder interface {
      Record(call *models.ToolCall) error
      Flush() error
  }
  ```

### Database Setup
- [ ] Create `pkg/db/db.go`:
  ```go
  // Database connection management
  type DBManager interface {
      Connect() error
      Close() error
      AutoMigrate() error
      GetDB() *gorm.DB
  }
  ```

### Integration Points
- [ ] Add logging middleware to tool execution pipeline
- [ ] Implement session tracking for tool calls
- [ ] Add profile ID support from configuration
- [ ] Create metrics collection for tool usage statistics

### Performance Considerations
- [ ] Implement connection pooling
- [ ] Add indexes for frequent queries
- [ ] Implement query result caching
- [ ] Handle large output storage efficiently

### Testing
- [ ] Unit tests for models and repositories
- [ ] Integration tests for database operations
- [ ] Performance tests for concurrent logging
- [ ] Mock implementations for testing

## Notes
- Use GORM's hooks for automatic timestamp management
- Implement context cancellation for long-running queries
- Consider implementing soft deletes
- Add proper error wrapping and logging
- Implement connection retry logic
- Consider implementing query timeout mechanisms
