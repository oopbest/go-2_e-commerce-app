package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient สร้างและตรวจสอบการเชื่อมต่อ Redis Client
func NewRedisClient(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,               // ใช้ Default DB (Database 0)
		PoolSize:     20,              // จำนวน Connection สูงสุดใน Pool
		MinIdleConns: 5,               // จำนวน Connection พักรอไว้ใน Pool
		DialTimeout:  5 * time.Second, // Timeout ในการต่อ Network
		ReadTimeout:  3 * time.Second, // Timeout ในการอ่านข้อมูล
		WriteTimeout: 3 * time.Second, // Timeout ในการเขียนข้อมูล
	})

	// ทดสอบ Ping ไปยัง Redis Server จริง
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis at %s: %w", addr, err)
	}

	log.Println("✅ Connected to Redis successfully!")
	return rdb, nil
}
