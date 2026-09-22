package model

import (
	"errors"
	"time"
)

const (
	WorkOrderStatusDraft     = "DRAFT"
	WorkOrderStatusReserved  = "RESERVED"
	WorkOrderStatusCompleted = "COMPLETED"
	WorkOrderStatusCancelled = "CANCELLED"
)

type WorkOrder struct {
	ID          uint64          `json:"id"`
	ProductID   uint64          `json:"product_id"`
	BOMID       uint64          `json:"bom_id"`
	Quantity    float64         `json:"quantity"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Items       []WorkOrderItem `json:"items,omitempty"`
}

type WorkOrderItem struct {
	ID               uint64  `json:"id"`
	WorkOrderID      uint64  `json:"work_order_id"`
	MaterialID       uint64  `json:"material_id"`
	RequiredQuantity float64 `json:"required_quantity"`
	ReservedQuantity float64 `json:"reserved_quantity"`
	IssuedQuantity   float64 `json:"issued_quantity"`
}

type CreateWorkOrderRequest struct {
	ProductID uint64  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

func (req *CreateWorkOrderRequest) Validate() error {
	if req.ProductID == 0 {
		return errors.New("product_id is required")
	}
	if req.Quantity <= 0 {
		return errors.New("quantity must be > 0")
	}
	return nil
}

type WorkOrderItemDetail struct {
	ID               uint64  `json:"id"`
	MaterialID       uint64  `json:"material_id"`
	MaterialSKU      string  `json:"material_sku"`
	MaterialName     string  `json:"material_name"`
	MaterialUnit     string  `json:"material_unit"`
	RequiredQuantity float64 `json:"required_quantity"`
	ReservedQuantity float64 `json:"reserved_quantity"`
	IssuedQuantity   float64 `json:"issued_quantity"`
}

type WorkOrderDetailResponse struct {
	ID          uint64                `json:"id"`
	ProductID   uint64                `json:"product_id"`
	ProductSKU  string                `json:"product_sku"`
	ProductName string                `json:"product_name"`
	BOMID       uint64                `json:"bom_id"`
	BOMVersion  uint32                `json:"bom_version"`
	Quantity    float64               `json:"quantity"`
	Status      string                `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Items       []WorkOrderItemDetail `json:"items"`
}
