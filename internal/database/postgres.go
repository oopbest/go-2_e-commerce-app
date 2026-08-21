package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // Driver สำหรับ PostgreSQL (Import แบบ Blank Identifier _)
)

// Config โครงสร้างสำหรับเก็บข้อมูลการเชื่อมต่อ Database
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB สร้างและตั้งค่า Connection Pool ของ PostgreSQL
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// sql.Open จะยังไม่เชื่อมต่อ Network ทันที แต่เป็นการเตรียม Pool
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// ตั้งค่า Connection Pool (Best Practice สำหรับ Production)
	db.SetMaxOpenConns(25)                 // จำนวน Connection สูงสุดที่เปิดพร้อมกันได้
	db.SetMaxIdleConns(25)                 // จำนวน Connection ที่พักรอไว้ใน Pool
	db.SetConnMaxLifetime(5 * time.Minute) // อายุสูงสุดของแต่ละ Connection ก่อนถูกปิดแล้วสร้างใหม่

	// ทดสอบ Ping ไปยัง Database จริงด้วย Context Timeout 5 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully!")
	return db, nil
}
