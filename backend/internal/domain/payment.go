package domain

import (
	"context"
	"errors"
	"time"
)

// ==========================================
// 1. Errors & Status Constants
// ==========================================

var (
	ErrPaymentNotFound      = errors.New("payment record not found")
	ErrOrderAlreadyPaid     = errors.New("order has already been paid")
	ErrOrderCancelled       = errors.New("cannot pay for a cancelled order")
	ErrInvalidPaymentAmount = errors.New("payment amount does not match order total")
	ErrInvalidPaymentMethod = errors.New("unsupported payment method")
	ErrPaymentFailed        = errors.New("payment transaction failed or was rejected")
)

// Payment Methods
const (
	PaymentMethodCreditCard   = "credit_card"
	PaymentMethodPromptPay    = "promptpay"
	PaymentMethodTrueMoney    = "truemoney"
	PaymentMethodBankTransfer = "bank_transfer"
)

// Payment Statuses
const (
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

// ==========================================
// 2. Entities & DTOs
// ==========================================

// Payment โครงสร้างข้อมูลประวัติการชำระเงินและใบเสร็จ
type Payment struct {
	ID             int       `json:"id"`
	OrderID        int       `json:"order_id"`
	Amount         float64   `json:"amount"`
	PaymentMethod  string    `json:"payment_method"`
	TransactionRef string    `json:"transaction_ref"`
	Status         string    `json:"status"`
	PaidAt         time.Time `json:"paid_at"`
}

// CreatePaymentIntentInput ข้อมูลที่รับเข้ามาเพื่อขอรับช่องทางชำระเงิน
type CreatePaymentIntentInput struct {
	OrderID       int    `json:"order_id"`
	PaymentMethod string `json:"payment_method"` // credit_card, promptpay, truemoney
}

// PaymentIntentResponse ข้อมูลช่องทางชำระเงินที่ระบบส่งกลับให้ลูกค้า (QR Code / URL)
type PaymentIntentResponse struct {
	TransactionRef string  `json:"transaction_ref"`
	OrderID        int     `json:"order_id"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentURL     string  `json:"payment_url,omitempty"`  // สำหรับบัตรเครดิต
	QRCodeData     string  `json:"qr_code_data,omitempty"` // สำหรับ PromptPay
	Status         string  `json:"status"`
}

// ConfirmPaymentInput ข้อมูลยืนยันการจ่ายเงิน (จาก Client หรือ Webhook)
type ConfirmPaymentInput struct {
	TransactionRef string  `json:"transaction_ref"`
	OrderID        int     `json:"order_id"`
	Amount         float64 `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
}

// ==========================================
// 3. Repository & Service Interfaces
// ==========================================

// PaymentRepository สัญญาณการเข้าถึงข้อมูล Payment ในฐานข้อมูล
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByOrderID(ctx context.Context, orderID int) (*Payment, error)
	GetByTransactionRef(ctx context.Context, ref string) (*Payment, error)
	// ConfirmOrderPaymentTx ทำการเปลี่ยน orders.status = 'paid' และบันทึกตาราง payments ใน Transaction เดียวกันแบบ Atomic
	ConfirmOrderPaymentTx(ctx context.Context, orderID int, payment *Payment) error
}

// PaymentGateway Interface กลางสำหรับคุยกับ Payment Provider (Mock, Stripe, Omise, 2C2P)
type PaymentGateway interface {
	GenerateIntent(ctx context.Context, orderID int, amount float64, method string) (*PaymentIntentResponse, error)
	VerifyTransaction(ctx context.Context, transactionRef string, expectedAmount float64) (bool, error)
}

// PaymentService Business Logic สำหรับระบบชำระเงิน
type PaymentService interface {
	CreatePaymentIntent(ctx context.Context, userID int, input CreatePaymentIntentInput) (*PaymentIntentResponse, error)
	ConfirmPayment(ctx context.Context, input ConfirmPaymentInput) (*Payment, error)
	GetPaymentByOrderID(ctx context.Context, userID, orderID int, userRole string) (*Payment, error)
}
