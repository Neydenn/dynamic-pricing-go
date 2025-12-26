package catalog

import (
    "dynamic-pricing/internal/models"
    "encoding/json"
    "time"
)

type Event struct {
    Type    string    `json:"type"`
    TS      time.Time `json:"ts"`
    Payload any       `json:"payload"`
}

type ProductPayload struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    BasePrice float64 `json:"base_price"`
    Stock     int     `json:"stock"`
}

func NewProductEvent(eventType string, p models.Product) ([]byte, error) {
    e := Event{
        Type: eventType,
        TS:   time.Now().UTC(),
        Payload: ProductPayload{
            ID:        p.ID.String(),
            Name:      p.Name,
            BasePrice: p.BasePrice,
            Stock:     p.Stock,
        },
    }
    return json.Marshal(e)
}

