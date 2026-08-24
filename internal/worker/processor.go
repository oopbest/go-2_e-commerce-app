package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

// TaskProcessor Interface สำหรับประมวลผลงาน
type TaskProcessor interface {
	ProcessTaskOrderCreatedEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskOrderTimeoutCheck(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor ตัวประมวลผลที่เชื่อมต่อกับ DB และ Redis
type RedisTaskProcessor struct {
	orderRepo domain.OrderRepository
	rdb       *redis.Client
}

func NewRedisTaskProcessor(orderRepo domain.OrderRepository, rdb *redis.Client) *RedisTaskProcessor {
	return &RedisTaskProcessor{
		orderRepo: orderRepo,
		rdb:       rdb,
	}
}

// ProcessTaskOrderCreatedEmail ประมวลผลการส่ง Email ยืนยันการสั่งซื้อ
func (p *RedisTaskProcessor) ProcessTaskOrderCreatedEmail(ctx context.Context, task *asynq.Task) error {
	var payload OrderCreatedEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", asynq.SkipRetry)
	}

	slog.Info("📧 [EMAIL WORKER] Processing order confirmation email...",
		"order_id", payload.OrderID,
		"to_email", payload.UserEmail,
		"total_amount", payload.TotalAmount,
	)

	// จำลองการส่ง Email สำเร็จ
	slog.Info("✅ [EMAIL WORKER] Order confirmation email delivered successfully!",
		"order_id", payload.OrderID,
		"to_email", payload.UserEmail,
	)

	return nil
}

// ProcessTaskOrderTimeoutCheck ประมวลผลการตรวจบิลหมดอายุและคืนสต็อกสินค้า
func (p *RedisTaskProcessor) ProcessTaskOrderTimeoutCheck(ctx context.Context, task *asynq.Task) error {
	var payload OrderTimeoutPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal timeout payload: %w", asynq.SkipRetry)
	}

	slog.Info("⏰ [TIMEOUT WORKER] Checking unpaid order status...", "order_id", payload.OrderID)

	// รันการคืนสต็อกใน DB Transaction
	err := p.orderRepo.CancelOrderAndRestoreStock(ctx, payload.OrderID)
	if err != nil {
		slog.Error("❌ [TIMEOUT WORKER] Failed to cancel order and restore stock", "order_id", payload.OrderID, "error", err)
		return err
	}

	// ล้างแคช Redis ของ Product เพื่อให้แสดงสต็อกใหม่ที่คืนแล้วทันที
	if p.rdb != nil {
		_ = p.rdb.Del(ctx, "products:all").Err()
	}

	slog.Warn("🚨 [TIMEOUT WORKER] Order timeout processed: Cancelled unpaid order & restored product stock successfully!",
		"order_id", payload.OrderID,
	)

	return nil
}
