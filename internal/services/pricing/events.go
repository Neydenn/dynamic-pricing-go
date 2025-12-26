package pricing

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

type PricePayload struct {
    ProductID    string  `json:"product_id"`
    CurrentPrice float64 `json:"current_price"`
}

func NewPriceEvent(p models.Price) ([]byte, error) {
    e := Event{
        Type: "price_updated",
        TS:   time.Now().UTC(),
        Payload: PricePayload{
            ProductID:    p.ProductID.String(),
            CurrentPrice: p.CurrentPrice,
        },
    }
    return json.Marshal(e)
}

