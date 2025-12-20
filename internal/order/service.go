package order

import (
	"context"

	"dynamic-pricing/internal/kafka"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
	prod *kafka.Producer
}

func NewService(repo *Repository, prod *kafka.Producer) *Service {
	return &Service{repo: repo, prod: prod}
}

func (s *Service) CreateUser(ctx context.Context, email string) (User, error) {
	return s.repo.CreateUser(ctx, email)
}

func (s *Service) PlaceOrder(ctx context.Context, userID uuid.UUID, productID uuid.UUID, qty int) (Order, error) {
	o, err := s.repo.CreateOrder(ctx, userID, productID, qty)
	if err != nil {
		return o, err
	}
	b, err := NewOrderEvent("order_placed", o)
	if err != nil {
		return o, err
	}
	if err := s.prod.Send(ctx, o.ID.String(), b); err != nil {
		return o, err
	}
	return o, nil
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	o, err := s.repo.CancelOrder(ctx, id)
	if err != nil {
		return o, err
	}
	b, err := NewOrderEvent("order_canceled", o)
	if err != nil {
		return o, err
	}
	if err := s.prod.Send(ctx, o.ID.String(), b); err != nil {
		return o, err
	}
	return o, nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	return s.repo.GetOrder(ctx, id)
}
