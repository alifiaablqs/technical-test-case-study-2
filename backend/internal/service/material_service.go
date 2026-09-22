package service

import (
	"context"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
)

type MaterialService interface {
	GetAllMaterials(ctx context.Context) ([]*model.Material, error)
	GetMaterialByID(ctx context.Context, id uint64) (*model.Material, error)
	CreateMaterial(ctx context.Context, req *model.MaterialRequest) (*model.Material, error)
	UpdateMaterial(ctx context.Context, id uint64, req *model.MaterialRequest) (*model.Material, error)
	DeleteMaterial(ctx context.Context, id uint64) error
}

type materialService struct {
	repo repository.MaterialRepository
}

func NewMaterialService(repo repository.MaterialRepository) MaterialService {
	return &materialService{repo: repo}
}

func (s *materialService) GetAllMaterials(ctx context.Context) ([]*model.Material, error) {
	return s.repo.FindAll(ctx)
}

func (s *materialService) GetMaterialByID(ctx context.Context, id uint64) (*model.Material, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *materialService) CreateMaterial(ctx context.Context, req *model.MaterialRequest) (*model.Material, error) {
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

	mat := &model.Material{
		SKU:      req.SKU,
		Name:     req.Name,
		Unit:     req.Unit,
		OnHand:   onHand,
		Reserved: reserved,
	}

	if err := s.repo.Create(ctx, mat); err != nil {
		return nil, err
	}

	mat.CalculateAvailable()
	return mat, nil
}

func (s *materialService) UpdateMaterial(ctx context.Context, id uint64, req *model.MaterialRequest) (*model.Material, error) {
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

func (s *materialService) DeleteMaterial(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
