package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
)

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Repo{}, &Commit{}, &FileIndex{}); err != nil {
		return nil, err
	}

	log.Printf("Database initialized: %s", path)
	return db, nil
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
