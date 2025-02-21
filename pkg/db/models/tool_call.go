package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ToolCall represents a record of a tool invocation
type ToolCall struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ToolName  string         `gorm:"index"`
	ToolType  string         `gorm:"index"`
	Arguments datatypes.JSON `gorm:"type:json"`
	Output    string         `gorm:"type:text"`
	StartTime time.Time      `gorm:"index"`
	EndTime   time.Time      `gorm:"index"`
	Duration  time.Duration
	SessionID string `gorm:"index"`
	ProfileID string `gorm:"index"`
	Status    string `gorm:"index"`
	Error     string `gorm:"type:text"`
}
