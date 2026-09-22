package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func setupTestRouterWithBOM() (http.Handler, *mockProductRepo, *mockMaterialRepo) {
	matRepo := newMockMaterialRepo()
	prodRepo := newMockProductRepo()
	bomRepo := newMockBOMRepo()

	matSvc := service.NewMaterialService(matRepo)
	prodSvc := service.NewProductService(prodRepo)
	bomSvc := service.NewBOMService(bomRepo, prodRepo, matRepo)

	matHdl := handler.NewMaterialHandler(matSvc)
	prodHdl := handler.NewProductHandler(prodSvc)
	bomHdl := handler.NewBOMHandler(bomSvc)

	return handler.NewRouter(matHdl, prodHdl, bomHdl), prodRepo, matRepo
}

func TestMaterialEndpoints(t *testing.T) {
	router, _, _ := setupTestRouterWithBOM()

	createPayload := map[string]interface{}{
		"sku":      "RM-001",
		"name":     "Kain",
		"unit":     "gram",
		"on_hand":  10000,
		"reserved": 0,
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/materials", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}
}

func TestProductEndpoints(t *testing.T) {
	router, _, _ := setupTestRouterWithBOM()

	createPayload := map[string]interface{}{
		"sku":      "FG-001",
		"name":     "Kemeja",
		"unit":     "pcs",
		"on_hand":  10,
		"reserved": 0,
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}
}

func TestBOMEndpoints_HTTPScenarios(t *testing.T) {
	router, prodRepo, matRepo := setupTestRouterWithBOM()
	ctx := context.Background()

	// Seed product and materials
	_ = prodRepo.Create(ctx, &model.Product{SKU: "FG-001", Name: "Kemeja", Unit: "pcs"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-001", Name: "Kain", Unit: "gram"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-002", Name: "Benang", Unit: "gram"})
	_ = matRepo.Create(ctx, &model.Material{SKU: "RM-003", Name: "Kancing", Unit: "pcs"})

	// Scenario A: Create BOM FG-001 version 1 (Expected 201)
	payloadA := map[string]interface{}{
		"product_id": 1,
		"version":    1,
		"items": []map[string]interface{}{
			{"material_id": 1, "quantity": 500},
			{"material_id": 2, "quantity": 50},
			{"material_id": 3, "quantity": 5},
		},
	}
	body, _ := json.Marshal(payloadA)
	req := httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Scenario A failed: expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario B: GET /api/products/1/bom (Expected 200)
	req = httptest.NewRequest("GET", "/api/products/1/bom", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Scenario B failed: expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario C: Product not found (Expected 404)
	payloadC := map[string]interface{}{
		"product_id": 999,
		"version":    1,
		"items":      []map[string]interface{}{{"material_id": 1, "quantity": 100}},
	}
	body, _ = json.Marshal(payloadC)
	req = httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Scenario C failed: expected 404 Not Found, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario D: Material not found (Expected 404)
	payloadD := map[string]interface{}{
		"product_id": 1,
		"version":    2,
		"items":      []map[string]interface{}{{"material_id": 999, "quantity": 100}},
	}
	body, _ = json.Marshal(payloadD)
	req = httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Scenario D failed: expected 404 Not Found, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario E: Quantity <= 0 (Expected 400)
	payloadE := map[string]interface{}{
		"product_id": 1,
		"version":    2,
		"items":      []map[string]interface{}{{"material_id": 1, "quantity": 0}},
	}
	body, _ = json.Marshal(payloadE)
	req = httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Scenario E failed: expected 400 Bad Request, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario F: Duplicate material in request (Expected 400)
	payloadF := map[string]interface{}{
		"product_id": 1,
		"version":    2,
		"items": []map[string]interface{}{
			{"material_id": 1, "quantity": 100},
			{"material_id": 1, "quantity": 200},
		},
	}
	body, _ = json.Marshal(payloadF)
	req = httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Scenario F failed: expected 400 Bad Request, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Scenario G: Duplicate product + version (Expected 409 Conflict)
	payloadG := map[string]interface{}{
		"product_id": 1,
		"version":    1, // Duplicate
		"items":      []map[string]interface{}{{"material_id": 1, "quantity": 100}},
	}
	body, _ = json.Marshal(payloadG)
	req = httptest.NewRequest("POST", "/api/boms", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("Scenario G failed: expected 409 Conflict, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
