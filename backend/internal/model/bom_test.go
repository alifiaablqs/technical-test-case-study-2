package model_test

import (
	"testing"

	"case-study-2/backend/internal/model"
)

func TestBOMValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     model.CreateBOMRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: model.CreateBOMRequest{
				ProductID: 1,
				Version:   1,
				Items: []model.CreateBOMItemRequest{
					{MaterialID: 1, Quantity: 500},
					{MaterialID: 2, Quantity: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "missing product_id",
			req: model.CreateBOMRequest{
				ProductID: 0,
				Version:   1,
				Items: []model.CreateBOMItemRequest{
					{MaterialID: 1, Quantity: 500},
				},
			},
			wantErr: true,
		},
		{
			name: "version <= 0",
			req: model.CreateBOMRequest{
				ProductID: 1,
				Version:   0,
				Items: []model.CreateBOMItemRequest{
					{MaterialID: 1, Quantity: 500},
				},
			},
			wantErr: true,
		},
		{
			name: "empty items",
			req: model.CreateBOMRequest{
				ProductID: 1,
				Version:   1,
				Items:     []model.CreateBOMItemRequest{},
			},
			wantErr: true,
		},
		{
			name: "quantity <= 0",
			req: model.CreateBOMRequest{
				ProductID: 1,
				Version:   1,
				Items: []model.CreateBOMItemRequest{
					{MaterialID: 1, Quantity: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate material in same request",
			req: model.CreateBOMRequest{
				ProductID: 1,
				Version:   1,
				Items: []model.CreateBOMItemRequest{
					{MaterialID: 1, Quantity: 500},
					{MaterialID: 1, Quantity: 200},
				},
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
