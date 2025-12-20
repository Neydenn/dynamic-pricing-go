package catalog

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

func (s *Service) Create(ctx context.Context, name string, basePrice float64, stock int) (Product, error) {
	p := Product{
		ID:        uuid.New(),
		Name:      name,
		BasePrice: basePrice,
		Stock:     stock,
	}
	p, err := s.repo.Create(ctx, p)
	if err != nil {
		return p, err
	}
	b, err := NewProductEvent("product_created", p)
	if err != nil {
		return p, err
	}
	if err := s.prod.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, name string, basePrice float64) (Product, error) {
	p, err := s.repo.Update(ctx, id, name, basePrice)
	if err != nil {
		return p, err
	}
	b, err := NewProductEvent("product_updated", p)
	if err != nil {
		return p, err
	}
	if err := s.prod.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) UpdateStock(ctx context.Context, id uuid.UUID, stock int) (Product, error) {
	p, err := s.repo.UpdateStock(ctx, id, stock)
	if err != nil {
		return p, err
	}
	b, err := NewProductEvent("stock_changed", p)
	if err != nil {
		return p, err
	}
	if err := s.prod.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Product, error) {
	return s.repo.Get(ctx, id)
}
