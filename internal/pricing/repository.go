package pricing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertPrice(ctx context.Context, productID uuid.UUID, currentPrice float64) (Price, error) {
	p := Price{
		ProductID:    productID,
		CurrentPrice: currentPrice,
		UpdatedAt:    time.Now().UTC(),
	}
	_, err := r.db.Exec(ctx, `insert into prices(product_id, current_price, updated_at) values($1,$2,$3)
		on conflict (product_id) do update set current_price=excluded.current_price, updated_at=excluded.updated_at`, p.ProductID, p.CurrentPrice, p.UpdatedAt)
	return p, err
}

func (r *Repository) GetPrice(ctx context.Context, productID uuid.UUID) (Price, error) {
	var p Price
	row := r.db.QueryRow(ctx, `select product_id, current_price, updated_at from prices where product_id=$1`, productID)
	err := row.Scan(&p.ProductID, &p.CurrentPrice, &p.UpdatedAt)
	return p, err
}
