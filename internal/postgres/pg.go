package postgres

import (
	"context"
	"fmt"
	"time"

	"dynamic-pricing/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, c config.Postgres) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Minute * 30
	return pgxpool.NewWithConfig(ctx, cfg)
}
