package payment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type repository struct {
	db *sql.DB
}

// NewRepository Constructor สำหรับสร้าง Payment Repository
func NewRepository(db *sql.DB) domain.PaymentRepository {
	return &repository{db: db}
}

// Create บันทึกข้อมูลการชำระเงินลงตาราง payments
func (r *repository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (order_id, amount, payment_method, transaction_ref, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, paid_at
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		payment.OrderID,
		payment.Amount,
		payment.PaymentMethod,
		payment.TransactionRef,
		payment.Status,
	).Scan(&payment.ID, &payment.PaidAt)
}

// GetByOrderID ดึงข้อมูลการชำระเงินจาก Order ID
func (r *repository) GetByOrderID(ctx context.Context, orderID int) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, payment_method, transaction_ref, status, paid_at
		FROM payments
		WHERE order_id = $1
		ORDER BY id DESC
		LIMIT 1
	`
	var p domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&p.ID,
		&p.OrderID,
		&p.Amount,
		&p.PaymentMethod,
		&p.TransactionRef,
		&p.Status,
		&p.PaidAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByTransactionRef ดึงข้อมูลการชำระเงินจาก Transaction Reference
func (r *repository) GetByTransactionRef(ctx context.Context, ref string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, amount, payment_method, transaction_ref, status, paid_at
		FROM payments
		WHERE transaction_ref = $1
	`
	var p domain.Payment
	err := r.db.QueryRowContext(ctx, query, ref).Scan(
		&p.ID,
		&p.OrderID,
		&p.Amount,
		&p.PaymentMethod,
		&p.TransactionRef,
		&p.Status,
		&p.PaidAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ConfirmOrderPaymentTx ทำการเปลี่ยน orders.status = 'paid' และบันทึกตาราง payments ใน Transaction เดียวกันแบบ Atomic
func (r *repository) ConfirmOrderPaymentTx(ctx context.Context, orderID int, payment *domain.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. ล็อกแถว Order และตรวจสอบสถานะปัจจุบันด้วย Pessimistic Lock
	var currentStatus string
	var totalAmount float64
	checkQuery := `SELECT status, total_amount FROM orders WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, checkQuery, orderID).Scan(&currentStatus, &totalAmount)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	if err != nil {
		return err
	}

	// 2. ตรวจสอบความถูกต้องของสถานะและยอดเงิน
	if currentStatus == "paid" {
		return domain.ErrOrderAlreadyPaid
	}
	if currentStatus == "cancelled" {
		return domain.ErrOrderCancelled
	}
	if payment.Amount != totalAmount {
		return domain.ErrInvalidPaymentAmount
	}

	// 3. ปรับสถานะคำสั่งซื้อเป็น 'paid'
	updateOrderQuery := `UPDATE orders SET status = 'paid' WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateOrderQuery, orderID); err != nil {
		return err
	}

	// 4. บันทึกประวัติลงตาราง payments
	insertPaymentQuery := `
		INSERT INTO payments (order_id, amount, payment_method, transaction_ref, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, paid_at
	`
	err = tx.QueryRowContext(
		ctx,
		insertPaymentQuery,
		payment.OrderID,
		payment.Amount,
		payment.PaymentMethod,
		payment.TransactionRef,
		payment.Status,
	).Scan(&payment.ID, &payment.PaidAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}
