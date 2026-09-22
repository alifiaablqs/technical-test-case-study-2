package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"case-study-2/backend/internal/handler"
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

type mockBOMRepo struct {
	boms     map[uint64]*model.BOM
	versions map[string]bool
	seq      uint64
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
	key := fmt.Sprintf("%d_%d", req.ProductID, req.Version)
	if m.versions[key] {
		return nil, repository.ErrDuplicateVersion
	}

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

type mockWorkOrderRepo struct {
	mu      sync.Mutex
	wos     map[uint64]*model.WorkOrder
	seq     uint64
	matRepo *mockMaterialRepo
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

	for _, item := range items {
		mat, err := m.matRepo.FindByID(ctx, item.MaterialID)
		if err != nil {
			return nil, err
		}
		available := mat.OnHand - mat.Reserved
		if available < item.RequiredQuantity {
			return nil, repository.ErrInsufficientStock
		}
	}

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

func setupTestRouterFull() (http.Handler, *mockProductRepo, *mockMaterialRepo, *mockBOMRepo) {
	matRepo := newMockMaterialRepo()
	prodRepo := newMockProductRepo()
	bomRepo := newMockBOMRepo()
	woRepo := newMockWorkOrderRepo(matRepo)

	matSvc := service.NewMaterialService(matRepo)
	prodSvc := service.NewProductService(prodRepo)
	bomSvc := service.NewBOMService(bomRepo, prodRepo, matRepo)
	woSvc := service.NewWorkOrderService(woRepo, bomRepo, prodRepo)

	matHdl := handler.NewMaterialHandler(matSvc)
	prodHdl := handler.NewProductHandler(prodSvc)
	bomHdl := handler.NewBOMHandler(bomSvc)
	woHdl := handler.NewWorkOrderHandler(woSvc)

	return handler.NewRouter(matHdl, prodHdl, bomHdl, woHdl), prodRepo, matRepo, bomRepo
}

func TestWorkOrderEndpoints(t *testing.T) {
	router, prodRepo, matRepo, bomRepo := setupTestRouterFull()
	ctx := context.Background()

	// Seed product and materials
	_ = prodRepo.Create(ctx, &model.Product{SKU: "FG-001", Name: "Kemeja", Unit: "pcs"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-001", Name: "Kain", Unit: "gram", OnHand: 10000, Reserved: 0})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-002", Name: "Benang", Unit: "gram", OnHand: 5000, Reserved: 0})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-003", Name: "Kancing", Unit: "pcs", OnHand: 1000, Reserved: 0})

	// Seed active BOM v1
	_, _ = bomRepo.Create(ctx, &model.CreateBOMRequest{
		ProductID: 1,
		Version:   1,
		Items: []model.CreateBOMItemRequest{
			{MaterialID: 1, Quantity: 500},
			{MaterialID: 2, Quantity: 50},
			{MaterialID: 3, Quantity: 5},
		},
	})

	// 1. Create Work Order (Expected 201 Created)
	payloadA := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}
	body, _ := json.Marshal(payloadA)
	req := httptest.NewRequest("POST", "/api/work-orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 2. GET Work Order by ID (Expected 200 OK)
	req = httptest.NewRequest("GET", "/api/work-orders/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 3. Create Work Order with Quantity <= 0 (Expected 400 Bad Request)
	payloadC := map[string]interface{}{"product_id": 1, "quantity": 0}
	body, _ = json.Marshal(payloadC)
	req = httptest.NewRequest("POST", "/api/work-orders", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}

	// 4. Product Not Found (Expected 404 Not Found)
	payloadD := map[string]interface{}{"product_id": 999, "quantity": 1}
	body, _ = json.Marshal(payloadD)
	req = httptest.NewRequest("POST", "/api/work-orders", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d", rec.Code)
	}

	// 5. Insufficient Stock (Expected 409 Conflict)
	payloadF := map[string]interface{}{"product_id": 1, "quantity": 500}
	body, _ = json.Marshal(payloadF)
	req = httptest.NewRequest("POST", "/api/work-orders", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
