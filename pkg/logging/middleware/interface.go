package middleware

import (
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
)

// LoggerMiddleware extends the ToolProvider interface with logging-specific operations
type LoggerMiddleware interface {
	pkg.ToolProvider

	// Flush ensures all pending log records are written
	Flush() error

	// Close cleans up resources and ensures all records are flushed
	Close() error
}

// LoggerConfig holds configuration for the logger middleware
type LoggerConfig struct {
	// BufferSize is the size of the channel buffer for async logging
	BufferSize int

	// FlushInterval is how often to flush pending records
	FlushInterval time.Duration

	// ErrorHandler is called when an error occurs during logging
	ErrorHandler func(error)
}
