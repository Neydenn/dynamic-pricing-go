package order

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

func (r *Repository) CreateUser(ctx context.Context, email string) (User, error) {
	u := User{ID: uuid.New(), Email: email, CreatedAt: time.Now().UTC()}
	_, err := r.db.Exec(ctx, `insert into users(id, email, created_at) values($1,$2,$3)`, u.ID, u.Email, u.CreatedAt)
	return u, err
}

func (r *Repository) CreateOrder(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (Order, error) {
	o := Order{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Qty:       qty,
		Status:    "placed",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_, err := r.db.Exec(ctx, `insert into orders(id, user_id, product_id, qty, status, created_at, updated_at) values($1,$2,$3,$4,$5,$6,$7)`, o.ID, o.UserID, o.ProductID, o.Qty, o.Status, o.CreatedAt, o.UpdatedAt)
	return o, err
}

func (r *Repository) CancelOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	var o Order
	o.UpdatedAt = time.Now().UTC()
	row := r.db.QueryRow(ctx, `update orders set status='canceled', updated_at=$2 where id=$1 returning id, user_id, product_id, qty, status, created_at, updated_at`, id, o.UpdatedAt)
	err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Qty, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *Repository) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	var o Order
	row := r.db.QueryRow(ctx, `select id, user_id, product_id, qty, status, created_at, updated_at from orders where id=$1`, id)
	err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Qty, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}
