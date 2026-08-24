package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// ==========================================
// 1. ชนิดของ Background Tasks
// ==========================================
const (
	// TypeOrderCreatedEmail: งานส่ง Email ยืนยันการสั่งซื้อ (ทำทันที)
	TypeOrderCreatedEmail = "order:email:confirmation"

	// TypeOrderTimeoutCheck: งานตรวจจับคำสั่งซื้อหมดอายุ (ทำแบบ Delay/หน่วงเวลา)
	TypeOrderTimeoutCheck = "order:timeout:check"
)

// ==========================================
// 2. ข้อมูลที่แนบไปกับแต่ละ Task (Payloads)
// ==========================================

// OrderCreatedEmailPayload ข้อมูลสำหรับส่ง Email
type OrderCreatedEmailPayload struct {
	OrderID     int     `json:"order_id"`
	UserID      int     `json:"user_id"`
	UserEmail   string  `json:"user_email"`
	TotalAmount float64 `json:"total_amount"`
}

// OrderTimeoutPayload ข้อมูลสำหรับตรวจสอบคำสั่งซื้อหมดอายุ
type OrderTimeoutPayload struct {
	OrderID int `json:"order_id"`
}

// ==========================================
// 3. Helper Functions ในการสร้าง Asynq Task
// ==========================================

// NewOrderCreatedEmailTask สร้าง Task สำหรับส่ง Email
func NewOrderCreatedEmailTask(payload OrderCreatedEmailPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeOrderCreatedEmail, data), nil
}

// NewOrderTimeoutCheckTask สร้าง Task สำหรับตรวจบิลหมดอายุ
func NewOrderTimeoutCheckTask(payload OrderTimeoutPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeOrderTimeoutCheck, data), nil
}
