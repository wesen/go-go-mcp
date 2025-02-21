# MCP Tool Call Database Logger Implementation Plan

## Overview
Implement a database-agnostic logging system for MCP tool calls with pluggable storage backends, starting with SQLite support.

## Directory Structure
```
pkg/
  db/
    models/
      tool_call.go       # Core model definitions ✓
      filters.go         # Filter types and utilities ✓
    repository/
      interface.go       # Repository interface definitions
      gorm/
        tool_call.go     # GORM-based repository implementation
        queries.go       # GORM query builders
    migrations/
      001_create_tool_calls.go
    db.go               # Database interface and factory
  logging/
    middleware/
      interface.go      # Middleware interface definitions
      db_logger.go      # Database logging middleware
      builder.go        # Builder interface and implementation
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

### Repository Layer
- [ ] Create `pkg/db/repository/tool_call.go`:
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

  // Factory for creating repositories
  type RepositoryFactory interface {
      NewRepository(config map[string]interface{}) (Repository, error)
  }
  ```

### GORM Implementation
- [ ] Create `pkg/db/repository/gorm/tool_call.go`:
  ```go
  type GormRepository struct {
      db *gorm.DB
      queryBuilder *QueryBuilder
  }

  func NewGormRepository(db *gorm.DB) *GormRepository {
      return &GormRepository{
          db: db,
          queryBuilder: NewQueryBuilder(),
      }
  }

  // Implement Repository interface methods...
  ```

### Logging Middleware Interface
- [ ] Create `pkg/logging/middleware/interface.go`:
  ```go
  type LoggerMiddleware interface {
      pkg.ToolProvider
      Flush() error
      Close() error
  }

  type LoggerBuilder interface {
      WithRepository(repo repository.Repository) LoggerBuilder
      WithBufferSize(size int) LoggerBuilder
      WithFlushInterval(d time.Duration) LoggerBuilder
      WithErrorHandler(handler func(error)) LoggerBuilder
      Build() middlewares.ToolProviderMiddleware
  }
  ```

### Database Logger Implementation
- [ ] Create `pkg/logging/middleware/db_logger.go`:
  ```go
  type dbLoggerProvider struct {
      next pkg.ToolProvider
      recorder recorder.Recorder
  }

  func (d *dbLoggerProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
      call := &models.ToolCall{
          ToolName: name,
          Arguments: arguments,
          StartTime: time.Now(),
          Status: "running",
      }

      // Extract metadata from context
      d.enrichCallWithMetadata(ctx, call)

      // Record start
      if err := d.recorder.Record(call); err != nil {
          log.Error().Err(err).Msg("Failed to record tool call start")
      }

      // Call next provider
      result, err := d.next.CallTool(ctx, name, arguments)

      // Update and record completion
      d.updateCallWithResult(call, result, err)
      if recordErr := d.recorder.Record(call); recordErr != nil {
          log.Error().Err(recordErr).Msg("Failed to record tool call completion")
      }

      return result, err
  }
  ```

### Database Recorder Interface
- [ ] Create `pkg/logging/recorder/interface.go`:
  ```go
  type Recorder interface {
      Record(call *models.ToolCall) error
      Flush() error
      Close() error
  }

  type RecorderFactory interface {
      NewRecorder(repo repository.Repository, config RecorderConfig) (Recorder, error)
  }

  type RecorderConfig struct {
      BufferSize     int
      FlushInterval  time.Duration
      ErrorHandler   func(error)
  }
  ```

