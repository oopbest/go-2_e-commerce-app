package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oopbest/ecommerce-app/internal/config"
	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/internal/order"
	"github.com/oopbest/ecommerce-app/internal/product"
	"github.com/oopbest/ecommerce-app/internal/user"
)

func main() {
	// ==========================================
	// 1. โหลด Configuration จาก .env
	// ==========================================
	cfg := config.Load()

	// ==========================================
	// 2. ตั้งค่า Structured Logger (log/slog)
	// ==========================================
	var logger *slog.Logger
	if cfg.AppEnv == "production" {
		// Production: ใช้ JSON Format สำหรับส่งไปยัง Cloud Logging (Loki, Datadog, CloudWatch)
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		// Development: ใช้ Text Format ให้อ่านง่ายบน Terminal
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	// ==========================================
	// 3. เชื่อมต่อ PostgreSQL Database
	// ==========================================
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
		slog.Info("🔒 Database connection pool closed")
	}()

	// ==========================================
	// 4. Dependency Injection: Modules
	// ==========================================
	// User Module
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, cfg.JWTSecret)
	userHandler := user.NewHandler(userService)

	// Product Module
	productRepo := product.NewPostgresRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)

	// Order Module
	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// ==========================================
	// 5. Router & Middleware Pipeline
	// ==========================================
	mux := http.NewServeMux()

	auth := middleware.AuthMiddleware(cfg.JWTSecret)

	// ลงทะเบียน Routes
	userHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux, auth)
	orderHandler.RegisterRoutes(mux, auth)

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"env":      cfg.AppEnv,
			"database": "PostgreSQL (Connected)",
		})
	})

	// Wrap Global Middlewares: Recovery -> Logger -> Mux
	handlerWithMiddlewares := middleware.Recovery(middleware.RequestLogger(mux))

	// ==========================================
	// 6. ตั้งค่า HTTP Server (พร้อม Timeouts ป้องกัน DoS)
	// ==========================================
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handlerWithMiddlewares,
		ReadTimeout:  10 * time.Second, // ป้องกัน Slowloris Attack
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ==========================================
	// 7. รัน Server ใน Goroutine (Background)
	// ==========================================
	go func() {
		slog.Info("🚀 Server started successfully", "port", cfg.Port, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// ==========================================
	// 8. ดักฟัง OS Signals สำหรับ Graceful Shutdown
	// ==========================================
	quit := make(chan os.Signal, 1)
	// ดักจับสัญญาณ Ctrl+C (SIGINT) และ SIGTERM จาก Docker/K8s
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// บล็อกรอจนกว่าจะได้รับสัญญาณปิด
	sig := <-quit
	slog.Info("🛑 Shutdown signal received", "signal", sig.String())

	// ให้เวลา Request ที่ค้างอยู่ทำงานให้เสร็จสูงสุด 10 วินาที
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("✅ Server exited gracefully")
}
