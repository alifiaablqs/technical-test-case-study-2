package service

import (
	"context"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]*model.Product, error)
	GetProductByID(ctx context.Context, id uint64) (*model.Product, error)
	CreateProduct(ctx context.Context, req *model.ProductRequest) (*model.Product, error)
	UpdateProduct(ctx context.Context, id uint64, req *model.ProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id uint64) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetAllProducts(ctx context.Context) ([]*model.Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *productService) GetProductByID(ctx context.Context, id uint64) (*model.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *productService) CreateProduct(ctx context.Context, req *model.ProductRequest) (*model.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	onHand := float64(0)
	if req.OnHand != nil {
		onHand = *req.OnHand
	}

	reserved := float64(0)
	if req.Reserved != nil {
		reserved = *req.Reserved
	}

	prod := &model.Product{
		SKU:      req.SKU,
		Name:     req.Name,
		Unit:     req.Unit,
		OnHand:   onHand,
		Reserved: reserved,
	}

	if err := s.repo.Create(ctx, prod); err != nil {
		return nil, err
	}

	prod.CalculateAvailable()
	return prod, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uint64, req *model.ProductRequest) (*model.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.SKU = req.SKU
	existing.Name = req.Name
	existing.Unit = req.Unit
	if req.OnHand != nil {
		existing.OnHand = *req.OnHand
	}
	if req.Reserved != nil {
		existing.Reserved = *req.Reserved
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	existing.CalculateAvailable()
	return existing, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
