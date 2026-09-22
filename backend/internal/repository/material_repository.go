package repository

import (
	"context"
	"database/sql"
	"errors"

	"case-study-2/backend/internal/model"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicateSKU = errors.New("sku already exists")
)

type MaterialRepository interface {
	FindAll(ctx context.Context) ([]*model.Material, error)
	FindByID(ctx context.Context, id uint64) (*model.Material, error)
	Create(ctx context.Context, m *model.Material) error
	Update(ctx context.Context, m *model.Material) error
	Delete(ctx context.Context, id uint64) error
}

type materialRepository struct {
	db *sql.DB
}

func NewMaterialRepository(db *sql.DB) MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) FindAll(ctx context.Context) ([]*model.Material, error) {
	query := `SELECT id, sku, name, unit, on_hand, reserved, version, created_at, updated_at 
	          FROM materials ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []*model.Material
	for rows.Next() {
		var m model.Material
		if err := rows.Scan(&m.ID, &m.SKU, &m.Name, &m.Unit, &m.OnHand, &m.Reserved, &m.Version, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.CalculateAvailable()
		materials = append(materials, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if materials == nil {
		materials = []*model.Material{}
	}

	return materials, nil
}

func (r *materialRepository) FindByID(ctx context.Context, id uint64) (*model.Material, error) {
	query := `SELECT id, sku, name, unit, on_hand, reserved, version, created_at, updated_at 
	          FROM materials WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var m model.Material
	err := row.Scan(&m.ID, &m.SKU, &m.Name, &m.Unit, &m.OnHand, &m.Reserved, &m.Version, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	m.CalculateAvailable()
	return &m, nil
}

func (r *materialRepository) Create(ctx context.Context, m *model.Material) error {
	query := `INSERT INTO materials (sku, name, unit, on_hand, reserved, version) 
	          VALUES (?, ?, ?, ?, ?, 1)`
	res, err := r.db.ExecContext(ctx, query, m.SKU, m.Name, m.Unit, m.OnHand, m.Reserved)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = uint64(id)

	// Fetch timestamps
	return r.db.QueryRowContext(ctx, "SELECT created_at, updated_at, version FROM materials WHERE id = ?", m.ID).
		Scan(&m.CreatedAt, &m.UpdatedAt, &m.Version)
}

func (r *materialRepository) Update(ctx context.Context, m *model.Material) error {
	query := `UPDATE materials SET sku = ?, name = ?, unit = ?, on_hand = ?, reserved = ?, version = version + 1 
	          WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, m.SKU, m.Name, m.Unit, m.OnHand, m.Reserved, m.ID)
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

	return r.db.QueryRowContext(ctx, "SELECT version, updated_at FROM materials WHERE id = ?", m.ID).
		Scan(&m.Version, &m.UpdatedAt)
}

func (r *materialRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM materials WHERE id = ?`
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
