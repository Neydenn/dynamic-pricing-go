package order

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Type    string    `json:"type"`
	TS      time.Time `json:"ts"`
	Payload any       `json:"payload"`
}

type OrderPayload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	Qty       int       `json:"qty"`
	Status    string    `json:"status"`
}

func NewOrderEvent(eventType string, o Order) ([]byte, error) {
	e := Event{
		Type: eventType,
		TS:   time.Now().UTC(),
		Payload: OrderPayload{
			ID:        o.ID,
			UserID:    o.UserID,
			ProductID: o.ProductID,
			Qty:       o.Qty,
			Status:    o.Status,
		},
	}
	return json.Marshal(e)
}
