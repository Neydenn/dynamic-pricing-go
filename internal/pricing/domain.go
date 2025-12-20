package pricing

import (
	"time"

	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID        uuid.UUID
	BasePrice float64
	Stock     int
	UpdatedAt time.Time
}

type Price struct {
	ProductID    uuid.UUID `json:"product_id"`
	CurrentPrice float64   `json:"current_price"`
	UpdatedAt    time.Time `json:"updated_at"`
}
