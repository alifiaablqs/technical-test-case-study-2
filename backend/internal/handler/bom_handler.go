package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"
)

type BOMHandler struct {
	service service.BOMService
}

func NewBOMHandler(service service.BOMService) *BOMHandler {
	return &BOMHandler{service: service}
}

func (h *BOMHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBOMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bom, err := h.service.CreateBOM(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		if errors.Is(err, service.ErrMaterialNotFound) {
			writeError(w, http.StatusNotFound, "material not found")
			return
		}
		if errors.Is(err, repository.ErrDuplicateVersion) {
			writeError(w, http.StatusConflict, "duplicate BOM version for product")
			return
		}
		if err.Error() == "product_id is required" || err.Error() == "version must be > 0" ||
			err.Error() == "items must contain at least 1 item" || err.Error() == "material_id is required" ||
			err.Error() == "quantity must be > 0" || err.Error() == "duplicate material_id in BOM items" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, bom)
}

func (h *BOMHandler) GetActiveByProductID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	productID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	bomDetail, err := h.service.GetActiveBOMByProductID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, service.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		if errors.Is(err, repository.ErrNoActiveBOM) {
			writeError(w, http.StatusNotFound, "active BOM not found for product")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bomDetail)
}
