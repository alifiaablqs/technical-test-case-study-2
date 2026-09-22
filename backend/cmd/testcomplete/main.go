package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"case-study-2/backend/config"
	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.LoadConfig()
	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup: Reset DB slightly for clean test
	db.Exec("DELETE FROM work_order_items")
	db.Exec("DELETE FROM work_orders")
	db.Exec("ALTER TABLE work_orders AUTO_INCREMENT = 10")
	db.Exec("UPDATE materials SET on_hand = 10000, reserved = 0 WHERE id = 1")
	db.Exec("UPDATE materials SET on_hand = 5000, reserved = 0 WHERE id = 2")
	db.Exec("UPDATE materials SET on_hand = 1000, reserved = 0 WHERE id = 3")
	db.Exec("UPDATE products SET on_hand = 10, reserved = 0 WHERE id = 1")

	woRepo := repository.NewWorkOrderRepository(db)
	bomRepo := repository.NewBOMRepository(db)
	prodRepo := repository.NewProductRepository(db)
	woService := service.NewWorkOrderService(woRepo, bomRepo, prodRepo)
	ctx := context.Background()

	// Create WO qty 2
	fmt.Println("--- Creating WO Qty 2 ---")
	req := &model.CreateWorkOrderRequest{ProductID: 1, Quantity: 2}
	wo, err := woService.CreateWorkOrder(ctx, req)
	if err != nil {
		log.Fatal("Failed to create WO:", err)
	}
	fmt.Printf("Created WO ID: %d, Status: %s\n", wo.ID, wo.Status)

	printStock(db)

	fmt.Println("\n--- Completing WO 10 ---")
	err = woService.CompleteWorkOrder(ctx, 10)
	if err != nil {
		log.Fatal("Failed to complete WO:", err)
	}
	fmt.Println("Complete WO 10: SUCCESS")

	printStock(db)

	fmt.Println("\n--- Completing WO 10 again (Expected 409) ---")
	err = woService.CompleteWorkOrder(ctx, 10)
	if err != nil {
		fmt.Printf("Complete WO 10 Again Error: %v\n", err)
	} else {
		log.Fatal("Expected error, got nil")
	}

	woDetail, _ := woService.GetWorkOrderByID(ctx, 10)
	fmt.Printf("\nWO Status: %s\n", woDetail.Status)
	fmt.Printf("Completed At: %v\n", woDetail.CompletedAt)
	for _, item := range woDetail.Items {
		fmt.Printf("Item %s: req=%.0f, res=%.0f, issued=%.0f\n", item.MaterialSKU, item.RequiredQuantity, item.ReservedQuantity, item.IssuedQuantity)
	}
}

func printStock(db *sql.DB) {
	fmt.Println("Current Stock:")
	rows, _ := db.Query("SELECT sku, on_hand, reserved FROM materials ORDER BY id ASC")
	defer rows.Close()
	for rows.Next() {
		var sku string
		var oh, res float64
		rows.Scan(&sku, &oh, &res)
		fmt.Printf("%s: on_hand %.0f, reserved %.0f\n", sku, oh, res)
	}
	var fgOh float64
	db.QueryRow("SELECT on_hand FROM products WHERE id = 1").Scan(&fgOh)
	fmt.Printf("FG-001: on_hand %.0f\n", fgOh)
}
