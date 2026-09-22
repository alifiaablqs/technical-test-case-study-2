package service_test

import (
	"context"
	"testing"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"
)

type mockMaterialRepo struct {
	items map[uint64]*model.Material
	seq   uint64
}

func newMockMaterialRepo() *mockMaterialRepo {
	return &mockMaterialRepo{
		items: make(map[uint64]*model.Material),
	}
}

func (m *mockMaterialRepo) FindAll(ctx context.Context) ([]*model.Material, error) {
	list := make([]*model.Material, 0, len(m.items))
	for _, item := range m.items {
		list = append(list, item)
	}
	return list, nil
}

func (m *mockMaterialRepo) FindByID(ctx context.Context, id uint64) (*model.Material, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return item, nil
}

func (m *mockMaterialRepo) Create(ctx context.Context, item *model.Material) error {
	m.seq++
	item.ID = m.seq
	item.Version = 1
	m.items[item.ID] = item
	return nil
}

func (m *mockMaterialRepo) Update(ctx context.Context, item *model.Material) error {
	if _, ok := m.items[item.ID]; !ok {
		return repository.ErrNotFound
	}
	item.Version++
	m.items[item.ID] = item
	return nil
}

func (m *mockMaterialRepo) Delete(ctx context.Context, id uint64) error {
	if _, ok := m.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

type mockProductRepo struct {
	items map[uint64]*model.Product
	seq   uint64
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		items: make(map[uint64]*model.Product),
	}
}

func (m *mockProductRepo) FindAll(ctx context.Context) ([]*model.Product, error) {
	list := make([]*model.Product, 0, len(m.items))
	for _, item := range m.items {
		list = append(list, item)
	}
	return list, nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return item, nil
}

func (m *mockProductRepo) Create(ctx context.Context, item *model.Product) error {
	m.seq++
	item.ID = m.seq
	item.Version = 1
	m.items[item.ID] = item
	return nil
}

func (m *mockProductRepo) Update(ctx context.Context, item *model.Product) error {
	if _, ok := m.items[item.ID]; !ok {
		return repository.ErrNotFound
	}
	item.Version++
	m.items[item.ID] = item
	return nil
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint64) error {
	if _, ok := m.items[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestMaterialService_CRUD(t *testing.T) {
	repo := newMockMaterialRepo()
	svc := service.NewMaterialService(repo)
	ctx := context.Background()

	// Create
	req := &model.MaterialRequest{
		SKU:      "RM-001",
		Name:     "Kain",
		Unit:     "meter",
		OnHand:   floatPtr(1000),
		Reserved: floatPtr(50),
	}

	created, err := svc.CreateMaterial(ctx, req)
	if err != nil {
		t.Fatalf("CreateMaterial failed: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("expected ID 1, got %d", created.ID)
	}
	if created.Available != 950 {
		t.Errorf("expected available 950, got %f", created.Available)
	}

	// GetByID
	fetched, err := svc.GetMaterialByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetMaterialByID failed: %v", err)
	}
	if fetched.SKU != "RM-001" {
		t.Errorf("expected SKU RM-001, got %s", fetched.SKU)
	}

	// Update
	updateReq := &model.MaterialRequest{
		SKU:      "RM-001-UPD",
		Name:     "Kain Premium",
		Unit:     "meter",
		OnHand:   floatPtr(1200),
		Reserved: floatPtr(100),
	}
	updated, err := svc.UpdateMaterial(ctx, 1, updateReq)
	if err != nil {
		t.Fatalf("UpdateMaterial failed: %v", err)
	}
	if updated.Name != "Kain Premium" {
		t.Errorf("expected updated name 'Kain Premium', got %s", updated.Name)
	}
	if updated.Available != 1100 {
		t.Errorf("expected available 1100, got %f", updated.Available)
	}

	// Delete
	err = svc.DeleteMaterial(ctx, 1)
	if err != nil {
		t.Fatalf("DeleteMaterial failed: %v", err)
	}

	_, err = svc.GetMaterialByID(ctx, 1)
	if err == nil {
		t.Errorf("expected error after delete, got nil")
	}
}
