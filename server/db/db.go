package db

import (
	"database/sql"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
)

// InitDB initializes the SQLite database with connection pooling
func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Repo{}, &Commit{}, &FileIndex{}, &OAuthState{}); err != nil {
		return nil, err
	}

	// Get the underlying *sql.DB to configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pooling for better performance
	// SQLite benefits from reasonable pool settings
	sqlDB.SetMaxIdleConns(25)           // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum open connections
	sqlDB.SetConnMaxLifetime(0)         // Connection lifetime (0 = infinite for SQLite)

	log.Printf("Database initialized: %s (max connections: 100)", path)
	return db, nil
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// OptimizeDB runs VACUUM and ANALYZE to optimize the SQLite database
func OptimizeDB(db *gorm.DB) error {
	// VACUUM to reclaim space
	if err := db.Exec("VACUUM").Error; err != nil {
		return err
	}

	// ANALYZE to update statistics for query planner
	if err := db.Exec("ANALYZE").Error; err != nil {
		return err
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
