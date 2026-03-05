package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypePostgreSQL DatabaseType = "postgres"
)

// Config holds database configuration
type Config struct {
	Type     DatabaseType
	Path     string // For SQLite: path to db file
	Host     string // For PostgreSQL
	Port     int    // For PostgreSQL
	User     string // For PostgreSQL
	Password string // For PostgreSQL
	Database string // For PostgreSQL
	SSLMode  string // For PostgreSQL (disable, require, verify-ca, verify-full)
}

// InitDB initializes the database with connection pooling
// For backward compatibility, it uses SQLite with the provided path
func InitDB(path string) (*gorm.DB, error) {
	cfg := &Config{
		Type: DatabaseTypeSQLite,
		Path: path,
	}
	return InitDBWithConfig(cfg)
}

// InitDBWithConfig initializes the database with the provided configuration
func InitDBWithConfig(cfg *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Type {
	case DatabaseTypePostgreSQL:
		db, err = initPostgreSQL(cfg)
	case DatabaseTypeSQLite:
		db, err = initSQLite(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return nil, err
	}

	// Auto migrate all models
	if err := db.AutoMigrate(
		&Repo{},
		&Commit{},
		&FileIndex{},
		&OAuthState{},
		&User{},
		&APIToken{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	// Get the underlying *sql.DB to configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pooling based on database type
	if cfg.Type == DatabaseTypePostgreSQL {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(0)
		log.Printf("PostgreSQL database initialized: %s@%s:%d/%s (max connections: 50)",
			cfg.User, cfg.Host, cfg.Port, cfg.Database)
	} else {
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(0)
		log.Printf("SQLite database initialized: %s (max connections: 100)", cfg.Path)
	}

	return db, nil
}

// initSQLite initializes a SQLite database
func initSQLite(cfg *Config) (*gorm.DB, error) {
	// Ensure the directory exists for SQLite
	if cfg.Path != ":memory:" {
		dir := cfg.Path
		if lastSlash := len(cfg.Path) - 1; lastSlash > 0 {
			for i := len(cfg.Path) - 1; i >= 0; i-- {
				if cfg.Path[i] == '/' || cfg.Path[i] == '\\' {
					dir = cfg.Path[:i]
					break
				}
			}
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	return db, nil
}

// initPostgreSQL initializes a PostgreSQL database
func initPostgreSQL(cfg *Config) (*gorm.DB, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port, sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// OptimizeDB runs optimization commands for the database
// For SQLite: runs VACUUM and ANALYZE
// For PostgreSQL: runs VACUUM ANALYZE
func OptimizeDB(db *gorm.DB, dbType DatabaseType) error {
	if dbType == DatabaseTypePostgreSQL {
		// PostgreSQL VACUUM ANALYZE
		if err := db.Exec("VACUUM ANALYZE").Error; err != nil {
			return fmt.Errorf("failed to vacuum analyze PostgreSQL: %w", err)
		}
	} else {
		// SQLite VACUUM and ANALYZE
		if err := db.Exec("VACUUM").Error; err != nil {
			return fmt.Errorf("failed to vacuum SQLite: %w", err)
		}
		if err := db.Exec("ANALYZE").Error; err != nil {
			return fmt.Errorf("failed to analyze SQLite: %w", err)
		}
	}
	return nil
}

// GetDBStats returns database connection pool statistics
func GetDBStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return &stats, nil
}

// LoadDatabaseConfigFromEnv loads database configuration from environment variables
func LoadDatabaseConfigFromEnv(defaultPath string) *Config {
	dbType := os.Getenv("HF_LOCAL_DB_TYPE")
	if dbType == "" {
		dbType = string(DatabaseTypeSQLite)
	}

	cfg := &Config{
		Type: DatabaseType(dbType),
	}

	if dbType == string(DatabaseTypePostgreSQL) {
		cfg.Host = getEnvOrDefault("HF_LOCAL_DB_HOST", "localhost")
		cfg.Port = 5432
		if port := os.Getenv("HF_LOCAL_DB_PORT"); port != "" {
			if _, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil {
				log.Printf("Invalid HF_LOCAL_DB_PORT value: %s, using default 5432", port)
			}
		}
		cfg.User = getEnvOrDefault("HF_LOCAL_DB_USER", "postgres")
		cfg.Password = os.Getenv("HF_LOCAL_DB_PASSWORD")
		cfg.Database = getEnvOrDefault("HF_LOCAL_DB_NAME", "hf_local_hub")
		cfg.SSLMode = getEnvOrDefault("HF_LOCAL_DB_SSLMODE", "disable")
	} else {
		cfg.Path = getEnvOrDefault("HF_LOCAL_DB_PATH", defaultPath)
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
