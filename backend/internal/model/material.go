package model

import (
	"errors"
	"strings"
	"time"
)

type Material struct {
	ID        uint64    `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	OnHand    float64   `json:"on_hand"`
	Reserved  float64   `json:"reserved"`
	Available float64   `json:"available"`
	Version   uint32    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *Material) CalculateAvailable() {
	m.Available = m.OnHand - m.Reserved
}

type MaterialRequest struct {
	SKU      string   `json:"sku"`
	Name     string   `json:"name"`
	Unit     string   `json:"unit"`
	OnHand   *float64 `json:"on_hand"`
	Reserved *float64 `json:"reserved"`
}

func (req *MaterialRequest) Validate() error {
	if strings.TrimSpace(req.SKU) == "" {
		return errors.New("sku is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.Unit) == "" {
		return errors.New("unit is required")
	}
	if req.OnHand != nil && *req.OnHand < 0 {
		return errors.New("on_hand must be >= 0")
	}
	if req.Reserved != nil && *req.Reserved < 0 {
		return errors.New("reserved must be >= 0")
	}
	return nil
}
