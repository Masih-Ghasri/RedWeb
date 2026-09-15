package services

import (
	"context"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/ports"
	"github.com/google/uuid"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, cmd ports.CreateProductCommand) (*domain.Product, error) {
	id := uuid.New().String()

	product, err := domain.NewProduct(id, cmd.Name, cmd.Description, cmd.Price, cmd.Stock)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}
