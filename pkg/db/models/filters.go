package models

import "time"

// ToolCallFilter defines the filtering options for tool call queries
type ToolCallFilter struct {
	ToolName   *string
	ToolType   *string
	SessionID  *string
	ProfileID  *string
	Status     *string
	TimeRange  *TimeRange
	Pagination *Pagination
}

// TimeRange represents a time period for filtering
type TimeRange struct {
	Start *time.Time
	End   *time.Time
}

// Pagination defines the pagination parameters for queries
type Pagination struct {
	Page     int
	PageSize int
	OrderBy  string
	Order    string
}

// NewToolCallFilter creates a new ToolCallFilter with default values
func NewToolCallFilter() *ToolCallFilter {
	return &ToolCallFilter{
		Pagination: &Pagination{
			Page:     1,
			PageSize: 20,
			OrderBy:  "created_at",
			Order:    "desc",
		},
	}
}

// WithToolName adds a tool name filter
func (f *ToolCallFilter) WithToolName(name string) *ToolCallFilter {
	f.ToolName = &name
	return f
}

// WithToolType adds a tool type filter
func (f *ToolCallFilter) WithToolType(toolType string) *ToolCallFilter {
	f.ToolType = &toolType
	return f
}

// WithSessionID adds a session ID filter
func (f *ToolCallFilter) WithSessionID(sessionID string) *ToolCallFilter {
	f.SessionID = &sessionID
	return f
}

// WithProfileID adds a profile ID filter
func (f *ToolCallFilter) WithProfileID(profileID string) *ToolCallFilter {
	f.ProfileID = &profileID
	return f
}

// WithStatus adds a status filter
func (f *ToolCallFilter) WithStatus(status string) *ToolCallFilter {
	f.Status = &status
	return f
}

// WithTimeRange adds a time range filter
func (f *ToolCallFilter) WithTimeRange(start, end time.Time) *ToolCallFilter {
	f.TimeRange = &TimeRange{
		Start: &start,
		End:   &end,
	}
	return f
}

// WithPagination sets pagination parameters
func (f *ToolCallFilter) WithPagination(page, pageSize int) *ToolCallFilter {
	if f.Pagination == nil {
		f.Pagination = &Pagination{}
	}
	f.Pagination.Page = page
	f.Pagination.PageSize = pageSize
	return f
}

// WithOrder sets ordering parameters
func (f *ToolCallFilter) WithOrder(orderBy, order string) *ToolCallFilter {
	if f.Pagination == nil {
		f.Pagination = &Pagination{}
	}
	f.Pagination.OrderBy = orderBy
	f.Pagination.Order = order
	return f
}
