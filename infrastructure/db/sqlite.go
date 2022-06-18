package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"markdown-enricher/pkg/logger"
)

type SqliteConfig struct {
	Connection string
}

func MakeSqliteConnection(config *SqliteConfig) (*DB, error) {
	open, err := gorm.Open(sqlite.Open(config.Connection), &gorm.Config{
		Logger: &logger.GormLogger{},
	})

	db := &DB{
		DB: open,
	}

	return db, err
}
