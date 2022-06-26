package logger

import (
	"context"
	"gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	logLevel logger.LogLevel
}

func (l *GormLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	l.logLevel = logLevel
	return l
}

func (l *GormLogger) Info(ctx context.Context, message string, values ...interface{}) {
	Info(ctx, message, values)
}
func (l *GormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	Warn(ctx, message, values)
}
func (l *GormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	Error(ctx, message, values)
}
func (l *GormLogger) Trace(ctx context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), _ error) {
	sql, rows := fc()

	Trace(ctx, "ROWS: %v, SQL: %v", rows, sql)
}
