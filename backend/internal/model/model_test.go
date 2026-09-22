package model_test

import (
	"testing"

	"case-study-2/backend/internal/model"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestMaterialValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     model.MaterialRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: model.MaterialRequest{
				SKU:      "RM-001",
				Name:     "Kain",
				Unit:     "meter",
				OnHand:   floatPtr(100),
				Reserved: floatPtr(10),
			},
			wantErr: false,
		},
		{
			name: "missing sku",
			req: model.MaterialRequest{
				SKU:  "",
				Name: "Kain",
				Unit: "meter",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			req: model.MaterialRequest{
				SKU:  "RM-001",
				Name: "",
				Unit: "meter",
			},
			wantErr: true,
		},
		{
			name: "missing unit",
			req: model.MaterialRequest{
				SKU:  "RM-001",
				Name: "Kain",
				Unit: "",
			},
			wantErr: true,
		},
		{
			name: "negative on_hand",
			req: model.MaterialRequest{
				SKU:    "RM-001",
				Name:   "Kain",
				Unit:   "meter",
				OnHand: floatPtr(-5),
			},
			wantErr: true,
		},
		{
			name: "negative reserved",
			req: model.MaterialRequest{
				SKU:      "RM-001",
				Name:     "Kain",
				Unit:     "meter",
				Reserved: floatPtr(-1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProductValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     model.ProductRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: model.ProductRequest{
				SKU:      "FG-001",
				Name:     "Kemeja",
				Unit:     "pcs",
				OnHand:   floatPtr(10),
				Reserved: floatPtr(2),
			},
			wantErr: false,
		},
		{
			name: "missing sku",
			req: model.ProductRequest{
				SKU:  "",
				Name: "Kemeja",
				Unit: "pcs",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAvailableCalculation(t *testing.T) {
	m := model.Material{
		OnHand:   1000,
		Reserved: 250,
	}
	m.CalculateAvailable()
	if m.Available != 750 {
		t.Errorf("expected available 750, got %f", m.Available)
	}

	p := model.Product{
		OnHand:   50,
		Reserved: 10,
	}
	p.CalculateAvailable()
	if p.Available != 40 {
		t.Errorf("expected available 40, got %f", p.Available)
	}
}
