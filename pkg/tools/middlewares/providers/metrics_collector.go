package providers

import "time"

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	// ObserveToolCall records a tool call with its duration and status
	ObserveToolCall(operation string, tool string, duration time.Duration, err error)
}

// NoopMetricsCollector is a metrics collector that does nothing
type NoopMetricsCollector struct{}

func NewNoopMetricsCollector() *NoopMetricsCollector {
	return &NoopMetricsCollector{}
}

func (n *NoopMetricsCollector) ObserveToolCall(operation string, tool string, duration time.Duration, err error) {
	// Do nothing
}
