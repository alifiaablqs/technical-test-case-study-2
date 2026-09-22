package repository

import (
	"context"
	"database/sql"
	"errors"

	"case-study-2/backend/internal/model"
)

type ProductRepository interface {
	FindAll(ctx context.Context) ([]*model.Product, error)
	FindByID(ctx context.Context, id uint64) (*model.Product, error)
	Create(ctx context.Context, p *model.Product) error
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id uint64) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll(ctx context.Context) ([]*model.Product, error) {
	query := `SELECT id, sku, name, unit, on_hand, reserved, version, created_at, updated_at 
	          FROM products ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Unit, &p.OnHand, &p.Reserved, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CalculateAvailable()
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if products == nil {
		products = []*model.Product{}
	}

	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	query := `SELECT id, sku, name, unit, on_hand, reserved, version, created_at, updated_at 
	          FROM products WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var p model.Product
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Unit, &p.OnHand, &p.Reserved, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	p.CalculateAvailable()
	return &p, nil
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	query := `INSERT INTO products (sku, name, unit, on_hand, reserved, version) 
	          VALUES (?, ?, ?, ?, ?, 1)`
	res, err := r.db.ExecContext(ctx, query, p.SKU, p.Name, p.Unit, p.OnHand, p.Reserved)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = uint64(id)

	return r.db.QueryRowContext(ctx, "SELECT created_at, updated_at, version FROM products WHERE id = ?", p.ID).
		Scan(&p.CreatedAt, &p.UpdatedAt, &p.Version)
}

func (r *productRepository) Update(ctx context.Context, p *model.Product) error {
	query := `UPDATE products SET sku = ?, name = ?, unit = ?, on_hand = ?, reserved = ?, version = version + 1 
	          WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, p.SKU, p.Name, p.Unit, p.OnHand, p.Reserved, p.ID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return r.db.QueryRowContext(ctx, "SELECT version, updated_at FROM products WHERE id = ?", p.ID).
		Scan(&p.Version, &p.UpdatedAt)
}

func (r *productRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM products WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
