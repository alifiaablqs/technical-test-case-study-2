package model

import (
	"errors"
	"time"
)

type BOM struct {
	ID        uint64    `json:"id"`
	ProductID uint64    `json:"product_id"`
	Version   uint32    `json:"version"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	Items     []BOMItem `json:"items,omitempty"`
}

type BOMItem struct {
	ID           uint64  `json:"id"`
	BOMID        uint64  `json:"bom_id"`
	MaterialID   uint64  `json:"material_id"`
	Quantity     float64 `json:"quantity"`
	MaterialSKU  string  `json:"sku,omitempty"`
	MaterialName string  `json:"name,omitempty"`
	MaterialUnit string  `json:"unit,omitempty"`
}

type CreateBOMItemRequest struct {
	MaterialID uint64  `json:"material_id"`
	Quantity   float64 `json:"quantity"`
}

type CreateBOMRequest struct {
	ProductID uint64                 `json:"product_id"`
	Version   uint32                 `json:"version"`
	Items     []CreateBOMItemRequest `json:"items"`
}

func (req *CreateBOMRequest) Validate() error {
	if req.ProductID == 0 {
		return errors.New("product_id is required")
	}
	if req.Version <= 0 {
		return errors.New("version must be > 0")
	}
	if len(req.Items) == 0 {
		return errors.New("items must contain at least 1 item")
	}

	seenMaterials := make(map[uint64]bool)
	for _, item := range req.Items {
		if item.MaterialID == 0 {
			return errors.New("material_id is required")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be > 0")
		}
		if seenMaterials[item.MaterialID] {
			return errors.New("duplicate material_id in BOM items")
		}
		seenMaterials[item.MaterialID] = true
	}

	return nil
}

type BOMDetailItem struct {
	MaterialID uint64  `json:"material_id"`
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	Quantity   float64 `json:"quantity"`
}

type BOMDetailResponse struct {
	ProductID   uint64          `json:"product_id"`
	ProductSKU  string          `json:"product_sku"`
	ProductName string          `json:"product_name"`
	BOMID       uint64          `json:"bom_id"`
	Version     uint32          `json:"version"`
	IsActive    bool            `json:"is_active"`
	Items       []BOMDetailItem `json:"items"`
}
