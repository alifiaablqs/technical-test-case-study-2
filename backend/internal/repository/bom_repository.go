package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"case-study-2/backend/internal/model"
)

var (
	ErrDuplicateVersion = errors.New("duplicate BOM version for product")
	ErrNoActiveBOM      = errors.New("no active BOM found for product")
)

type BOMRepository interface {
	Create(ctx context.Context, req *model.CreateBOMRequest) (*model.BOM, error)
	FindActiveByProductID(ctx context.Context, productID uint64) (*model.BOMDetailResponse, error)
	ExistsVersion(ctx context.Context, productID uint64, version uint32) (bool, error)
}

type bomRepository struct {
	db *sql.DB
}

func NewBOMRepository(db *sql.DB) BOMRepository {
	return &bomRepository{db: db}
}

func (r *bomRepository) ExistsVersion(ctx context.Context, productID uint64, version uint32) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM boms WHERE product_id = ? AND version = ?`
	err := r.db.QueryRowContext(ctx, query, productID, version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *bomRepository) Create(ctx context.Context, req *model.CreateBOMRequest) (*model.BOM, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Deactivate any currently active BOM for this product
	deactivateQuery := `UPDATE boms SET is_active = FALSE WHERE product_id = ? AND is_active = TRUE`
	if _, err := tx.ExecContext(ctx, deactivateQuery, req.ProductID); err != nil {
		return nil, err
	}

	// Insert main BOM record
	insertBOMQuery := `INSERT INTO boms (product_id, version, is_active) VALUES (?, ?, TRUE)`
	res, err := tx.ExecContext(ctx, insertBOMQuery, req.ProductID, req.Version)
	if err != nil {
		if strings.Contains(err.Error(), "1062") || strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, ErrDuplicateVersion
		}
		return nil, err
	}

	bomID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert BOM items
	insertItemQuery := `INSERT INTO bom_items (bom_id, material_id, quantity) VALUES (?, ?, ?)`
	stmt, err := tx.PrepareContext(ctx, insertItemQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var items []model.BOMItem
	for _, itemReq := range req.Items {
		itemRes, err := stmt.ExecContext(ctx, bomID, itemReq.MaterialID, itemReq.Quantity)
		if err != nil {
			return nil, err
		}
		itemID, _ := itemRes.LastInsertId()
		items = append(items, model.BOMItem{
			ID:         uint64(itemID),
			BOMID:      uint64(bomID),
			MaterialID: itemReq.MaterialID,
			Quantity:   itemReq.Quantity,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	bom := &model.BOM{
		ID:        uint64(bomID),
		ProductID: req.ProductID,
		Version:   req.Version,
		IsActive:  true,
		Items:     items,
	}

	return bom, nil
}

func (r *bomRepository) FindActiveByProductID(ctx context.Context, productID uint64) (*model.BOMDetailResponse, error) {
	// First check if product exists
	var productSKU, productName string
	prodQuery := `SELECT sku, name FROM products WHERE id = ?`
	err := r.db.QueryRowContext(ctx, prodQuery, productID).Scan(&productSKU, &productName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Fetch active BOM for product
	var bomID uint64
	var version uint32
	var isActive bool
	bomQuery := `SELECT id, version, is_active FROM boms WHERE product_id = ? AND is_active = TRUE LIMIT 1`
	err = r.db.QueryRowContext(ctx, bomQuery, productID).Scan(&bomID, &version, &isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveBOM
		}
		return nil, err
	}

	// Fetch BOM items
	itemsQuery := `SELECT bi.material_id, m.sku, m.name, m.unit, bi.quantity 
	               FROM bom_items bi 
	               JOIN materials m ON bi.material_id = m.id 
	               WHERE bi.bom_id = ? 
	               ORDER BY bi.id ASC`
	rows, err := r.db.QueryContext(ctx, itemsQuery, bomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.BOMDetailItem, 0)
	for rows.Next() {
		var item model.BOMDetailItem
		if err := rows.Scan(&item.MaterialID, &item.SKU, &item.Name, &item.Unit, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	response := &model.BOMDetailResponse{
		ProductID:   productID,
		ProductSKU:  productSKU,
		ProductName: productName,
		BOMID:       bomID,
		Version:     version,
		IsActive:    isActive,
		Items:       items,
	}

	return response, nil
}
