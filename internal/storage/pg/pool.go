package pg

import (
    "context"
    "fmt"

    "dynamic-pricing/config"

    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, c config.Postgres) (*pgxpool.Pool, error) {
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
    return pgxpool.New(ctx, dsn)
}

