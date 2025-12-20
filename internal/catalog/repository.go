package catalog

import (
	"context"
	"errors"
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

func (r *Repository) Create(ctx context.Context, p Product) (Product, error) {
	p.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx, `insert into products(id, name, base_price, stock, updated_at) values($1,$2,$3,$4,$5)`, p.ID, p.Name, p.BasePrice, p.Stock, p.UpdatedAt)
	return p, err
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, name string, basePrice float64) (Product, error) {
	var p Product
	p.UpdatedAt = time.Now().UTC()
	row := r.db.QueryRow(ctx, `update products set name=$2, base_price=$3, updated_at=$4 where id=$1 returning id, name, base_price, stock, updated_at`, id, name, basePrice, p.UpdatedAt)
	if err := row.Scan(&p.ID, &p.Name, &p.BasePrice, &p.Stock, &p.UpdatedAt); err != nil {
		return p, err
	}
	return p, nil
}

func (r *Repository) UpdateStock(ctx context.Context, id uuid.UUID, stock int) (Product, error) {
	var p Product
	p.UpdatedAt = time.Now().UTC()
	row := r.db.QueryRow(ctx, `update products set stock=$2, updated_at=$3 where id=$1 returning id, name, base_price, stock, updated_at`, id, stock, p.UpdatedAt)
	if err := row.Scan(&p.ID, &p.Name, &p.BasePrice, &p.Stock, &p.UpdatedAt); err != nil {
		return p, err
	}
	return p, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Product, error) {
	var p Product
	row := r.db.QueryRow(ctx, `select id, name, base_price, stock, updated_at from products where id=$1`, id)
	if err := row.Scan(&p.ID, &p.Name, &p.BasePrice, &p.Stock, &p.UpdatedAt); err != nil {
		return p, err
	}
	if p.ID == uuid.Nil {
		return p, errors.New("not found")
	}
	return p, nil
}
