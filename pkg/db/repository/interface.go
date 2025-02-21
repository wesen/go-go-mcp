package repository

import (
	"context"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/db/models"
)

// ToolCallStats represents statistics about tool calls
type ToolCallStats struct {
	TotalCalls      int64
	SuccessfulCalls int64
	FailedCalls     int64
	AverageDuration time.Duration
	LastCallTime    time.Time
}

// ToolCallRepository defines the interface for tool call data access
type ToolCallRepository interface {
	// Create stores a new tool call record
	Create(ctx context.Context, call *models.ToolCall) error

	// Update modifies an existing tool call record
	Update(ctx context.Context, call *models.ToolCall) error

	// GetByID retrieves a tool call by its ID
	GetByID(ctx context.Context, id uint) (*models.ToolCall, error)

	// List retrieves tool calls based on filter criteria
	List(ctx context.Context, filter *models.ToolCallFilter) ([]*models.ToolCall, error)

	// GetRunning retrieves currently running tool calls
	GetRunning(ctx context.Context) ([]*models.ToolCall, error)

	// GetStats retrieves statistics about tool calls
	GetStats(ctx context.Context) (*ToolCallStats, error)

	// Cleanup removes tool call records older than the specified time
	Cleanup(ctx context.Context, before time.Time) error

	// Vacuum performs database maintenance (implementation specific)
	Vacuum(ctx context.Context) error
}

// RepositoryFactory creates new repository instances
type RepositoryFactory interface {
	// NewRepository creates a new repository with the given configuration
	NewRepository(config map[string]interface{}) (ToolCallRepository, error)
}
