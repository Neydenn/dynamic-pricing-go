package pricing

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dynamic-pricing/internal/models"
	"dynamic-pricing/internal/services"

	"github.com/google/uuid"
)

type mockPriceRepo struct {
	upserts []models.Price
	getMap  map[uuid.UUID]models.Price
	upErr   error
	getErr  error
}

func (m *mockPriceRepo) UpsertPrice(ctx context.Context, productID uuid.UUID, currentPrice float64) (models.Price, error) {
	if m.upErr != nil {
		return models.Price{}, m.upErr
	}
	p := models.Price{ProductID: productID, CurrentPrice: currentPrice, UpdatedAt: time.Now().UTC()}
	m.upserts = append(m.upserts, p)
	return p, nil
}

func (m *mockPriceRepo) GetPrice(ctx context.Context, productID uuid.UUID) (models.Price, error) {
	if m.getErr != nil {
		return models.Price{}, m.getErr
	}
	if m.getMap == nil {
		return models.Price{}, m.getErr
	}
	return m.getMap[productID], nil
}

type mockBus struct {
	sends []struct {
		key   string
		value []byte
	}
	err error
}

func (m *mockBus) Send(ctx context.Context, key string, value []byte) error {
	if m.err != nil {
		return m.err
	}
	m.sends = append(m.sends, struct {
		key   string
		value []byte
	}{key, value})
	return nil
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestHandleCatalogEvent_InitialPriceUpsert(t *testing.T) {
	repo := &mockPriceRepo{}
	bus := &mockBus{}
	eng := NewEngine(repo, bus)

	pid := uuid.New()
	ev := struct {
		Type    string    `json:"type"`
		TS      time.Time `json:"ts"`
		Payload any       `json:"payload"`
	}{
		Type: "product_created",
		TS:   time.Now().UTC(),
		Payload: map[string]any{
			"id":         pid,
			"name":       "A",
			"base_price": 100.0,
			"stock":      5,
		},
	}
	if err := eng.HandleCatalogEvent(mustJSON(t, ev)); err != nil {
		t.Fatalf("HandleCatalogEvent err: %v", err)
	}
	if len(repo.upserts) != 1 {
		t.Fatalf("expected 1 upsert, got %d", len(repo.upserts))
	}
	got := repo.upserts[0].CurrentPrice
	if got != 120.0 {
		t.Fatalf("unexpected price: got %v want 120.0", got)
	}
}

func TestHandleOrderEvent_UnknownProduct(t *testing.T) {
	repo := &mockPriceRepo{}
	bus := &mockBus{}
	eng := NewEngine(repo, bus)

	pid := uuid.New()
	orderEv := struct {
		Type    string    `json:"type"`
		TS      time.Time `json:"ts"`
		Payload any       `json:"payload"`
	}{
		Type: "order_placed",
		TS:   time.Now().UTC(),
		Payload: map[string]any{
			"product_id": pid,
			"qty":        1,
		},
	}
	if _, err := eng.HandleOrderEvent(context.Background(), mustJSON(t, orderEv)); err == nil || err != ErrUnknownProduct {
		t.Fatalf("expected ErrUnknownProduct, got %v", err)
	}
	if len(repo.upserts) != 0 {
		t.Fatalf("unexpected upserts: %d", len(repo.upserts))
	}
	if len(bus.sends) != 0 {
		t.Fatalf("unexpected bus sends: %d", len(bus.sends))
	}
}

func TestHandleOrderEvent_PriceUpdatedAndEventSent(t *testing.T) {
	repo := &mockPriceRepo{}
	bus := &mockBus{}
	eng := NewEngine(repo, bus)

	pid := uuid.New()
	cat := struct {
		Type    string    `json:"type"`
		TS      time.Time `json:"ts"`
		Payload any       `json:"payload"`
	}{
		Type: "product_created",
		TS:   time.Now().UTC(),
		Payload: map[string]any{
			"id":         pid,
			"name":       "X",
			"base_price": 100.0,
			"stock":      10,
		},
	}
	if err := eng.HandleCatalogEvent(mustJSON(t, cat)); err != nil {
		t.Fatalf("seed catalog: %v", err)
	}

	ord := struct {
		Type    string    `json:"type"`
		TS      time.Time `json:"ts"`
		Payload any       `json:"payload"`
	}{
		Type: "order_placed",
		TS:   time.Now().UTC(),
		Payload: map[string]any{
			"product_id": pid,
			"qty":        1,
		},
	}
	p, err := eng.HandleOrderEvent(context.Background(), mustJSON(t, ord))
	if err != nil {
		t.Fatalf("HandleOrderEvent err: %v", err)
	}
	if p == nil || p.CurrentPrice != 102.0 {
		t.Fatalf("unexpected price: %+v", p)
	}
	if len(repo.upserts) == 0 {
		t.Fatalf("expected upserts")
	}
	last := repo.upserts[len(repo.upserts)-1]
	if last.CurrentPrice != 102.0 {
		t.Fatalf("unexpected last upsert: %v", last.CurrentPrice)
	}
	if len(bus.sends) != 1 {
		t.Fatalf("expected 1 event, got %d", len(bus.sends))
	}
	if bus.sends[0].key != pid.String() {
		t.Fatalf("event key mismatch: %s", bus.sends[0].key)
	}
}

func TestComputePrice_Table(t *testing.T) {
	cases := []struct {
		name   string
		base   float64
		stock  int
		demand int
		want   float64
	}{
		{"no_modifiers", 100, 10, 0, 100.0},
		{"low_stock", 100, 5, 0, 120.0},
		{"out_of_stock", 100, 0, 0, 170.0},
		{"demand_cap", 100, 10, 100, 130.0},
		{"combined", 100, 3, 10, 140.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := computePrice(tc.base, tc.stock, tc.demand); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
