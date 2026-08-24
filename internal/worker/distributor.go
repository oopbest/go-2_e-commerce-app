package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// TaskDistributor Interface สำหรับส่งงานเข้าคิว
type TaskDistributor interface {
	DistributeTaskOrderCreatedEmail(ctx context.Context, payload OrderCreatedEmailPayload, opts ...asynq.Option) error
	DistributeTaskOrderTimeoutCheck(ctx context.Context, payload OrderTimeoutPayload, opts ...asynq.Option) error
}

// RedisTaskDistributor Implementation ที่เชื่อมต่อกับ Redis
type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisConnOpt) *RedisTaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{client: client}
}

// DistributeTaskOrderCreatedEmail ส่งงานส่ง Email เข้าคิว (ทำทันที)
func (d *RedisTaskDistributor) DistributeTaskOrderCreatedEmail(ctx context.Context, payload OrderCreatedEmailPayload, opts ...asynq.Option) error {
	task, err := NewOrderCreatedEmailTask(payload)
	if err != nil {
		return err
	}

	// กำหนดให้ retry ได้สูงสุด 3 ครั้งถ้าส่งไม่ผ่าน
	defaultOpts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(10 * time.Second),
	}
	allOpts := append(defaultOpts, opts...)

	info, err := d.client.EnqueueContext(ctx, task, allOpts...)
	if err != nil {
		slog.Error("Failed to enqueue email task", "error", err)
		return err
	}

	slog.Info("📬 Enqueued email task", "type", task.Type(), "queue", info.Queue, "max_retry", info.MaxRetry)
	return nil
}

// DistributeTaskOrderTimeoutCheck ส่งงานตรวจคำสั่งซื้อหมดอายุเข้าคิว (หน่วงเวลาล่วงหน้า)
func (d *RedisTaskDistributor) DistributeTaskOrderTimeoutCheck(ctx context.Context, payload OrderTimeoutPayload, opts ...asynq.Option) error {
	task, err := NewOrderTimeoutCheckTask(payload)
	if err != nil {
		return err
	}

	defaultOpts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(10 * time.Second),
	}
	allOpts := append(defaultOpts, opts...)

	info, err := d.client.EnqueueContext(ctx, task, allOpts...)
	if err != nil {
		slog.Error("Failed to enqueue order timeout task", "error", err)
		return err
	}

	slog.Info("⏰ Enqueued delayed timeout task", "type", task.Type(), "queue", info.Queue)
	return nil
}
