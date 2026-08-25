package payment

import (
	"context"
	"errors"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type service struct {
	repo      domain.PaymentRepository
	orderRepo domain.OrderRepository
	gateway   domain.PaymentGateway
}

// NewService Constructor สำหรับสร้าง Payment Service
func NewService(repo domain.PaymentRepository, orderRepo domain.OrderRepository, gateway domain.PaymentGateway) domain.PaymentService {
	return &service{
		repo:      repo,
		orderRepo: orderRepo,
		gateway:   gateway,
	}
}

// CreatePaymentIntent สร้างช่องทางชำระเงิน (QR/URL) ให้กับลูกค้า
func (s *service) CreatePaymentIntent(ctx context.Context, userID int, input domain.CreatePaymentIntentInput) (*domain.PaymentIntentResponse, error) {
	if input.OrderID <= 0 {
		return nil, errors.New("invalid order ID")
	}

	// 1. ดึงข้อมูล Order และตรวจสอบสิทธิ์การเป็นเจ้าของบิล
	order, err := s.orderRepo.FindOrderByID(ctx, input.OrderID, userID)
	if err != nil {
		return nil, err
	}

	// 2. ตรวจสอบสถานะบิล
	if order.Status == "paid" {
		return nil, domain.ErrOrderAlreadyPaid
	}
	if order.Status == "cancelled" {
		return nil, domain.ErrOrderCancelled
	}

	// 3. เรียก Payment Gateway เพื่อสร้าง Payment Intent (QR Code / URL)
	return s.gateway.GenerateIntent(ctx, order.ID, order.TotalAmount, input.PaymentMethod)
}

// ConfirmPayment ยืนยันการชำระเงิน เปลี่ยนสถานะ Order เป็น 'paid' และออกใบเสร็จ
func (s *service) ConfirmPayment(ctx context.Context, input domain.ConfirmPaymentInput) (*domain.Payment, error) {
	if input.OrderID <= 0 || input.TransactionRef == "" || input.Amount <= 0 {
		return nil, errors.New("invalid payment confirmation payload")
	}

	// 1. ตรวจสอบความถูกต้องกับ Gateway
	valid, err := s.gateway.VerifyTransaction(ctx, input.TransactionRef, input.Amount)
	if err != nil || !valid {
		return nil, domain.ErrPaymentFailed
	}

	// 2. สร้างโครงสร้าง Payment Record
	payment := &domain.Payment{
		OrderID:        input.OrderID,
		Amount:         input.Amount,
		PaymentMethod:  input.PaymentMethod,
		TransactionRef: input.TransactionRef,
		Status:         domain.PaymentStatusCompleted,
	}

	// 3. ทำ Atomic Transaction เปลี่ยน orders.status = 'paid' และบันทึก payments
	if err := s.repo.ConfirmOrderPaymentTx(ctx, input.OrderID, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// GetPaymentByOrderID ดูข้อมูลประวัติการชำระเงินและใบเสร็จ
func (s *service) GetPaymentByOrderID(ctx context.Context, userID, orderID int, userRole string) (*domain.Payment, error) {
	// 1. ตรวจสอบสิทธิ์การเข้าถึงบิล
	queryUserID := userID
	if userRole == domain.RoleAdmin {
		queryUserID = 0 // Admin ดูได้ทุกบิล
	}

	if _, err := s.orderRepo.FindOrderByID(ctx, orderID, queryUserID); err != nil {
		return nil, err
	}

	return s.repo.GetByOrderID(ctx, orderID)
}
