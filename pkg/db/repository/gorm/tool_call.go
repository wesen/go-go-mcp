package gorm

import (
	"context"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/db/models"
	"github.com/go-go-golems/go-go-mcp/pkg/db/repository"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GormRepository implements ToolCallRepository using GORM
type GormRepository struct {
	db           *gorm.DB
	queryBuilder *QueryBuilder
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db:           db,
		queryBuilder: NewQueryBuilder(db),
	}
}

// Create stores a new tool call record
func (r *GormRepository) Create(ctx context.Context, call *models.ToolCall) error {
	result := r.db.WithContext(ctx).Create(call)
	return errors.Wrap(result.Error, "failed to create tool call record")
}

// Update modifies an existing tool call record
func (r *GormRepository) Update(ctx context.Context, call *models.ToolCall) error {
	result := r.db.WithContext(ctx).Save(call)
	return errors.Wrap(result.Error, "failed to update tool call record")
}

// GetByID retrieves a tool call by its ID
func (r *GormRepository) GetByID(ctx context.Context, id uint) (*models.ToolCall, error) {
	var call models.ToolCall
	result := r.db.WithContext(ctx).First(&call, id)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to get tool call record")
	}
	return &call, nil
}

// List retrieves tool calls based on filter criteria
func (r *GormRepository) List(ctx context.Context, filter *models.ToolCallFilter) ([]*models.ToolCall, error) {
	var calls []*models.ToolCall
	query := r.queryBuilder.BuildToolCallQuery(filter).WithContext(ctx)

	result := query.Find(&calls)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to list tool call records")
	}
	return calls, nil
}

// GetRunning retrieves currently running tool calls
func (r *GormRepository) GetRunning(ctx context.Context) ([]*models.ToolCall, error) {
	filter := models.NewToolCallFilter().WithStatus("running")
	return r.List(ctx, filter)
}

// GetStats retrieves statistics about tool calls
func (r *GormRepository) GetStats(ctx context.Context) (*repository.ToolCallStats, error) {
	var stats repository.ToolCallStats
	var lastCall models.ToolCall

	// Get total calls
	if err := r.db.WithContext(ctx).Model(&models.ToolCall{}).Count(&stats.TotalCalls).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get total calls")
	}

	// Get successful and failed calls
	if err := r.db.WithContext(ctx).Model(&models.ToolCall{}).Where("error = ''").Count(&stats.SuccessfulCalls).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get successful calls")
	}
	stats.FailedCalls = stats.TotalCalls - stats.SuccessfulCalls

	// Get average duration
	var avgDuration float64
	if err := r.db.WithContext(ctx).Model(&models.ToolCall{}).
		Select("avg(duration)").
		Row().
		Scan(&avgDuration); err != nil {
		return nil, errors.Wrap(err, "failed to get average duration")
	}
	stats.AverageDuration = time.Duration(avgDuration)

	// Get last call time
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		First(&lastCall).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(err, "failed to get last call time")
	}
	if err != gorm.ErrRecordNotFound {
		stats.LastCallTime = lastCall.CreatedAt
	}

	return &stats, nil
}

// Cleanup removes tool call records older than the specified time
func (r *GormRepository) Cleanup(ctx context.Context, before time.Time) error {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&models.ToolCall{})
	return errors.Wrap(result.Error, "failed to cleanup old tool call records")
}

// Vacuum performs database maintenance
func (r *GormRepository) Vacuum(ctx context.Context) error {
	// For SQLite, we can execute VACUUM directly
	result := r.db.WithContext(ctx).Exec("VACUUM")
	return errors.Wrap(result.Error, "failed to vacuum database")
}
