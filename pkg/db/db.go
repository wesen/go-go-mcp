package db

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/db/models"
	"github.com/go-go-golems/go-go-mcp/pkg/db/repository"
	"github.com/go-go-golems/go-go-mcp/pkg/db/repository/gorm"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	g "gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds configuration for database connection
type DatabaseConfig struct {
	// Driver specifies the database type (currently only "sqlite" is supported)
	Driver string `json:"driver" yaml:"driver"`

	// DSN is the data source name (connection string)
	// For SQLite, this is the path to the database file
	DSN string `json:"dsn" yaml:"dsn"`

	// Debug enables GORM debug mode with detailed logging
	Debug bool `json:"debug" yaml:"debug"`

	// Options contains driver-specific options
	Options map[string]interface{} `json:"options" yaml:"options"`
}

// NewDatabase creates a new GORM database connection
func NewDatabase(config DatabaseConfig) (*g.DB, error) {
	var dialector g.Dialector

	switch config.Driver {
	case "sqlite":
		dialector = sqlite.Open(config.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	gormConfig := &g.Config{}
	if config.Debug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := g.Open(dialector, gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database connection")
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.ToolCall{}); err != nil {
		return nil, errors.Wrap(err, "failed to migrate database schema")
	}

	return db, nil
}

// NewRepository creates a new repository instance for the given database
func NewRepository(db *g.DB) repository.ToolCallRepository {
	return gorm.NewGormRepository(db)
}

// InitializeDatabase is a helper function that creates both the database connection
// and repository in one step
func InitializeDatabase(config DatabaseConfig) (repository.ToolCallRepository, error) {
	db, err := NewDatabase(config)
	if err != nil {
		return nil, err
	}

	return NewRepository(db), nil
}
