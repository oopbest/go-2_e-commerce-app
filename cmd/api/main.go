package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/product"
)

func main() {
	// 1. ประกอบชิ้นส่วน (Dependency Injection) จากล่างขึ้นบน
	productRepo := product.NewInMemoryRepository()
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// 2. สร้าง ServeMux Router และลงทะเบียน Routes
	mux := http.NewServeMux()

	// ลงทะเบียน Routes ของ Product
	productHandler.RegisterRoutes(mux)

	// Global Health Check Route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"message": "E-Commerce API is running (Clean Architecture)",
		})
	})

	// 3. เริ่มต้น Server
	port := ":8080"
	fmt.Printf("🚀 E-Commerce API server started on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
