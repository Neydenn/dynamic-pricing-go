package catalog

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	BasePrice float64   `json:"base_price"`
	Stock     int       `json:"stock"`
	UpdatedAt time.Time `json:"updated_at"`
}
