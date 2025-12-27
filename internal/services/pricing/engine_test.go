package pricing

import (
    "context"
    "encoding/json"
    "testing"
    "time"

    "dynamic-pricing/internal/models"
    pmocks "dynamic-pricing/internal/services/pricing/mocks"
    smocks "dynamic-pricing/internal/services/mocks"

    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestHandleCatalogEvent_InitialPriceUpsert(t *testing.T) {
    repo := pmocks.NewPriceRepository(t)
    bus := smocks.NewEventBus(t)
    eng := NewEngine(repo, bus)

    pid := uuid.New()

    // Expect initial upsert with computed price 120.0 (base=100, stock=5)
    repo.EXPECT().
        UpsertPrice(mock.Anything, pid, 120.0).
        Return(models.Price{ProductID: pid, CurrentPrice: 120.0, UpdatedAt: time.Now().UTC()}, nil)

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
    require.NoError(t, eng.HandleCatalogEvent(mustJSON(t, ev)))
}

func TestHandleOrderEvent_UnknownProduct(t *testing.T) {
    repo := pmocks.NewPriceRepository(t)
    bus := smocks.NewEventBus(t)
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
    _, err := eng.HandleOrderEvent(context.Background(), mustJSON(t, orderEv))
    require.Error(t, err)
    require.Equal(t, ErrUnknownProduct, err)

    // Ensure no side-effects
    repo.AssertNotCalled(t, "UpsertPrice", mock.Anything, mock.Anything, mock.Anything)
    bus.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestHandleOrderEvent_PriceUpdatedAndEventSent(t *testing.T) {
    repo := pmocks.NewPriceRepository(t)
    bus := smocks.NewEventBus(t)
    eng := NewEngine(repo, bus)

    pid := uuid.New()

    // Catalog snapshot leads to initial price 100.0 (base=100, stock=10)
    repo.EXPECT().
        UpsertPrice(mock.Anything, pid, 100.0).
        Return(models.Price{ProductID: pid, CurrentPrice: 100.0, UpdatedAt: time.Now().UTC()}, nil)

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
    require.NoError(t, eng.HandleCatalogEvent(mustJSON(t, cat)))

    // Order triggers new price 102.0 and an event
    repo.EXPECT().
        UpsertPrice(mock.Anything, pid, 102.0).
        Return(models.Price{ProductID: pid, CurrentPrice: 102.0, UpdatedAt: time.Now().UTC()}, nil)
    bus.EXPECT().
        Send(mock.Anything, pid.String(), mock.Anything).
        Return(nil)

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
    require.NoError(t, err)
    require.NotNil(t, p)
    require.InDelta(t, 102.0, p.CurrentPrice, 0.0001)
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
