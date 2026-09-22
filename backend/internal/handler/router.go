package handler

import (
	"net/http"
)

func NewRouter(materialHandler *MaterialHandler, productHandler *ProductHandler, bomHandler *BOMHandler, workOrderHandler *WorkOrderHandler) http.Handler {
	mux := http.NewServeMux()

	// Material routes
	mux.HandleFunc("GET /api/materials", materialHandler.GetAll)
	mux.HandleFunc("POST /api/materials", materialHandler.Create)
	mux.HandleFunc("GET /api/materials/{id}", materialHandler.GetByID)
	mux.HandleFunc("PUT /api/materials/{id}", materialHandler.Update)
	mux.HandleFunc("DELETE /api/materials/{id}", materialHandler.Delete)

	// Product routes
	mux.HandleFunc("GET /api/products", productHandler.GetAll)
	mux.HandleFunc("POST /api/products", productHandler.Create)
	mux.HandleFunc("GET /api/products/{id}", productHandler.GetByID)
	mux.HandleFunc("PUT /api/products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /api/products/{id}", productHandler.Delete)

	// BOM routes
	mux.HandleFunc("POST /api/boms", bomHandler.Create)
	mux.HandleFunc("GET /api/products/{id}/bom", bomHandler.GetActiveByProductID)

	// Work Order routes
	mux.HandleFunc("POST /api/work-orders", workOrderHandler.Create)
	mux.HandleFunc("GET /api/work-orders/{id}", workOrderHandler.GetByID)
	mux.HandleFunc("POST /api/work-orders/{id}/complete", workOrderHandler.Complete)

	return mux
}
