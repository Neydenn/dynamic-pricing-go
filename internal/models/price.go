package models

import (
    "time"

    "github.com/google/uuid"
)

type Price struct {
    ProductID    uuid.UUID `json:"product_id"`
    CurrentPrice float64   `json:"current_price"`
    UpdatedAt    time.Time `json:"updated_at"`
}

