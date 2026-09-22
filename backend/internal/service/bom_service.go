package service

import (
	"context"
	"errors"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrMaterialNotFound = errors.New("material not found")
)

type BOMService interface {
	CreateBOM(ctx context.Context, req *model.CreateBOMRequest) (*model.BOM, error)
	GetActiveBOMByProductID(ctx context.Context, productID uint64) (*model.BOMDetailResponse, error)
}

type bomService struct {
	bomRepo  repository.BOMRepository
	prodRepo repository.ProductRepository
	matRepo  repository.MaterialRepository
}

func NewBOMService(bomRepo repository.BOMRepository, prodRepo repository.ProductRepository, matRepo repository.MaterialRepository) BOMService {
	return &bomService{
		bomRepo:  bomRepo,
		prodRepo: prodRepo,
		matRepo:  matRepo,
	}
}

func (s *bomService) CreateBOM(ctx context.Context, req *model.CreateBOMRequest) (*model.BOM, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Verify Product existence
	_, err := s.prodRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 2. Verify all Materials existence
	for _, item := range req.Items {
		_, err := s.matRepo.FindByID(ctx, item.MaterialID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrMaterialNotFound
			}
			return nil, err
		}
	}

	// 3. Check for Duplicate Product + Version
	exists, err := s.bomRepo.ExistsVersion(ctx, req.ProductID, req.Version)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repository.ErrDuplicateVersion
	}

	// 4. Create BOM within transaction
	return s.bomRepo.Create(ctx, req)
}

func (s *bomService) GetActiveBOMByProductID(ctx context.Context, productID uint64) (*model.BOMDetailResponse, error) {
	return s.bomRepo.FindActiveByProductID(ctx, productID)
}
