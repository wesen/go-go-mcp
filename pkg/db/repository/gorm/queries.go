package gorm

import (
	"github.com/go-go-golems/go-go-mcp/pkg/db/models"
	"gorm.io/gorm"
)

// QueryBuilder helps construct complex GORM queries
type QueryBuilder struct {
	db *gorm.DB
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db: db,
	}
}

// ApplyToolCallFilter applies the filter criteria to the query
func (b *QueryBuilder) ApplyToolCallFilter(query *gorm.DB, filter *models.ToolCallFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	if filter.ToolName != nil {
		query = query.Where("tool_name = ?", *filter.ToolName)
	}
	if filter.ToolType != nil {
		query = query.Where("tool_type = ?", *filter.ToolType)
	}
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", *filter.SessionID)
	}
	if filter.ProfileID != nil {
		query = query.Where("profile_id = ?", *filter.ProfileID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.TimeRange != nil {
		if filter.TimeRange.Start != nil {
			query = query.Where("start_time >= ?", filter.TimeRange.Start)
		}
		if filter.TimeRange.End != nil {
			query = query.Where("end_time <= ?", filter.TimeRange.End)
		}
	}

	return query
}

// ApplyPagination applies pagination parameters to the query
func (b *QueryBuilder) ApplyPagination(query *gorm.DB, pagination *models.Pagination) *gorm.DB {
	if pagination == nil {
		return query
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query = query.Offset(offset).Limit(pagination.PageSize)

	if pagination.OrderBy != "" {
		order := pagination.Order
		if order == "" {
			order = "asc"
		}
		query = query.Order(pagination.OrderBy + " " + order)
	}

	return query
}

// BuildToolCallQuery builds a complete query for tool calls
func (b *QueryBuilder) BuildToolCallQuery(filter *models.ToolCallFilter) *gorm.DB {
	query := b.db.Model(&models.ToolCall{})
	query = b.ApplyToolCallFilter(query, filter)
	if filter != nil && filter.Pagination != nil {
		query = b.ApplyPagination(query, filter.Pagination)
	}
	return query
}
