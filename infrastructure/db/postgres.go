package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
	"markdown-enricher/pkg/logger"
)

type PostgresConfig struct {
	Connection string
	Schema     string
}

type DB struct {
	*gorm.DB
}

func MakePostgresConnection(config *PostgresConfig) (*DB, error) {
	open, err := gorm.Open(postgres.Open(config.Connection), &gorm.Config{
		Logger: &logger.GormLogger{},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: config.Schema + ".",
		},
	})

	open.Use(prometheus.New(prometheus.Config{
		DBName:          config.Schema,
		RefreshInterval: 15,
		//MetricsCollector: []prometheus.MetricsCollector{
		//	&prometheus.Postgres{
		//		VariableNames: []string{"Threads_running"},
		//	},
		//},
	}))

	db := &DB{
		DB: open,
	}

	return db, err
}
