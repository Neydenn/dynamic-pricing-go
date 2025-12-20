package pricing

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

type PricePayload struct {
	ProductID    uuid.UUID `json:"product_id"`
	CurrentPrice float64   `json:"current_price"`
}

func NewPriceEvent(p Price) ([]byte, error) {
	e := Event{
		Type: "price_updated",
		TS:   time.Now().UTC(),
		Payload: PricePayload{
			ProductID:    p.ProductID,
			CurrentPrice: p.CurrentPrice,
		},
	}
	return json.Marshal(e)
}
