# MCP Tool Call Database Logger Implementation Plan

## Overview
Implement a database-agnostic logging system for MCP tool calls with pluggable storage backends, starting with SQLite support.

## Directory Structure ✓
```
pkg/
  db/
    models/
      tool_call.go       # Core model definitions ✓
      filters.go         # Filter types and utilities ✓
    repository/
      interface.go       # Repository interface definitions ✓
      gorm/
        tool_call.go     # GORM-based repository implementation ✓
        queries.go       # GORM query builders ✓
    migrations/
      001_create_tool_calls.go # Handled by GORM auto-migrate ✓
    db.go               # Database interface and factory ✓
  logging/
    middleware/
      interface.go      # Middleware interface definitions ✓
      db_logger.go      # Database logging middleware ✓
      builder.go        # Builder interface and implementation ✓
    recorder/
      interface.go      # Recorder interface definitions
      db/
        recorder.go     # Database recorder implementation
```

## Implementation Tasks

### Core Database Models ✓
- [x] Create `pkg/db/models/tool_call.go`:
  ```go
  type ToolCall struct {
      ID          uint           `gorm:"primarykey"`
      CreatedAt   time.Time
      UpdatedAt   time.Time
      DeletedAt   gorm.DeletedAt `gorm:"index"`
      
      ToolName    string
      ToolType    string
      Arguments   datatypes.JSON  
      Output      string         
      StartTime   time.Time
      EndTime     time.Time
      Duration    time.Duration
      SessionID   string         `gorm:"index"`
      ProfileID   string         `gorm:"index"`
      Status      string         
      Error       string         
  }
  ```

- [x] Create `pkg/db/models/filters.go`:
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

### Repository Layer ✓
- [x] Create `pkg/db/repository/interface.go`:
  ```go
  type ToolCallRepository interface {
      Create(ctx context.Context, call *models.ToolCall) error
      Update(ctx context.Context, call *models.ToolCall) error
      GetByID(ctx context.Context, id uint) (*models.ToolCall, error)
      
      // Query operations
      List(ctx context.Context, filter *models.ToolCallFilter) ([]*models.ToolCall, error)
      GetRunning(ctx context.Context) ([]*models.ToolCall, error)
      GetStats(ctx context.Context) (*ToolCallStats, error)
      
      // Maintenance operations
      Cleanup(ctx context.Context, before time.Time) error
      Vacuum(ctx context.Context) error
  }
  ```

- [x] Create `pkg/db/repository/gorm/tool_call.go`:
  ```go
  type GormRepository struct {
      db *gorm.DB
      queryBuilder *QueryBuilder
  }
  ```

- [x] Create `pkg/db/repository/gorm/queries.go`:
  ```go
  type QueryBuilder struct {
      db *gorm.DB
  }
  ```

### Database Factory and Migrations ✓
- [x] Create `pkg/db/db.go`:
  ```go
  type DatabaseConfig struct {
      Driver   string
      DSN      string
      Options  map[string]interface{}
  }

  func NewDatabase(config DatabaseConfig) (*gorm.DB, error)
  ```

- [x] Create migrations (using GORM auto-migrate) ✓

### Logging Middleware ✓
- [x] Create `pkg/logging/middleware/interface.go`:
  ```go
  type LoggerMiddleware interface {
      pkg.ToolProvider
      Flush() error
      Close() error
  }
  ```

- [x] Create `pkg/logging/middleware/db_logger.go`:
  ```go
  type dbLoggerProvider struct {
      next pkg.ToolProvider
      recorder recorder.Recorder
  }
  ```

- [x] Create `pkg/logging/middleware/builder.go`:
  ```go
  type LoggerBuilder interface {
      WithRepository(repo repository.Repository) LoggerBuilder
      WithBufferSize(size int) LoggerBuilder
      WithFlushInterval(d time.Duration) LoggerBuilder
      WithErrorHandler(handler func(error)) LoggerBuilder
      Build() middlewares.ToolProviderMiddleware
  }
  ```

### Recorder Implementation
- [ ] Create `pkg/logging/recorder/interface.go`:
  ```go
  type Recorder interface {
      Record(call *models.ToolCall) error
      Flush() error
      Close() error
  }
  ```

- [ ] Create `pkg/logging/recorder/db/recorder.go`:
  ```go
  type dbRecorder struct {
      repo       repository.Repository
      buffer     chan *models.ToolCall
      config     RecorderConfig
      done       chan struct{}
      closeOnce  sync.Once
  }
  ```

## Notes and Future Improvements
- [x] Implement context cancellation for long-running queries
- [x] Use prepared statements for better performance
- [x] Implement periodic vacuum for database maintenance
- [x] Consider implementing log rotation by date
- [ ] Add support for other databases (PostgreSQL, MySQL)
- [ ] Add query caching for frequently accessed data
- [ ] Add support for bulk operations
- [ ] Add support for transaction management
- [ ] Add support for database connection pooling
- [ ] Add support for database connection retries
- [ ] Add support for database connection timeouts
- [ ] Add support for database connection keep-alive
- [ ] Add support for database connection monitoring
- [ ] Add support for database connection metrics
