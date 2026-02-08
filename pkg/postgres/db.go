package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

type ZapTracer struct {
	logger *zap.Logger
}

func (zt *ZapTracer) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		zt.logger.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		zt.logger.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		zt.logger.Warn(msg, fields...)
	case tracelog.LogLevelError:
		zt.logger.Error(msg, fields...)
	default:
		zt.logger.Info(msg, fields...)
	}
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string, log *zap.Logger) (*Postgres, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: invalid config: %w", err)
	}

	if log != nil {
		poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   &ZapTracer{logger: log},
			LogLevel: tracelog.LogLevelInfo,
		}
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping failed: %w", err)
	}

	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
