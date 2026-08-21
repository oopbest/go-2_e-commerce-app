package order

import (
	"context"
	"errors"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type service struct {
	repo domain.OrderRepository
}

// NewService Constructor สำหรับสร้าง Order Service
func NewService(repo domain.OrderRepository) domain.OrderService {
	return &service{
		repo: repo,
	}
}

// Checkout ตรวจสอบความถูกต้องและสร้างคำสั่งซื้อ
func (s *service) Checkout(ctx context.Context, userID int, input domain.CheckoutInput) (*domain.Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// 1. ตรวจสอบว่ามีรายการสินค้าในตะกร้าหรือไม่
	if len(input.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	// 2. ตรวจสอบความถูกต้องของแต่ละรายการ
	for _, item := range input.Items {
		if item.ProductID <= 0 {
			return nil, errors.New("invalid product ID")
		}
		if item.Quantity <= 0 {
			return nil, domain.ErrInvalidQuantity
		}
	}

	// 3. ส่งต่อไปยัง Repository เพื่อทำ Database Transaction & ตัดสต็อก
	return s.repo.CreateOrder(ctx, userID, input.Items)
}

// GetOrderByID ดึงข้อมูลคำสั่งซื้อ (พร้อมตรวจสอบสิทธิ์ความเป็นเจ้าของ)
func (s *service) GetOrderByID(ctx context.Context, orderID, userID int, userRole string) (*domain.Order, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}

	// ถ้าเป็น Admin ให้ดูบิลของใครก็ได้ (ส่ง userID = 0)
	queryUserID := userID
	if userRole == "admin" {
		queryUserID = 0
	}

	return s.repo.FindOrderByID(ctx, orderID, queryUserID)
}

// GetUserOrders ดึงประวัติคำสั่งซื้อของลูกค้า
func (s *service) GetUserOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindOrdersByUserID(ctx, userID)
}

// GetAllOrders ดึงคำสั่งซื้อทั้งหมดในระบบ (สำหรับ Admin)
func (s *service) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	return s.repo.FindAllOrders(ctx)
}
