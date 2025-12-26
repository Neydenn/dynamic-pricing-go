package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"dynamic-pricing/internal/models"
	"dynamic-pricing/internal/services"

	"github.com/google/uuid"
)

type mockProductRepo struct {
	created []models.Product
	updates []struct {
		id   uuid.UUID
		name string
		base float64
	}
	stocks []struct {
		id    uuid.UUID
		stock int
	}
	getCalls []uuid.UUID
	err      error
}

func (m *mockProductRepo) Create(ctx context.Context, p models.Product) (models.Product, error) {
	if m.err != nil {
		return models.Product{}, m.err
	}
	m.created = append(m.created, p)
	p.UpdatedAt = time.Now().UTC()
	return p, nil
}
func (m *mockProductRepo) Update(ctx context.Context, id uuid.UUID, name string, basePrice float64) (models.Product, error) {
	if m.err != nil {
		return models.Product{}, m.err
	}
	m.updates = append(m.updates, struct {
		id   uuid.UUID
		name string
		base float64
	}{id, name, basePrice})
	return models.Product{ID: id, Name: name, BasePrice: basePrice, UpdatedAt: time.Now().UTC()}, nil
}
func (m *mockProductRepo) UpdateStock(ctx context.Context, id uuid.UUID, stock int) (models.Product, error) {
	if m.err != nil {
		return models.Product{}, m.err
	}
	m.stocks = append(m.stocks, struct {
		id    uuid.UUID
		stock int
	}{id, stock})
	return models.Product{ID: id, Stock: stock, UpdatedAt: time.Now().UTC()}, nil
}
func (m *mockProductRepo) Get(ctx context.Context, id uuid.UUID) (models.Product, error) {
	m.getCalls = append(m.getCalls, id)
	if m.err != nil {
		return models.Product{}, m.err
	}
	return models.Product{ID: id}, nil
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

func decodeType(t *testing.T, b []byte) string {
	t.Helper()
	var ev struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return ev.Type
}

func TestService_Create_SendsEvent(t *testing.T) {
	repo := &mockProductRepo{}
	bus := &mockBus{}
	svc := NewService(repo, bus)

	p, err := svc.Create(context.Background(), "A", 10.5, 5)
	if err != nil {
		t.Fatalf("Create err: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("repo.Create not called")
	}
	if len(bus.sends) != 1 {
		t.Fatalf("bus.Send not called")
	}
	if bus.sends[0].key != p.ID.String() {
		t.Fatalf("key mismatch: %s", bus.sends[0].key)
	}
	if got := decodeType(t, bus.sends[0].value); got != "product_created" {
		t.Fatalf("event type=%s", got)
	}
}

func TestService_Update_SendsEvent(t *testing.T) {
	repo := &mockProductRepo{}
	bus := &mockBus{}
	svc := NewService(repo, bus)
	id := uuid.New()
	p, err := svc.Update(context.Background(), id, "B", 20)
	if err != nil {
		t.Fatalf("Update err: %v", err)
	}
	if p.ID != id {
		t.Fatalf("id mismatch")
	}
	if len(bus.sends) != 1 {
		t.Fatalf("expected 1 send")
	}
	if got := decodeType(t, bus.sends[0].value); got != "product_updated" {
		t.Fatalf("type=%s", got)
	}
}

func TestService_UpdateStock_SendsEvent(t *testing.T) {
	repo := &mockProductRepo{}
	bus := &mockBus{}
	svc := NewService(repo, bus)
	id := uuid.New()
	_, err := svc.UpdateStock(context.Background(), id, 7)
	if err != nil {
		t.Fatalf("UpdateStock err: %v", err)
	}
	if len(bus.sends) != 1 {
		t.Fatalf("expected 1 send")
	}
	if got := decodeType(t, bus.sends[0].value); got != "product_stock_updated" {
		t.Fatalf("type=%s", got)
	}
}

func TestService_RepoError_Propagates_NoEvent(t *testing.T) {
	repo := &mockProductRepo{err: errors.New("boom")}
	bus := &mockBus{}
	svc := NewService(repo, bus)
	if _, err := svc.Create(context.Background(), "A", 1, 1); err == nil {
		t.Fatalf("want error")
	}
	if len(bus.sends) != 0 {
		t.Fatalf("no events expected")
	}
}
