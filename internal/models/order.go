package models

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID        uuid.UUID `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type Order struct {
    ID        uuid.UUID `json:"id"`
    UserID    uuid.UUID `json:"user_id"`
    ProductID uuid.UUID `json:"product_id"`
    Qty       int       `json:"qty"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

