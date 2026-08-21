package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/product"
)

func main() {
	// ==========================================
	// 1. เชื่อมต่อ PostgreSQL Database
	// ==========================================
	dbCfg := database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgrespassword",
		DBName:   "ecommerce_db",
		SSLMode:  "disable",
	}

	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close() // ปิด Database Connection เมื่อ Server ดับ

	// ==========================================
	// 2. Dependency Injection (สลับมาใช้ Postgres Repository!)
	// ==========================================
	// เดิม: productRepo := product.NewInMemoryRepository()
	productRepo := product.NewPostgresRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// ==========================================
	// 3. Router Setup & Route Registration
	// ==========================================
	mux := http.NewServeMux()

	// ลงทะเบียน Routes ของ Product
	productHandler.RegisterRoutes(mux)

	// Global Health Check Route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"message":  "E-Commerce API is running",
			"database": "PostgreSQL (Connected)",
		})
	})

	// ==========================================
	// 4. Start Server
	// ==========================================
	port := ":8080"
	fmt.Printf("🚀 E-Commerce API server started on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
