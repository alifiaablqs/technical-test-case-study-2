package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"case-study-2/backend/config"
	"case-study-2/backend/internal/handler"
	"case-study-2/backend/internal/repository"
	"case-study-2/backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Warning: Database ping failed: %v. Server will start, but database operations may fail until DB is connected.", err)
	} else {
		log.Println("Database connection established successfully.")
	}

	// Repositories
	materialRepo := repository.NewMaterialRepository(db)
	productRepo := repository.NewProductRepository(db)
	bomRepo := repository.NewBOMRepository(db)

	// Services
	materialSvc := service.NewMaterialService(materialRepo)
	productSvc := service.NewProductService(productRepo)
	bomSvc := service.NewBOMService(bomRepo, productRepo, materialRepo)

	// Handlers
	materialHdl := handler.NewMaterialHandler(materialSvc)
	productHdl := handler.NewProductHandler(productSvc)
	bomHdl := handler.NewBOMHandler(bomSvc)

	// Router
	r := handler.NewRouter(materialHdl, productHdl, bomHdl)

	addr := ":" + cfg.ServerPort
	log.Printf("Starting backend server on %s ...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
