package providers

import (
	"context"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/middlewares"
)

type metricsToolProvider struct {
	next      pkg.ToolProvider
	collector MetricsCollector
}

// MetricsBuilder helps configure the metrics middleware
type MetricsBuilder struct {
	collector MetricsCollector
}

// NewMetricsBuilder creates a new metrics middleware builder
func NewMetricsBuilder() *MetricsBuilder {
	return &MetricsBuilder{
		collector: NewNoopMetricsCollector(),
	}
}

// WithCollector sets the metrics collector
func (b *MetricsBuilder) WithCollector(collector MetricsCollector) *MetricsBuilder {
	b.collector = collector
	return b
}

// Build creates the metrics middleware
func (b *MetricsBuilder) Build() middlewares.ToolProviderMiddleware {
	return func(next pkg.ToolProvider) pkg.ToolProvider {
		return &metricsToolProvider{
			next:      next,
			collector: b.collector,
		}
	}
}

func (m *metricsToolProvider) ListTools(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	start := time.Now()
	tools, nextCursor, err := m.next.ListTools(ctx, cursor)
	duration := time.Since(start)

	m.collector.ObserveToolCall("list", "", duration, err)

	return tools, nextCursor, err
}

func (m *metricsToolProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.ToolResult, error) {
	start := time.Now()
	result, err := m.next.CallTool(ctx, name, arguments)
	duration := time.Since(start)

	m.collector.ObserveToolCall("call", name, duration, err)

	return result, err
}
