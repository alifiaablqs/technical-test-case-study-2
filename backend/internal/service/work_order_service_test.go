package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"
)

type mockWorkOrderRepo struct {
	mu        sync.Mutex
	wos       map[uint64]*model.WorkOrder
	seq       uint64
	matRepo   *mockMaterialRepo
	simulate  bool
	failOnMat uint64 // if set, simulate stock check failure for specific material
}

func newMockWorkOrderRepo(matRepo *mockMaterialRepo) *mockWorkOrderRepo {
	return &mockWorkOrderRepo{
		wos:     make(map[uint64]*model.WorkOrder),
		matRepo: matRepo,
	}
}

func (m *mockWorkOrderRepo) CreateWithReservation(ctx context.Context, wo *model.WorkOrder, items []model.WorkOrderItem) (*model.WorkOrder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check stock for ALL items first (atomicity/rollback verification)
	for _, item := range items {
		if m.failOnMat > 0 && item.MaterialID == m.failOnMat {
			return nil, repository.ErrInsufficientStock
		}

		mat, err := m.matRepo.FindByID(ctx, item.MaterialID)
		if err != nil {
			return nil, err
		}

		available := mat.OnHand - mat.Reserved
		if available < item.RequiredQuantity {
			return nil, repository.ErrInsufficientStock
		}
	}

	// 2. Reserve stock only if ALL materials have sufficient stock
	var createdItems []model.WorkOrderItem
	for idx, item := range items {
		mat, _ := m.matRepo.FindByID(ctx, item.MaterialID)
		mat.Reserved += item.RequiredQuantity

		createdItems = append(createdItems, model.WorkOrderItem{
			ID:               uint64(idx + 1),
			WorkOrderID:      m.seq + 1,
			MaterialID:       item.MaterialID,
			RequiredQuantity: item.RequiredQuantity,
			ReservedQuantity: item.RequiredQuantity,
			IssuedQuantity:   0,
		})
	}

	m.seq++
	wo.ID = m.seq
	wo.Status = model.WorkOrderStatusReserved
	wo.Items = createdItems
	m.wos[m.seq] = wo

	return wo, nil
}

