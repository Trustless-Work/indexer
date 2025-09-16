package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

func ConfigFromEnv() Config {
	return Config{
		DSN:             getenv("DB_DSN", "postgres://indexer:indexer@localhost:5432/indexer?sslmode=disable"),
		MaxConns:        int32(getenvInt("DB_MAX_CONNS", 10)),
		MinConns:        int32(getenvInt("DB_MIN_CONNS", 1)),
		MaxConnLifetime: getenvDur("DB_MAX_LIFETIME", 30*time.Minute),
		MaxConnIdleTime: getenvDur("DB_MAX_IDLE_TIME", 5*time.Minute),
		ConnectTimeout:  getenvDur("DB_CONNECT_TIMEOUT", 5*time.Second),
	}
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pconf, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}
	pconf.MaxConns = cfg.MaxConns
	pconf.MinConns = cfg.MinConns
	pconf.MaxConnLifetime = cfg.MaxConnLifetime
	pconf.MaxConnIdleTime = cfg.MaxConnIdleTime
	pconf.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	return pgxpool.NewWithConfig(ctx, pconf)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var x int
		_, _ = fmt.Sscanf(v, "%d", &x)
		if x > 0 {
			return x
		}
	}
	return def
}
func getenvDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
