package middleware

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/db/models"
	"github.com/go-go-golems/go-go-mcp/pkg/db/repository"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
)

// dbLoggerProvider implements LoggerMiddleware using a database backend
type dbLoggerProvider struct {
	next         pkg.ToolProvider
	repo         repository.ToolCallRepository
	buffer       chan *models.ToolCall
	done         chan struct{}
	closeOnce    sync.Once
	config       LoggerConfig
	errorHandler func(error)
}

// NewDBLoggerProvider creates a new database logger provider
func NewDBLoggerProvider(next pkg.ToolProvider, repo repository.ToolCallRepository, config LoggerConfig) *dbLoggerProvider {
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(err error) {
			log.Error().Err(err).Msg("Database logging error")
		}
	}

	p := &dbLoggerProvider{
		next:         next,
		repo:         repo,
		buffer:       make(chan *models.ToolCall, config.BufferSize),
		done:         make(chan struct{}),
		config:       config,
		errorHandler: config.ErrorHandler,
	}

	go p.flushLoop()
	return p
}

// ListTools implements ToolProvider interface
func (p *dbLoggerProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	return p.next.ListTools(ctx, cursor)
}

// CallTool implements ToolProvider interface and logs tool calls
func (p *dbLoggerProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	// Create initial tool call record
	call := &models.ToolCall{
		ToolName:  name,
		StartTime: time.Now(),
		Status:    "running",
	}

	// Convert arguments to JSON
	if arguments != nil {
		argsJSON, err := json.Marshal(arguments)
		if err != nil {
			p.errorHandler(errors.Wrap(err, "failed to marshal arguments"))
		} else {
			call.Arguments = datatypes.JSON(argsJSON)
		}
	}

	// Extract session and profile IDs from context if available
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		call.SessionID = sessionID
	}
	if profileID, ok := ctx.Value("profile_id").(string); ok {
		call.ProfileID = profileID
	}

	// Record start
	p.recordCall(call)

	// Call the next provider
	result, err := p.next.CallTool(ctx, name, arguments)

	// Update and record completion
	call.EndTime = time.Now()
	call.Duration = call.EndTime.Sub(call.StartTime)

	if err != nil {
		call.Status = "error"
		call.Error = err.Error()
	} else {
		call.Status = "success"
		if result != nil && result.Content != nil {
			outputJSON, err := json.Marshal(result.Content)
			if err != nil {
				p.errorHandler(errors.Wrap(err, "failed to marshal result"))
			} else {
				call.Output = string(outputJSON)
			}
		}
	}

	p.recordCall(call)
	return result, err
}

// Flush implements LoggerMiddleware interface
func (p *dbLoggerProvider) Flush() error {
	// Create a temporary channel to receive completion signal
	done := make(chan struct{})

	// Send special flush marker
	p.buffer <- &models.ToolCall{Status: "__flush__"}

	// Wait for flush to complete or context to cancel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close implements LoggerMiddleware interface
func (p *dbLoggerProvider) Close() error {
	p.closeOnce.Do(func() {
		close(p.done)
	})
	return p.Flush()
}

// recordCall sends a tool call record to the buffer channel
func (p *dbLoggerProvider) recordCall(call *models.ToolCall) {
	select {
	case p.buffer <- call:
		// Successfully buffered
	default:
		// Buffer full, log synchronously
		if err := p.repo.Create(context.Background(), call); err != nil {
			p.errorHandler(errors.Wrap(err, "failed to record tool call"))
		}
	}
}

// flushLoop runs in a goroutine and periodically flushes the buffer
func (p *dbLoggerProvider) flushLoop() {
	ticker := time.NewTicker(p.config.FlushInterval)
	defer ticker.Stop()

	batch := make([]*models.ToolCall, 0, p.config.BufferSize)

	for {
		select {
		case <-p.done:
			p.flushBatch(batch)
			return

		case call := <-p.buffer:
			if call.Status == "__flush__" {
				p.flushBatch(batch)
				batch = batch[:0]
				continue
			}

			batch = append(batch, call)
			if len(batch) >= p.config.BufferSize {
				p.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				p.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch writes a batch of tool calls to the database
func (p *dbLoggerProvider) flushBatch(batch []*models.ToolCall) {
	for _, call := range batch {
		if err := p.repo.Create(context.Background(), call); err != nil {
			p.errorHandler(errors.Wrap(err, "failed to flush tool call"))
		}
	}
}
