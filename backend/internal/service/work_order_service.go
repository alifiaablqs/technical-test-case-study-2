package service

import (
	"context"
	"errors"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
)

var (
	ErrNoActiveBOMForProduct = errors.New("product has no active BOM")
)

type WorkOrderService interface {
	CreateWorkOrder(ctx context.Context, req *model.CreateWorkOrderRequest) (*model.WorkOrder, error)
	GetWorkOrderByID(ctx context.Context, id uint64) (*model.WorkOrderDetailResponse, error)
	CompleteWorkOrder(ctx context.Context, id uint64) error
}

type workOrderService struct {
	woRepo   repository.WorkOrderRepository
	bomRepo  repository.BOMRepository
	prodRepo repository.ProductRepository
}

func NewWorkOrderService(woRepo repository.WorkOrderRepository, bomRepo repository.BOMRepository, prodRepo repository.ProductRepository) WorkOrderService {
	return &workOrderService{
		woRepo:   woRepo,
		bomRepo:  bomRepo,
		prodRepo: prodRepo,
	}
}

func (s *workOrderService) CreateWorkOrder(ctx context.Context, req *model.CreateWorkOrderRequest) (*model.WorkOrder, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Verify Product existence
	_, err := s.prodRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 2. Fetch Active BOM for product
	activeBOM, err := s.bomRepo.FindActiveByProductID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveBOM) || errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNoActiveBOMForProduct
		}
		return nil, err
	}

	// 3. Perform BOM Explosion (required_quantity = bom_item.quantity * work_order.quantity)
	woItems := make([]model.WorkOrderItem, 0, len(activeBOM.Items))
	for _, item := range activeBOM.Items {
		requiredQty := item.Quantity * req.Quantity
		woItems = append(woItems, model.WorkOrderItem{
			MaterialID:       item.MaterialID,
			RequiredQuantity: requiredQty,
			ReservedQuantity: requiredQty,
		})
	}

	wo := &model.WorkOrder{
		ProductID: req.ProductID,
		BOMID:     activeBOM.BOMID,
		Quantity:  req.Quantity,
		Status:    model.WorkOrderStatusReserved,
	}

	// 4. Create Work Order and reserve materials in single transaction
	return s.woRepo.CreateWithReservation(ctx, wo, woItems)
}

func (s *workOrderService) GetWorkOrderByID(ctx context.Context, id uint64) (*model.WorkOrderDetailResponse, error) {
	return s.woRepo.FindByID(ctx, id)
}

func (s *workOrderService) CompleteWorkOrder(ctx context.Context, id uint64) error {
	return s.woRepo.Complete(ctx, id)
}
