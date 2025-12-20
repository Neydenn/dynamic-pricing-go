package pricing

import (
    "context"
    "encoding/json"
    "errors"
    "log/slog"
    "math"
    "sync"
    "time"

    "dynamic-pricing/internal/models"
    "dynamic-pricing/internal/services"

    "github.com/google/uuid"
)

// PriceRepository abstracts price persistence.
type PriceRepository interface {
    UpsertPrice(ctx context.Context, productID uuid.UUID, currentPrice float64) (models.Price, error)
    GetPrice(ctx context.Context, productID uuid.UUID) (models.Price, error)
}

var ErrUnknownProduct = errors.New("unknown product")

// Engine encapsulates pricing logic and demand tracking.
type Engine struct {
    repo     PriceRepository
    bus      services.EventBus
    mu       sync.RWMutex
    products map[uuid.UUID]models.ProductSnapshot
    demandTS map[uuid.UUID][]time.Time
    window   time.Duration
}

func NewEngine(repo PriceRepository, bus services.EventBus) *Engine {
    return &Engine{
        repo:     repo,
        bus:      bus,
        products: make(map[uuid.UUID]models.ProductSnapshot),
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
    e.products[p.ID] = models.ProductSnapshot{
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

func (e *Engine) HandleOrderEvent(ctx context.Context, b []byte) (*models.Price, error) {
    var ev struct {
        Type    string          `json:"type"`
        TS      time.Time       `json:"ts"`
        Payload json.RawMessage `json:"payload"`
    }
    if err := json.Unmarshal(b, &ev); err != nil {
        return nil, err
    }
    var o struct {
        ProductID uuid.UUID `json:"product_id"`
        Qty       int       `json:"qty"`
    }
    if err := json.Unmarshal(ev.Payload, &o); err != nil {
        return nil, err
    }

    e.mu.Lock()
    snap, ok := e.products[o.ProductID]
    if !ok {
        e.mu.Unlock()
        slog.Warn("pricing: order for unknown product (no snapshot)", "product_id", o.ProductID)
        return nil, ErrUnknownProduct
    }
    ts := e.demandTS[o.ProductID]
    cutoff := time.Now().UTC().Add(-e.window)
    var kept []time.Time
    for _, t := range ts {
        if t.After(cutoff) {
            kept = append(kept, t)
        }
    }
    for i := 0; i < max(1, o.Qty); i++ {
        kept = append(kept, time.Now().UTC())
    }
    e.demandTS[o.ProductID] = kept
    e.mu.Unlock()

    demand := len(kept)
    price := computePrice(snap.BasePrice, snap.Stock, demand)
    stored, err := e.repo.UpsertPrice(ctx, o.ProductID, price)
    if err != nil {
        return nil, err
    }
    msg, err := NewPriceEvent(stored)
    if err != nil {
        return nil, err
    }
    if err := e.bus.Send(ctx, o.ProductID.String(), msg); err != nil {
        return nil, err
    }
    return &stored, nil
}

func (e *Engine) ComputeAndPersistCurrentPrice(ctx context.Context, productID uuid.UUID) (*models.Price, error) {
    e.mu.RLock()
    snap, ok := e.products[productID]
    ts := e.demandTS[productID]
    e.mu.RUnlock()
    if !ok {
        return nil, ErrUnknownProduct
    }
    demand := len(ts)
    price := computePrice(snap.BasePrice, snap.Stock, demand)
    stored, err := e.repo.UpsertPrice(ctx, productID, price)
    if err != nil {
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

