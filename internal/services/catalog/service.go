package catalog

import (
	"context"

	"dynamic-pricing/internal/models"
	"dynamic-pricing/internal/services"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, p models.Product) (models.Product, error)
	Update(ctx context.Context, id uuid.UUID, name string, basePrice float64) (models.Product, error)
	UpdateStock(ctx context.Context, id uuid.UUID, stock int) (models.Product, error)
	Get(ctx context.Context, id uuid.UUID) (models.Product, error)
}

type Service struct {
	repo ProductRepository
	bus  services.EventBus
}

func NewService(repo ProductRepository, bus services.EventBus) *Service {
	return &Service{repo: repo, bus: bus}
}

func (s *Service) Create(ctx context.Context, name string, basePrice float64, stock int) (models.Product, error) {
	p := models.Product{
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
	if err := s.bus.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, name string, basePrice float64) (models.Product, error) {
	p, err := s.repo.Update(ctx, id, name, basePrice)
	if err != nil {
		return p, err
	}
	b, err := NewProductEvent("product_updated", p)
	if err != nil {
		return p, err
	}
	if err := s.bus.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) UpdateStock(ctx context.Context, id uuid.UUID, stock int) (models.Product, error) {
	p, err := s.repo.UpdateStock(ctx, id, stock)
	if err != nil {
		return p, err
	}
	b, err := NewProductEvent("product_stock_updated", p)
	if err != nil {
		return p, err
	}
	if err := s.bus.Send(ctx, p.ID.String(), b); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (models.Product, error) {
	return s.repo.Get(ctx, id)
}
