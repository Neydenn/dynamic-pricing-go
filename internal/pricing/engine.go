package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"sync"
	"time"

	"dynamic-pricing/internal/kafka"

	"github.com/google/uuid"
)

type Engine struct {
	repo     *Repository
	prod     *kafka.Producer
	mu       sync.RWMutex
	products map[uuid.UUID]ProductSnapshot
	demandTS map[uuid.UUID][]time.Time
	window   time.Duration
}

func NewEngine(repo *Repository, prod *kafka.Producer) *Engine {
	return &Engine{
		repo:     repo,
		prod:     prod,
		products: make(map[uuid.UUID]ProductSnapshot),
		demandTS: make(map[uuid.UUID][]time.Time),
		window:   2 * time.Minute,
	}
}

func (e *Engine) HandleCatalogEvent(b []byte) error {
	var ev struct {
		Type    string          `json:"type"`
		TS      time.Time       `json:"ts"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(b, &ev); err != nil {
		return err
	}
	var p struct {
		ID        uuid.UUID `json:"id"`
		BasePrice float64   `json:"base_price"`
		Stock     int       `json:"stock"`
	}
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return err
	}
	e.mu.Lock()
	e.products[p.ID] = ProductSnapshot{
		ID:        p.ID,
		BasePrice: p.BasePrice,
		Stock:     p.Stock,
		UpdatedAt: ev.TS,
	}

	snap := e.products[p.ID]
	e.mu.Unlock()
	slog.Info("pricing: catalog snapshot", "product_id", p.ID, "base_price", p.BasePrice, "stock", p.Stock)

	price := computePrice(snap.BasePrice, snap.Stock, 0)
	stored, err := e.repo.UpsertPrice(context.Background(), p.ID, price)
	if err == nil {
		slog.Info("pricing: initial price", "product_id", stored.ProductID, "price", stored.CurrentPrice)
	}
	return err
}

var ErrUnknownProduct = errors.New("unknown product")

func (e *Engine) ComputeAndPersistCurrentPrice(ctx context.Context, productID uuid.UUID) (Price, error) {
	e.mu.RLock()
	snap, ok := e.products[productID]
	ts := e.demandTS[productID]
	e.mu.RUnlock()
	if !ok {
		return Price{}, ErrUnknownProduct
	}
	demand := len(ts)
	price := computePrice(snap.BasePrice, snap.Stock, demand)
	return e.repo.UpsertPrice(ctx, productID, price)
}

func (e *Engine) HandleOrderEvent(ctx context.Context, b []byte) (*Price, error) {
	var ev struct {
		Type    string          `json:"type"`
		TS      time.Time       `json:"ts"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(b, &ev); err != nil {
		return nil, err
	}
	if ev.Type != "order_placed" && ev.Type != "order_canceled" {
		return nil, nil
	}
	var o struct {
		ProductID uuid.UUID `json:"product_id"`
		Qty       int       `json:"qty"`
		Status    string    `json:"status"`
	}
	if err := json.Unmarshal(ev.Payload, &o); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	e.mu.Lock()
	if ev.Type == "order_placed" {
		for i := 0; i < max(1, o.Qty); i++ {
			e.demandTS[o.ProductID] = append(e.demandTS[o.ProductID], now)
		}
	}
	ts := e.demandTS[o.ProductID]
	cut := now.Add(-e.window)
	j := 0
	for _, t := range ts {
		if t.After(cut) {
			ts[j] = t
			j++
		}
	}
	ts = ts[:j]
	e.demandTS[o.ProductID] = ts
	snap, ok := e.products[o.ProductID]
	e.mu.Unlock()

	if !ok {
		slog.Warn("pricing: order for unknown product (no snapshot)", "product_id", o.ProductID)
		return nil, nil
	}

	demand := len(ts)
	price := computePrice(snap.BasePrice, snap.Stock, demand)
	stored, err := e.repo.UpsertPrice(ctx, o.ProductID, price)
	if err != nil {
		return nil, err
	}
	msg, err := NewPriceEvent(stored)
	if err != nil {
		return nil, err
	}
	if err := e.prod.Send(ctx, o.ProductID.String(), msg); err != nil {
		return nil, err
	}
	return &stored, nil
}

func computePrice(base float64, stock int, demand int) float64 {
	m := 1.0
	m += math.Min(0.30, float64(demand)*0.02)
	if stock <= 5 {
		m += 0.20
	}
	if stock <= 0 {
		m += 0.50
	}
	v := base * m
	return math.Round(v*100) / 100
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
