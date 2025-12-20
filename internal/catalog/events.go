package catalog

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

type ProductPayload struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	BasePrice float64   `json:"base_price"`
	Stock     int       `json:"stock"`
}

func NewProductEvent(eventType string, p Product) ([]byte, error) {
	e := Event{
		Type: eventType,
		TS:   time.Now().UTC(),
		Payload: ProductPayload{
			ID:        p.ID,
			Name:      p.Name,
			BasePrice: p.BasePrice,
			Stock:     p.Stock,
		},
	}
	return json.Marshal(e)
}
