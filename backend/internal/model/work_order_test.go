package model_test

import (
	"testing"

	"case-study-2/backend/internal/model"
)

func TestWorkOrderValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     model.CreateWorkOrderRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     model.CreateWorkOrderRequest{ProductID: 1, Quantity: 2},
			wantErr: false,
		},
		{
			name:    "missing product_id",
			req:     model.CreateWorkOrderRequest{ProductID: 0, Quantity: 2},
			wantErr: true,
		},
		{
			name:    "quantity <= 0",
			req:     model.CreateWorkOrderRequest{ProductID: 1, Quantity: 0},
			wantErr: true,
		},
		{
			name:    "negative quantity",
			req:     model.CreateWorkOrderRequest{ProductID: 1, Quantity: -5},
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
