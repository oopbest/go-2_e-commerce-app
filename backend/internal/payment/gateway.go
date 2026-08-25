package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type mockGateway struct{}

// NewMockGateway Constructor สำหรับสร้าง Mock Payment Gateway
func NewMockGateway() domain.PaymentGateway {
	return &mockGateway{}
}

// GenerateIntent จำลองการสร้าง Payment Transaction และออก QR / Payment URL
func (g *mockGateway) GenerateIntent(ctx context.Context, orderID int, amount float64, method string) (*domain.PaymentIntentResponse, error) {
	// สร้าง Transaction Reference จำลอง เช่น TXN_MOCK_1740001234_1
	txnRef := fmt.Sprintf("TXN_MOCK_%d_%d", time.Now().Unix(), orderID)

	response := &domain.PaymentIntentResponse{
		TransactionRef: txnRef,
		OrderID:        orderID,
		Amount:         amount,
		PaymentMethod:  method,
		Status:         "pending_payment",
	}

	switch method {
	case domain.PaymentMethodPromptPay:
		// จำลองข้อมูล EMVCo QR Code สำหรับ PromptPay
		response.QRCodeData = fmt.Sprintf("00020101021129370016A0000006770101115802TH5303764540%0.2f6304MOCK", amount)
	case domain.PaymentMethodCreditCard, domain.PaymentMethodTrueMoney:
		// จำลอง Payment Gateway Checkout URL
		response.PaymentURL = fmt.Sprintf("https://mock-payment-gateway.internal/pay/%s", txnRef)
	default:
		return nil, domain.ErrInvalidPaymentMethod
	}

	return response, nil
}

// VerifyTransaction จำลองการตรวจสอบความถูกต้องของยอดเงิน
func (g *mockGateway) VerifyTransaction(ctx context.Context, transactionRef string, expectedAmount float64) (bool, error) {
	if transactionRef == "" || expectedAmount <= 0 {
		return false, domain.ErrPaymentFailed
	}
	// จำลองว่าการชำระเงินผ่าน Gateway สำเร็จเสมอ
	return true, nil
}