func (m *mockWorkOrderRepo) FindByID(ctx context.Context, id uint64) (*model.WorkOrderDetailResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wo, ok := m.wos[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	items := make([]model.WorkOrderItemDetail, 0, len(wo.Items))
	for _, item := range wo.Items {
		mat, _ := m.matRepo.FindByID(ctx, item.MaterialID)
		items = append(items, model.WorkOrderItemDetail{
			ID:               item.ID,
			MaterialID:       item.MaterialID,
			MaterialSKU:      mat.SKU,
			MaterialName:     mat.Name,
			MaterialUnit:     mat.Unit,
			RequiredQuantity: item.RequiredQuantity,
			ReservedQuantity: item.ReservedQuantity,
			IssuedQuantity:   item.IssuedQuantity,
		})
	}

	return &model.WorkOrderDetailResponse{
		ID:          wo.ID,
		ProductID:   wo.ProductID,
		ProductSKU:  "FG-001",
		ProductName: "Kemeja",
		BOMID:       wo.BOMID,
		BOMVersion:  1,
		Quantity:    wo.Quantity,
		Status:      wo.Status,
		Items:       items,
	}, nil
}

func (m *mockWorkOrderRepo) Complete(ctx context.Context, id uint64) error {
	return nil
}

func TestWorkOrderService_Scenarios(t *testing.T) {
	matRepo := newMockMaterialRepo()
	prodRepo := newMockProductRepo()
	bomRepo := newMockBOMRepo()
	woRepo := newMockWorkOrderRepo(matRepo)

	ctx := context.Background()

	// Seed product and materials
	_ = prodRepo.Create(ctx, &model.Product{SKU: "FG-001", Name: "Kemeja", Unit: "pcs"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-001", Name: "Kain", Unit: "gram", OnHand: 10000, Reserved: 0})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-002", Name: "Benang", Unit: "gram", OnHand: 5000, Reserved: 0})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-003", Name: "Kancing", Unit: "pcs", OnHand: 1000, Reserved: 0})

	// Seed active BOM (v1: 500 Kain, 50 Benang, 5 Kancing)
	_, _ = bomRepo.Create(ctx, &model.CreateBOMRequest{
		ProductID: 1,
		Version:   1,
		Items: []model.CreateBOMItemRequest{
			{MaterialID: 1, Quantity: 500},
			{MaterialID: 2, Quantity: 50},
			{MaterialID: 3, Quantity: 5},
		},
	})

	svc := service.NewWorkOrderService(woRepo, bomRepo, prodRepo)

	// Scenario A & B & F: Create Work Order Qty 2 -> Expected BOM explosion
	reqA := &model.CreateWorkOrderRequest{ProductID: 1, Quantity: 2}
	woA, err := svc.CreateWorkOrder(ctx, reqA)
	if err != nil {
		t.Fatalf("Scenario A failed: %v", err)
	}
	if woA.Status != model.WorkOrderStatusReserved {
		t.Errorf("Scenario A expected status RESERVED, got %s", woA.Status)
	}
	if woA.BOMID != 1 {
		t.Errorf("Scenario F expected BOMID 1, got %d", woA.BOMID)
	}

	// Verify BOM explosion: 500*2 = 1000, 50*2 = 100, 5*2 = 10
	expectedExplosion := map[uint64]float64{1: 1000, 2: 100, 3: 10}
	for _, item := range woA.Items {
		if expectedExplosion[item.MaterialID] != item.RequiredQuantity {
			t.Errorf("Scenario B material %d expected required_qty %f, got %f", item.MaterialID, expectedExplosion[item.MaterialID], item.RequiredQuantity)
		}
	}

	// Scenario C: Verify material reserved count increased
	m1, _ := matRepo.FindByID(ctx, 1)
	if m1.Reserved != 1000 {
		t.Errorf("Scenario C material 1 expected reserved 1000, got %f", m1.Reserved)
	}

	// Scenario D: Create WO with stock insufficient
	reqD := &model.CreateWorkOrderRequest{ProductID: 1, Quantity: 500} // Requires 250,000 Kain, on_hand is 10,000
	_, err = svc.CreateWorkOrder(ctx, reqD)
	if !errors.Is(err, repository.ErrInsufficientStock) {
		t.Errorf("Scenario D expected ErrInsufficientStock, got %v", err)
	}

	// Scenario E: Verify all-or-nothing rollback (partial reservation must NOT happen)
	// Material 1 has enough stock for qty 10 (5000 Kain), but set mock to fail on Material 3
	woRepo.failOnMat = 3
	reservedM1Before := m1.Reserved
	reqE := &model.CreateWorkOrderRequest{ProductID: 1, Quantity: 10}
	_, err = svc.CreateWorkOrder(ctx, reqE)
	if !errors.Is(err, repository.ErrInsufficientStock) {
		t.Errorf("Scenario E expected failure, got %v", err)
	}
	m1After, _ := matRepo.FindByID(ctx, 1)
	if m1After.Reserved != reservedM1Before {
		t.Errorf("Scenario E failed: Material 1 reserved changed from %f to %f (partial reservation leak!)", reservedM1Before, m1After.Reserved)
	}
	woRepo.failOnMat = 0

	// Scenario G: Concurrent reservation test
	var wg sync.WaitGroup
	concurrentErrors := 0
	var errMu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateWorkOrder(ctx, &model.CreateWorkOrderRequest{ProductID: 1, Quantity: 5})
			if err != nil {
				errMu.Lock()
				concurrentErrors++
				errMu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Remaining available stock after first WO (qty 2):
	// Kain: 10000 - 1000 = 9000
	// 5 concurrent requests of Qty 5 require 5 * 2500 = 12,500 Kain total, which exceeds 9000 available.
	// Therefore, some requests MUST succeed and some MUST fail with stock error without corrupting state.
	if concurrentErrors == 0 {
		t.Errorf("Scenario G expected at least some concurrent requests to fail due to stock limit")
	}
}
