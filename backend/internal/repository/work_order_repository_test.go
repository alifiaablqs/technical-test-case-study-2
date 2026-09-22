package repository_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"case-study-2/backend/internal/model"
	"case-study-2/backend/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(127.0.0.1:3306)/case_study_2?parseTime=true&multiStatements=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Skipping integration test: failed to open MySQL connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test: MySQL not reachable at %s: %v", dsn, err)
	}

	// Reset tables & seed test data
	seedSQL := `
	SET FOREIGN_KEY_CHECKS = 0;
	TRUNCATE TABLE work_order_items;
	TRUNCATE TABLE work_orders;
	TRUNCATE TABLE bom_items;
	TRUNCATE TABLE boms;
	TRUNCATE TABLE products;
	TRUNCATE TABLE materials;
	SET FOREIGN_KEY_CHECKS = 1;

	INSERT INTO materials (id, sku, name, unit, on_hand, reserved, version) VALUES
	(1, 'RM-001', 'Kain', 'meter', 1000.000, 0.000, 1),
	(2, 'RM-002', 'Benang', 'gram', 5000.000, 0.000, 1),
	(3, 'RM-003', 'Kancing', 'pcs', 1000.000, 0.000, 1);

	INSERT INTO products (id, sku, name, unit, on_hand, reserved, version) VALUES
	(1, 'FG-001', 'Kemeja', 'pcs', 10.000, 0.000, 1);

	INSERT INTO boms (id, product_id, version, is_active) VALUES
	(1, 1, 1, TRUE);

	INSERT INTO bom_items (id, bom_id, material_id, quantity) VALUES
	(1, 1, 1, 1.500),
	(2, 1, 2, 50.000),
	(3, 1, 3, 5.000);
	`
	_, err = db.Exec(seedSQL)
	if err != nil {
		t.Fatalf("Failed to seed test database: %v", err)
	}

	return db
}

func TestWorkOrderRepository_MySQLIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewWorkOrderRepository(db)
	ctx := context.Background()

	// Test Case A & B & F: Create WO with Qty 2 -> Expected BOM explosion (3, 100, 10) and Status RESERVED
	items := []model.WorkOrderItem{
		{MaterialID: 1, RequiredQuantity: 3},   // 1.5 * 2
		{MaterialID: 2, RequiredQuantity: 100}, // 50 * 2
		{MaterialID: 3, RequiredQuantity: 10},  // 5 * 2
	}

	wo := &model.WorkOrder{
		ProductID: 1,
		BOMID:     1,
		Quantity:  2,
	}

	createdWO, err := repo.CreateWithReservation(ctx, wo, items)
	if err != nil {
		t.Fatalf("Test A failed: CreateWithReservation returned error: %v", err)
	}

	if createdWO.Status != model.WorkOrderStatusReserved {
		t.Errorf("Test A failed: expected status RESERVED, got %s", createdWO.Status)
	}

	if createdWO.BOMID != 1 {
		t.Errorf("Test F failed: expected BOMID 1, got %d", createdWO.BOMID)
	}

	if len(createdWO.Items) != 3 {
		t.Fatalf("Test B failed: expected 3 items, got %d", len(createdWO.Items))
	}

	// Test Case C: Verify reserved quantity in materials table
	var r1, r2, r3 float64
	_ = db.QueryRow("SELECT reserved FROM materials WHERE id = 1").Scan(&r1)
	_ = db.QueryRow("SELECT reserved FROM materials WHERE id = 2").Scan(&r2)
	_ = db.QueryRow("SELECT reserved FROM materials WHERE id = 3").Scan(&r3)

	if r1 != 3 || r2 != 100 || r3 != 10 {
		t.Errorf("Test C failed: expected reserved quantities (3, 100, 10), got (%f, %f, %f)", r1, r2, r3)
	}

	// Test Case D: Create WO with insufficient stock (Requires 2000 Kain, available is 1000-3=997)
	insufficientItems := []model.WorkOrderItem{
		{MaterialID: 1, RequiredQuantity: 2000},
		{MaterialID: 2, RequiredQuantity: 2000},
		{MaterialID: 3, RequiredQuantity: 200},
	}
	woFail := &model.WorkOrder{ProductID: 1, BOMID: 1, Quantity: 40}
	_, err = repo.CreateWithReservation(ctx, woFail, insufficientItems)
	if err != repository.ErrInsufficientStock {
		t.Errorf("Test D failed: expected ErrInsufficientStock, got %v", err)
	}

	// Test Case E: Ensure all-or-nothing (partial reservation leak check)
	// Material 3 needs 2000 (only 990 available), but Material 2 needs 10 (4900 available)
	partialItems := []model.WorkOrderItem{
		{MaterialID: 2, RequiredQuantity: 10},   // Has enough stock
		{MaterialID: 3, RequiredQuantity: 2000}, // Not enough stock
	}
	woPartial := &model.WorkOrder{ProductID: 1, BOMID: 1, Quantity: 1}
	_, err = repo.CreateWithReservation(ctx, woPartial, partialItems)
	if err != repository.ErrInsufficientStock {
		t.Errorf("Test E failed: expected ErrInsufficientStock, got %v", err)
	}

	var r2After float64
	_ = db.QueryRow("SELECT reserved FROM materials WHERE id = 2").Scan(&r2After)
	if r2After != 100 {
		t.Errorf("Test E failed: partial reservation leaked! Material 2 reserved changed from 100 to %f", r2After)
	}

	// Test Case G: Concurrent reservation test
	// Remaining available stock: Kain = 997 (1000-3). Each request tries to reserve 250 Kain.
	// Max successful requests: 997 / 250 = 3. 2 requests should fail with ErrInsufficientStock.
	var wg sync.WaitGroup
	var succCount, failCount int
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reqItems := []model.WorkOrderItem{
				{MaterialID: 1, RequiredQuantity: 250},
				{MaterialID: 2, RequiredQuantity: 250},
				{MaterialID: 3, RequiredQuantity: 25},
			}
			cWO := &model.WorkOrder{ProductID: 1, BOMID: 1, Quantity: 5}
			_, err := repo.CreateWithReservation(ctx, cWO, reqItems)

			mu.Lock()
			if err == nil {
				succCount++
			} else if err == repository.ErrInsufficientStock {
				failCount++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	if succCount != 3 || failCount != 2 {
		t.Errorf("Test G failed: expected 3 successes and 2 failures under concurrency, got %d successes and %d failures", succCount, failCount)
	}
}
