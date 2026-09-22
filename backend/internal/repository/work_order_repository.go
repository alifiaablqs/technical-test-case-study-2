package repository

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"case-study-2/backend/internal/model"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidStatus     = errors.New("invalid status for operation")
)

type WorkOrderRepository interface {
	CreateWithReservation(ctx context.Context, wo *model.WorkOrder, items []model.WorkOrderItem) (*model.WorkOrder, error)
	FindByID(ctx context.Context, id uint64) (*model.WorkOrderDetailResponse, error)
	Complete(ctx context.Context, id uint64) error
}

type workOrderRepository struct {
	db *sql.DB
}

func NewWorkOrderRepository(db *sql.DB) WorkOrderRepository {
	return &workOrderRepository{db: db}
}

func (r *workOrderRepository) CreateWithReservation(ctx context.Context, wo *model.WorkOrder, items []model.WorkOrderItem) (*model.WorkOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Sort items by MaterialID to prevent deadlock during concurrent execution
	sort.Slice(items, func(i, j int) bool {
		return items[i].MaterialID < items[j].MaterialID
	})

	// 1. Pessimistic lock and verify stock for each material
	for _, item := range items {
		var onHand, reserved float64
		var version uint32
		lockQuery := `SELECT on_hand, reserved, version FROM materials WHERE id = ? FOR UPDATE`
		err := tx.QueryRowContext(ctx, lockQuery, item.MaterialID).Scan(&onHand, &reserved, &version)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}

		available := onHand - reserved
		if available < item.RequiredQuantity {
			return nil, ErrInsufficientStock
		}

		// Update reserved quantity
		updateQuery := `UPDATE materials SET reserved = reserved + ?, version = version + 1 WHERE id = ?`
		if _, err := tx.ExecContext(ctx, updateQuery, item.RequiredQuantity, item.MaterialID); err != nil {
			return nil, err
		}
	}

	// 2. Insert Work Order record
	insertWOQuery := `INSERT INTO work_orders (product_id, bom_id, quantity, status) VALUES (?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, insertWOQuery, wo.ProductID, wo.BOMID, wo.Quantity, model.WorkOrderStatusReserved)
	if err != nil {
		return nil, err
	}

	woID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	wo.ID = uint64(woID)
	wo.Status = model.WorkOrderStatusReserved

	// 3. Insert Work Order Items
	insertWOIQuery := `INSERT INTO work_order_items (work_order_id, material_id, required_quantity, reserved_quantity, issued_quantity) VALUES (?, ?, ?, ?, ?)`
	stmt, err := tx.PrepareContext(ctx, insertWOIQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var createdItems []model.WorkOrderItem
	for _, item := range items {
		itemRes, err := stmt.ExecContext(ctx, wo.ID, item.MaterialID, item.RequiredQuantity, item.RequiredQuantity, 0)
		if err != nil {
			return nil, err
		}
		itemID, _ := itemRes.LastInsertId()
		createdItems = append(createdItems, model.WorkOrderItem{
			ID:               uint64(itemID),
			WorkOrderID:      wo.ID,
			MaterialID:       item.MaterialID,
			RequiredQuantity: item.RequiredQuantity,
			ReservedQuantity: item.RequiredQuantity,
			IssuedQuantity:   0,
		})
	}

	// Fetch created_at timestamp
	_ = tx.QueryRowContext(ctx, "SELECT created_at FROM work_orders WHERE id = ?", wo.ID).Scan(&wo.CreatedAt)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	wo.Items = createdItems
	return wo, nil
}

func (r *workOrderRepository) FindByID(ctx context.Context, id uint64) (*model.WorkOrderDetailResponse, error) {
	var resp model.WorkOrderDetailResponse
	var completedAt sql.NullTime

	woQuery := `SELECT wo.id, wo.product_id, p.sku, p.name, wo.bom_id, b.version, wo.quantity, wo.status, wo.created_at, wo.completed_at
	            FROM work_orders wo
	            JOIN products p ON wo.product_id = p.id
	            JOIN boms b ON wo.bom_id = b.id
	            WHERE wo.id = ?`
	err := r.db.QueryRowContext(ctx, woQuery, id).Scan(
		&resp.ID, &resp.ProductID, &resp.ProductSKU, &resp.ProductName,
		&resp.BOMID, &resp.BOMVersion, &resp.Quantity, &resp.Status,
		&resp.CreatedAt, &completedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if completedAt.Valid {
		resp.CompletedAt = &completedAt.Time
	}

	itemsQuery := `SELECT woi.id, woi.material_id, m.sku, m.name, m.unit, woi.required_quantity, woi.reserved_quantity, woi.issued_quantity
	               FROM work_order_items woi
	               JOIN materials m ON woi.material_id = m.id
	               WHERE woi.work_order_id = ?
	               ORDER BY woi.id ASC`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.WorkOrderItemDetail, 0)
	for rows.Next() {
		var item model.WorkOrderItemDetail
		if err := rows.Scan(&item.ID, &item.MaterialID, &item.MaterialSKU, &item.MaterialName, &item.MaterialUnit, &item.RequiredQuantity, &item.ReservedQuantity, &item.IssuedQuantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp.Items = items
	return &resp, nil
}

func (r *workOrderRepository) Complete(ctx context.Context, id uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1 & 2. Read Work Order and validate status
	var status string
	var productID uint64
	var quantity float64
	woQuery := `SELECT status, product_id, quantity FROM work_orders WHERE id = ? FOR UPDATE`
	err = tx.QueryRowContext(ctx, woQuery, id).Scan(&status, &productID, &quantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if status != model.WorkOrderStatusReserved {
		return ErrInvalidStatus
	}

	// Read items to get required/reserved quantities
	itemsQuery := `SELECT id, material_id, reserved_quantity, required_quantity FROM work_order_items WHERE work_order_id = ? ORDER BY material_id ASC`
	rows, err := tx.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return err
	}

	type woItem struct {
		id               uint64
		materialID       uint64
		reservedQuantity float64
		requiredQuantity float64
	}
	var items []woItem
	for rows.Next() {
		var item woItem
		if err := rows.Scan(&item.id, &item.materialID, &item.reservedQuantity, &item.requiredQuantity); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return err
	}

	// 3 & 4. Lock materials and validate
	for _, item := range items {
		var onHand, reserved float64
		var version uint32
		lockQuery := `SELECT on_hand, reserved, version FROM materials WHERE id = ? FOR UPDATE`
		err := tx.QueryRowContext(ctx, lockQuery, item.materialID).Scan(&onHand, &reserved, &version)
		if err != nil {
			return err
		}

		issuedQuantity := item.reservedQuantity

		if reserved < issuedQuantity || onHand < issuedQuantity {
			return ErrInsufficientStock
		}

		// 5. Update material
		updateMatQuery := `UPDATE materials SET on_hand = on_hand - ?, reserved = reserved - ?, version = version + 1 WHERE id = ?`
		if _, err := tx.ExecContext(ctx, updateMatQuery, issuedQuantity, issuedQuantity, item.materialID); err != nil {
			return err
		}

		// 6. Update work_order_items
		updateItemQuery := `UPDATE work_order_items SET issued_quantity = ? WHERE id = ?`
		if _, err := tx.ExecContext(ctx, updateItemQuery, item.requiredQuantity, item.id); err != nil {
			return err
		}
	}

	// 7. Update finished good (products)
	updateProdQuery := `UPDATE products SET on_hand = on_hand + ? WHERE id = ?`
	if _, err := tx.ExecContext(ctx, updateProdQuery, quantity, productID); err != nil {
		return err
	}

	// 8. Update work_orders
	updateWOQuery := `UPDATE work_orders SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`
	if _, err := tx.ExecContext(ctx, updateWOQuery, model.WorkOrderStatusCompleted, id); err != nil {
		return err
	}

	// 9. COMMIT
	return tx.Commit()
}
