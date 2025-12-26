package order

import (
	"context"

	"dynamic-pricing/internal/models"
	"dynamic-pricing/internal/services"

	"github.com/google/uuid"
)

type OrderRepository interface {
	CreateUser(ctx context.Context, email string) (models.User, error)
	CreateOrder(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (models.Order, error)
	CancelOrder(ctx context.Context, id uuid.UUID) (models.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (models.Order, error)
}

type Service struct {
	repo OrderRepository
	bus  services.EventBus
}

func NewService(repo OrderRepository, bus services.EventBus) *Service {
	return &Service{repo: repo, bus: bus}
}

func (s *Service) CreateUser(ctx context.Context, email string) (models.User, error) {
	return s.repo.CreateUser(ctx, email)
}

func (s *Service) PlaceOrder(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (models.Order, error) {
	o, err := s.repo.CreateOrder(ctx, userID, productID, qty)
	if err != nil {
		return o, err
	}
	b, err := NewOrderEvent("order_placed", o)
	if err != nil {
		return o, err
	}
	if err := s.bus.Send(ctx, o.ID.String(), b); err != nil {
		return o, err
	}
	return o, nil
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID) (models.Order, error) {
	o, err := s.repo.CancelOrder(ctx, id)
	if err != nil {
		return o, err
	}
	b, err := NewOrderEvent("order_canceled", o)
	if err != nil {
		return o, err
	}
	if err := s.bus.Send(ctx, o.ID.String(), b); err != nil {
		return o, err
	}
	return o, nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (models.Order, error) {
	return s.repo.GetOrder(ctx, id)
}
