package order

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

type mockOrderRepo struct {
	createUser []string
	createOrd  []struct {
		user    uuid.UUID
		product uuid.UUID
		qty     int
	}
	cancelOrd []uuid.UUID
	getOrd    []uuid.UUID
	err       error
}

func (m *mockOrderRepo) CreateUser(ctx context.Context, email string) (models.User, error) {
	if m.err != nil {
		return models.User{}, m.err
	}
	m.createUser = append(m.createUser, email)
	return models.User{ID: uuid.New(), Email: email, CreatedAt: time.Now().UTC()}, nil
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (models.Order, error) {
	if m.err != nil {
		return models.Order{}, m.err
	}
	m.createOrd = append(m.createOrd, struct {
		user    uuid.UUID
		product uuid.UUID
		qty     int
	}{userID, productID, qty})
	return models.Order{ID: uuid.New(), UserID: userID, ProductID: productID, Qty: qty, Status: "placed", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil
}

func (m *mockOrderRepo) CancelOrder(ctx context.Context, id uuid.UUID) (models.Order, error) {
	if m.err != nil {
		return models.Order{}, m.err
	}
	m.cancelOrd = append(m.cancelOrd, id)
	return models.Order{ID: id, Status: "canceled", UpdatedAt: time.Now().UTC()}, nil
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (models.Order, error) {
	m.getOrd = append(m.getOrd, id)
	if m.err != nil {
		return models.Order{}, m.err
	}
	return models.Order{ID: id}, nil
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

func TestPlaceOrder_SendsEvent(t *testing.T) {
	repo := &mockOrderRepo{}
	bus := &mockBus{}
	svc := NewService(repo, bus)

	uid := uuid.New()
	pid := uuid.New()
	o, err := svc.PlaceOrder(context.Background(), uid, pid, 2)
	if err != nil {
		t.Fatalf("PlaceOrder err: %v", err)
	}
	if o.UserID != uid || o.ProductID != pid || o.Qty != 2 {
		t.Fatalf("order mismatch: %+v", o)
	}
	if len(bus.sends) != 1 {
		t.Fatalf("expected 1 send")
	}
	if got := decodeType(t, bus.sends[0].value); got != "order_placed" {
		t.Fatalf("type=%s", got)
	}
}

func TestCancelOrder_SendsEvent(t *testing.T) {
	repo := &mockOrderRepo{}
	bus := &mockBus{}
	svc := NewService(repo, bus)

	id := uuid.New()
	o, err := svc.CancelOrder(context.Background(), id)
	if err != nil {
		t.Fatalf("CancelOrder err: %v", err)
	}
	if o.Status != "canceled" {
		t.Fatalf("status=%s", o.Status)
	}
	if len(bus.sends) != 1 {
		t.Fatalf("expected 1 send")
	}
	if got := decodeType(t, bus.sends[0].value); got != "order_canceled" {
		t.Fatalf("type=%s", got)
	}
}

func TestService_RepoError_NoEvent(t *testing.T) {
	repo := &mockOrderRepo{err: errors.New("boom")}
	bus := &mockBus{}
	svc := NewService(repo, bus)
	if _, err := svc.PlaceOrder(context.Background(), uuid.New(), uuid.New(), 1); err == nil {
		t.Fatalf("want error")
	}
	if len(bus.sends) != 0 {
		t.Fatalf("no events expected")
	}
}
