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

type MaterialHandler struct {
	service service.MaterialService
}

func NewMaterialHandler(service service.MaterialService) *MaterialHandler {
	return &MaterialHandler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": status >= 200 && status < 300,
		"data":    data,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func (h *MaterialHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	materials, err := h.service.GetAllMaterials(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, materials)
}

func (h *MaterialHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid material ID")
		return
	}

	material, err := h.service.GetMaterialByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "material not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, material)
}

func (h *MaterialHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.MaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	material, err := h.service.CreateMaterial(r.Context(), &req)
	if err != nil {
		if err.Error() == "sku is required" || err.Error() == "name is required" ||
			err.Error() == "unit is required" || err.Error() == "on_hand must be >= 0" ||
			err.Error() == "reserved must be >= 0" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, material)
}

func (h *MaterialHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid material ID")
		return
	}

	var req model.MaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	material, err := h.service.UpdateMaterial(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "material not found")
			return
		}
		if err.Error() == "sku is required" || err.Error() == "name is required" ||
			err.Error() == "unit is required" || err.Error() == "on_hand must be >= 0" ||
			err.Error() == "reserved must be >= 0" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, material)
}

func (h *MaterialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid material ID")
		return
	}

	if err := h.service.DeleteMaterial(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "material not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "material deleted successfully"})
}