### Database Recorder Implementation
- [ ] Create `pkg/logging/recorder/db/recorder.go`:
  ```go
  type dbRecorder struct {
      repo       repository.Repository
      buffer     chan *models.ToolCall
      config     RecorderConfig
      done       chan struct{}
      closeOnce  sync.Once
  }

  func NewDBRecorder(repo repository.Repository, config RecorderConfig) *dbRecorder {
      r := &dbRecorder{
          repo:   repo,
          buffer: make(chan *models.ToolCall, config.BufferSize),
          config: config,
          done:   make(chan struct{}),
      }
      
      go r.flushLoop()
      return r
  }

  func (r *dbRecorder) Record(call *models.ToolCall) error {
      select {
      case r.buffer <- call:
          return nil
      default:
          // Buffer full, flush synchronously
          return r.repo.Create(context.Background(), call)
      }
  }

  func (r *dbRecorder) flushLoop() {
      ticker := time.NewTicker(r.config.FlushInterval)
      defer ticker.Stop()

      batch := make([]*models.ToolCall, 0, r.config.BufferSize)
      
      for {
          select {
          case <-r.done:
              return
          case call := <-r.buffer:
              batch = append(batch, call)
              if len(batch) >= r.config.BufferSize {
                  r.flushBatch(batch)
                  batch = batch[:0]
              }
          case <-ticker.C:
              if len(batch) > 0 {
                  r.flushBatch(batch)
                  batch = batch[:0]
              }
          }
      }
  }

  func (r *dbRecorder) flushBatch(batch []*models.ToolCall) {
      ctx := context.Background()
      for _, call := range batch {
          if err := r.repo.Create(ctx, call); err != nil && r.config.ErrorHandler != nil {
              r.config.ErrorHandler(err)
          }
      }
  }
  ```

### Builder Implementation
- [ ] Create `pkg/logging/middleware/builder.go`:
  ```go
  type dbLoggerBuilder struct {
      repo          repository.Repository
      recorderConfig recorder.RecorderConfig
  }

  func NewDBLoggerBuilder() *dbLoggerBuilder {
      return &dbLoggerBuilder{
          recorderConfig: recorder.RecorderConfig{
              BufferSize:    100,
              FlushInterval: 5 * time.Second,
          },
      }
  }

  func (b *dbLoggerBuilder) WithRepository(repo repository.Repository) LoggerBuilder {
      b.repo = repo
      return b
  }

  func (b *dbLoggerBuilder) Build() middlewares.ToolProviderMiddleware {
      recorder := recorder.NewDBRecorder(b.repo, b.recorderConfig)
      
      return func(next pkg.ToolProvider) pkg.ToolProvider {
          return &dbLoggerProvider{
              next: next,
              recorder: recorder,
          }
      }
  }
  ```

### Usage Example

```go
// Create a SQLite repository
config := map[string]interface{}{
    "driver": "sqlite",
    "dsn": "logs/tools.db",
}
repo, err := repository.NewRepository(config)
if err != nil {
    log.Fatal().Err(err).Msg("Failed to create repository")
}

// Create a tool provider with database logging
provider := NewChainBuilder().
    With(NewDBLoggerBuilder().
        WithRepository(repo).
        WithBufferSize(200).
        WithFlushInterval(10 * time.Second).
        WithErrorHandler(func(err error) {
            log.Error().Err(err).Msg("Database logging error")
        }).
        Build()).
    Build(baseProvider)
```

### Integration Points
- [ ] Add logging middleware to tool execution pipeline
- [ ] Implement session tracking for tool calls
- [ ] Add profile ID support from configuration
- [ ] Create metrics collection for tool usage statistics

### Performance Considerations
- [ ] Use buffered channels for async logging
- [ ] Implement batched inserts for better performance
- [ ] Add indexes for frequent queries

### Testing
- [ ] Unit tests for models and repositories
- [ ] Integration tests for database operations
- [ ] Performance tests for concurrent logging
- [ ] Mock implementations for testing
- [ ] Test buffer overflow scenarios
- [ ] Test flush interval behavior
- [ ] Test error handling and recovery
- [ ] Test different database backends

## Notes
- Implement context cancellation for long-running queries
- Use prepared statements for better performance
- Implement periodic vacuum for database maintenance
- Consider implementing log rotation by date
- Consider adding support for other databases (PostgreSQL, MySQL)
