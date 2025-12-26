package order

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

type OrderPayload struct {
    ID        string `json:"id"`
    UserID    string `json:"user_id"`
    ProductID string `json:"product_id"`
    Qty       int    `json:"qty"`
    Status    string `json:"status"`
}

func NewOrderEvent(eventType string, o models.Order) ([]byte, error) {
    e := Event{
        Type: eventType,
        TS:   time.Now().UTC(),
        Payload: OrderPayload{
            ID:        o.ID.String(),
            UserID:    o.UserID.String(),
            ProductID: o.ProductID.String(),
            Qty:       o.Qty,
            Status:    o.Status,
        },
    }
    return json.Marshal(e)
}

