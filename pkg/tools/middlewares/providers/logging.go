package providers

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/middlewares"
)

// LoggingConfig defines configuration options for logging middleware
type LoggingConfig struct {
	LogLevel string
	Format   string
	Output   string
}

type loggingToolProvider struct {
	next pkg.ToolProvider
	log  *slog.Logger
}

// LoggingBuilder helps configure the logging middleware
type LoggingBuilder struct {
	config LoggingConfig
}

// NewLoggingBuilder creates a new logging middleware builder
func NewLoggingBuilder() *LoggingBuilder {
	return &LoggingBuilder{
		config: LoggingConfig{
			LogLevel: "info",
			Format:   "text",
			Output:   "stdout",
		},
	}
}

// WithLogLevel sets the log level
func (b *LoggingBuilder) WithLogLevel(level string) *LoggingBuilder {
	b.config.LogLevel = level
	return b
}

// WithFormat sets the log format
func (b *LoggingBuilder) WithFormat(format string) *LoggingBuilder {
	b.config.Format = format
	return b
}

// WithOutput sets the log output
func (b *LoggingBuilder) WithOutput(output string) *LoggingBuilder {
	b.config.Output = output
	return b
}

// Build creates the logging middleware
func (b *LoggingBuilder) Build() middlewares.ToolProviderMiddleware {
	return func(next pkg.ToolProvider) pkg.ToolProvider {
		var level slog.Level
		if b.config.LogLevel != "" {
			_ = level.UnmarshalText([]byte(b.config.LogLevel))
		}

		// Create a new handler with the specified level
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
		logger := slog.New(handler)

		return &loggingToolProvider{
			next: next,
			log:  logger,
		}
	}
}

func (l *loggingToolProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	start := time.Now()
	tools, nextCursor, err := l.next.ListTools(ctx, cursor)
	duration := time.Since(start)

	l.log.InfoContext(ctx, "ListTools called",
		"duration", duration,
		"cursor", cursor,
		"toolCount", len(tools),
		"nextCursor", nextCursor,
		"error", err)

	return tools, nextCursor, err
}

func (l *loggingToolProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	start := time.Now()
	result, err := l.next.CallTool(ctx, name, arguments)
	duration := time.Since(start)

	l.log.InfoContext(ctx, "CallTool called",
		"duration", duration,
		"tool", name,
		"arguments", arguments,
		"error", err)

	return result, err
}
