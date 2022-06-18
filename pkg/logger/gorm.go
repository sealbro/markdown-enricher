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

func (l *GormLogger) Info(_ context.Context, message string, values ...interface{}) {
	Infof(message, values)
}
func (l *GormLogger) Warn(_ context.Context, message string, values ...interface{}) {
	Warnf(message, values)
}
func (l *GormLogger) Error(_ context.Context, message string, values ...interface{}) {
	Errorf(message, values)
}
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()

	Tracef("ROWS: %v, SQL: %v", rows, sql)
}
