package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/internal/product"
	"github.com/oopbest/ecommerce-app/internal/user"
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
	defer db.Close()

	// JWT Secret Key (ในระบบจริงควรอ่านจาก Environment Variable)
	jwtSecret := "my-super-secret-jwt-key-for-learning-32bytes"

	// ==========================================
	// 2. Dependency Injection: User & Auth Module
	// ==========================================
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, jwtSecret)
	userHandler := user.NewHandler(userService)

	// ==========================================
	// 3. Dependency Injection: Product Module
	// ==========================================
	productRepo := product.NewPostgresRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// ==========================================
	// 4. Middleware & Router Setup
	// ==========================================
	mux := http.NewServeMux()

	// Auth Middleware
	auth := middleware.AuthMiddleware(jwtSecret)

	// ลงทะเบียน Routes
	userHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux, auth)

	// Health Check Route
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"message":  "E-Commerce API is running (Auth & Security Enabled)",
			"database": "PostgreSQL (Connected)",
		})
	})

	// ==========================================
	// 5. Start Server
	// ==========================================
	port := ":8080"
	fmt.Printf("🚀 E-Commerce API server started on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
