package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/oopbest/ecommerce-app/internal/config"
	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/order"
	"github.com/oopbest/ecommerce-app/internal/worker"
)

func main() {
	// 1. โหลด Config
	cfg := config.Load()

	// 2. ตั้งค่า Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	slog.Info("⚙️ Starting Background Worker Server...")

	// 3. เชื่อมต่อ PostgreSQL
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
		slog.Error("Failed to initialize database for worker", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. เชื่อมต่อ Redis
	rdb, err := database.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		slog.Error("Failed to initialize Redis for worker", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// 5. สร้าง Task Processor (Consumer)
	orderRepo := order.NewRepository(db, rdb)
	processor := worker.NewRedisTaskProcessor(orderRepo, rdb)

	// 6. ตั้งค่า Asynq Server (Worker Pool)
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10, // รับและประมวลผลพร้อมกันได้ 10 Goroutines
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// 7. ผูก Route ของ Task เข้ากับ Processor Handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeOrderCreatedEmail, processor.ProcessTaskOrderCreatedEmail)
	mux.HandleFunc(worker.TypeOrderTimeoutCheck, processor.ProcessTaskOrderTimeoutCheck)

	// 8. รัน Worker Server ใน Goroutine
	go func() {
		slog.Info("🚀 Background Worker is running and listening for tasks...")
		if err := srv.Run(mux); err != nil {
			slog.Error("Worker server stopped with error", "error", err)
		}
	}()

	// 9. Graceful Shutdown สำหรับ Worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	slog.Info("🛑 Worker shutdown signal received", "signal", sig.String())

	srv.Shutdown()
	slog.Info("✅ Worker server exited gracefully")
}
