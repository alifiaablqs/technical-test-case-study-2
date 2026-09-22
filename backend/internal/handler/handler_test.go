package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func setupTestRouter() http.Handler {
	matRepo := newMockMaterialRepo()
	prodRepo := newMockProductRepo()

	matSvc := service.NewMaterialService(matRepo)
	prodSvc := service.NewProductService(prodRepo)

	matHdl := handler.NewMaterialHandler(matSvc)
	prodHdl := handler.NewProductHandler(prodSvc)

	return handler.NewRouter(matHdl, prodHdl)
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestMaterialEndpoints(t *testing.T) {
	router := setupTestRouter()

	// 1. Create Material
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
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 2. Get All Materials
	req = httptest.NewRequest("GET", "/api/materials", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 3. Get Material By ID
	req = httptest.NewRequest("GET", "/api/materials/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 4. Update Material
	updatePayload := map[string]interface{}{
		"sku":      "RM-001-MOD",
		"name":     "Kain Katun",
		"unit":     "gram",
		"on_hand":  12000,
		"reserved": 2000,
	}
	body, _ = json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", "/api/materials/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 5. Delete Material
	req = httptest.NewRequest("DELETE", "/api/materials/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
}

func TestProductEndpoints(t *testing.T) {
	router := setupTestRouter()

	// 1. Create Product
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
		t.Fatalf("expected status 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// 2. Get All Products
	req = httptest.NewRequest("GET", "/api/products", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 3. Get Product By ID
	req = httptest.NewRequest("GET", "/api/products/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
}

func TestValidationErrors(t *testing.T) {
	router := setupTestRouter()

	// Create with empty SKU
	payload := map[string]interface{}{
		"sku":  "",
		"name": "Invalid",
		"unit": "pcs",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/materials", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}
}
