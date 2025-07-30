package dblogger

import (
	"context"
	"time"

	"github.com/go-solana-parse/src/seelog"
	"gorm.io/gorm/logger"
)

type AppGormLogger struct {
}

func (l *AppGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *AppGormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	seelog.AppInfof(s, i)
}

func (l *AppGormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	seelog.AppWarnf(s, i)
}

func (l *AppGormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	seelog.AppErrorf(s, i)

}

func (l *AppGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	data, derr := fc()
	if len(data) > 10240 {
		data = data[0:10240]
	}
	if err != nil {
		seelog.AppErrorf("sql:%s - result: %v - rowsAffected: %d", data, err, derr)
	} else {
		seelog.AppInfof("sql:%s - result: %s - rowsAffected: %d", data, "OK", derr)
	}
}
