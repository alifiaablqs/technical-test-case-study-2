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

type WorkOrderHandler struct {
	service service.WorkOrderService
}

func NewWorkOrderHandler(service service.WorkOrderService) *WorkOrderHandler {
	return &WorkOrderHandler{service: service}
}

func (h *WorkOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wo, err := h.service.CreateWorkOrder(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		if errors.Is(err, service.ErrNoActiveBOMForProduct) {
			writeError(w, http.StatusConflict, "product has no active BOM")
			return
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, "insufficient stock for material reservation")
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "material not found")
			return
		}
		if err.Error() == "product_id is required" || err.Error() == "quantity must be > 0" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, wo)
}

func (h *WorkOrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	woDetail, err := h.service.GetWorkOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, woDetail)
}

func (h *WorkOrderHandler) Complete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	err = h.service.CompleteWorkOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "work order not found")
			return
		}
		if errors.Is(err, repository.ErrInvalidStatus) {
			writeError(w, http.StatusConflict, "invalid status for operation")
			return
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, "insufficient stock")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
