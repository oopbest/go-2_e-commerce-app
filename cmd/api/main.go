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

	_ "github.com/oopbest/ecommerce-app/docs" // Import Docs ที่จะถูกสร้างโดย swag init
	"github.com/oopbest/ecommerce-app/internal/config"
	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/internal/order"
	"github.com/oopbest/ecommerce-app/internal/product"
	"github.com/oopbest/ecommerce-app/internal/user"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/hibiken/asynq"
	"github.com/oopbest/ecommerce-app/internal/worker"
)

// @title           Go E-Commerce REST API
// @version         1.0
// @description     High-performance production-ready E-Commerce Backend with JWT Auth, PostgreSQL, and Redis Cache.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/oopbest/go-2_e-commerce-app

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOi..."

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
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
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
	// 4. เชื่อมต่อ Redis In-Memory Cache
	// ==========================================
	rdb, err := database.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = rdb.Close()
		slog.Info("🔒 Redis connection closed")
	}()

	// ==========================================
	// 5. Dependency Injection: Modules & Task Distributor
	// ==========================================
	// ⚡ สร้าง Task Distributor สำหรับส่งงานเข้า Redis Queue
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, cfg.JWTSecret)
	userHandler := user.NewHandler(userService)
	productPostgresRepo := product.NewPostgresRepository(db)
	productRepo := product.NewCachedRepository(productPostgresRepo, rdb, 5*time.Minute)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)
	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo, taskDistributor) // 👈 ฉีด taskDistributor เข้าไป
	orderHandler := order.NewHandler(orderService)

	// ==========================================
	// 6. Router & Middleware Pipeline
	// ==========================================
	mux := http.NewServeMux()

	auth := middleware.AuthMiddleware(cfg.JWTSecret)

	// ลงทะเบียน Routes
	userHandler.RegisterRoutes(mux)
	productHandler.RegisterRoutes(mux, auth)
	orderHandler.RegisterRoutes(mux, auth)

	// 📖 Swagger UI Endpoint (ใหม่!)
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"env":      cfg.AppEnv,
			"database": "PostgreSQL (Connected)",
			"redis":    "Redis (Connected)",
		})
	})

	// Wrap Global Middlewares
	handlerWithMiddlewares := middleware.Recovery(middleware.RequestLogger(mux))

	// ==========================================
	// 7. ตั้งค่า HTTP Server
	// ==========================================
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handlerWithMiddlewares,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ==========================================
	// 8. รัน Server ใน Goroutine
	// ==========================================
	go func() {
		slog.Info("🚀 Server started successfully", "port", cfg.Port, "env", cfg.AppEnv)
		slog.Info("📖 Swagger UI available at", "url", "http://localhost:"+cfg.Port+"/swagger/index.html")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// ==========================================
	// 9. ดักฟัง OS Signals สำหรับ Graceful Shutdown
	// ==========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	slog.Info("🛑 Shutdown signal received", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("✅ Server exited gracefully")
}
