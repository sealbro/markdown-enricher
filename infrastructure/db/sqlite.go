package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
	"markdown-enricher/pkg/logger"
)

type SqliteConfig struct {
	Connection string
}

func MakeSqliteConnection(config *SqliteConfig) (*DB, error) {
	open, err := gorm.Open(sqlite.Open(config.Connection), &gorm.Config{
		Logger: &logger.GormLogger{},
	})

	err = open.Use(prometheus.New(prometheus.Config{
		DBName:          config.Connection,
		RefreshInterval: 15,
	}))

	db := &DB{
		DB: open,
	}

	return db, err
}
