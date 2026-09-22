package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"
)

type mockBOMRepo struct {
	boms           map[uint64]*model.BOM
	versions       map[string]bool // key: productID_version
	seq            uint64
	simulateFailed bool
}

func newMockBOMRepo() *mockBOMRepo {
	return &mockBOMRepo{
		boms:     make(map[uint64]*model.BOM),
		versions: make(map[string]bool),
	}
}

func (m *mockBOMRepo) ExistsVersion(ctx context.Context, productID uint64, version uint32) (bool, error) {
	key := fmt.Sprintf("%d_%d", productID, version)
	return m.versions[key], nil
}

func (m *mockBOMRepo) Create(ctx context.Context, req *model.CreateBOMRequest) (*model.BOM, error) {
	if m.simulateFailed {
		// Simulates transactional rollback when insert fails
		return nil, errors.New("simulated transaction failure: insert failed")
	}

	key := fmt.Sprintf("%d_%d", req.ProductID, req.Version)
	if m.versions[key] {
		return nil, repository.ErrDuplicateVersion
	}

	// Deactivate other BOMs for product
	for _, b := range m.boms {
		if b.ProductID == req.ProductID {
			b.IsActive = false
		}
	}

	m.seq++
	m.versions[key] = true

	var items []model.BOMItem
	for idx, itemReq := range req.Items {
		items = append(items, model.BOMItem{
			ID:         uint64(idx + 1),
			BOMID:      m.seq,
			MaterialID: itemReq.MaterialID,
			Quantity:   itemReq.Quantity,
		})
	}

	bom := &model.BOM{
		ID:        m.seq,
		ProductID: req.ProductID,
		Version:   req.Version,
		IsActive:  true,
		Items:     items,
	}
	m.boms[m.seq] = bom
	return bom, nil
}

func (m *mockBOMRepo) FindActiveByProductID(ctx context.Context, productID uint64) (*model.BOMDetailResponse, error) {
	for _, bom := range m.boms {
		if bom.ProductID == productID && bom.IsActive {
			var items []model.BOMDetailItem
			for _, item := range bom.Items {
				items = append(items, model.BOMDetailItem{
					MaterialID: item.MaterialID,
					SKU:        fmt.Sprintf("RM-%03d", item.MaterialID),
					Name:       fmt.Sprintf("Material %d", item.MaterialID),
					Unit:       "pcs",
					Quantity:   item.Quantity,
				})
			}
			return &model.BOMDetailResponse{
				ProductID:   productID,
				ProductSKU:  "FG-001",
				ProductName: "Kemeja",
				BOMID:       bom.ID,
				Version:     bom.Version,
				IsActive:    true,
				Items:       items,
			}, nil
		}
	}
	return nil, repository.ErrNoActiveBOM
}

func TestBOMService_Scenarios(t *testing.T) {
	matRepo := newMockMaterialRepo()
	prodRepo := newMockProductRepo()
	bomRepo := newMockBOMRepo()

	ctx := context.Background()

	// Populate mock product and materials
	_ = prodRepo.Create(ctx, &model.Product{SKU: "FG-001", Name: "Kemeja", Unit: "pcs"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-001", Name: "Kain", Unit: "meter"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-002", Name: "Benang", Unit: "gram"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-003", Name: "Kancing", Unit: "pcs"})

	svc := service.NewBOMService(bomRepo, prodRepo, matRepo)

	// Scenario A: Create BOM FG-001 version 1
	reqA := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   1,
		Items: []model.CreateBOMItemRequest{
			{MaterialID: 1, Quantity: 1.5},
			{MaterialID: 2, Quantity: 50},
			{MaterialID: 3, Quantity: 5},
		},
	}
	bomA, err := svc.CreateBOM(ctx, reqA)
	if err != nil {
		t.Fatalf("Scenario A failed: %v", err)
	}
	if bomA.Version != 1 || len(bomA.Items) != 3 {
		t.Errorf("Scenario A invalid output: %+v", bomA)
	}

	// Scenario B: Get active BOM
	activeBOM, err := svc.GetActiveBOMByProductID(ctx, 1)
	if err != nil {
		t.Fatalf("Scenario B failed: %v", err)
	}
	if activeBOM.ProductSKU != "FG-001" || len(activeBOM.Items) != 3 {
		t.Errorf("Scenario B invalid output: %+v", activeBOM)
	}

	// Scenario C: Product not found
	reqC := &model.CreateBOMRequest{
		ProductID: 999,
		Version:   1,
		Items:     []model.CreateBOMItemRequest{{MaterialID: 1, Quantity: 100}},
	}
	_, err = svc.CreateBOM(ctx, reqC)
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Errorf("Scenario C expected ErrProductNotFound, got %v", err)
	}

	// Scenario D: Material not found
	reqD := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   2,
		Items:     []model.CreateBOMItemRequest{{MaterialID: 999, Quantity: 100}},
	}
	_, err = svc.CreateBOM(ctx, reqD)
	if !errors.Is(err, service.ErrMaterialNotFound) {
		t.Errorf("Scenario D expected ErrMaterialNotFound, got %v", err)
	}

	// Scenario E: Quantity <= 0
	reqE := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   2,
		Items:     []model.CreateBOMItemRequest{{MaterialID: 1, Quantity: -10}},
	}
	_, err = svc.CreateBOM(ctx, reqE)
	if err == nil || err.Error() != "quantity must be > 0" {
		t.Errorf("Scenario E expected quantity validation error, got %v", err)
	}

	// Scenario F: Duplicate material in request
	reqF := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   2,
		Items: []model.CreateBOMItemRequest{
			{MaterialID: 1, Quantity: 100},
			{MaterialID: 1, Quantity: 200},
		},
	}
	_, err = svc.CreateBOM(ctx, reqF)
	if err == nil || err.Error() != "duplicate material_id in BOM items" {
		t.Errorf("Scenario F expected duplicate material error, got %v", err)
	}

	// Scenario G: Duplicate product + version
	reqG := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   1, // Already created in Scenario A
		Items:     []model.CreateBOMItemRequest{{MaterialID: 1, Quantity: 100}},
	}
	_, err = svc.CreateBOM(ctx, reqG)
	if !errors.Is(err, repository.ErrDuplicateVersion) {
		t.Errorf("Scenario G expected ErrDuplicateVersion, got %v", err)
	}

	// Scenario H: Transaction rollback simulation on insert failure
	bomRepo.simulateFailed = true
	reqH := &model.CreateBOMRequest{
		ProductID: 1,
		Version:   3,
		Items:     []model.CreateBOMItemRequest{{MaterialID: 1, Quantity: 100}},
	}
	_, err = svc.CreateBOM(ctx, reqH)
	if err == nil {
		t.Errorf("Scenario H expected error on failed transaction, got nil")
	}
}
