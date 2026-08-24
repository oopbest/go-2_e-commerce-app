package order

import (
	"context"
	"errors"
	"time"

	"github.com/hibiken/asynq"
	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/internal/worker"
	"github.com/oopbest/ecommerce-app/pkg/security"
)

type service struct {
	repo        domain.OrderRepository
	distributor worker.TaskDistributor
}

// NewService Constructor สำหรับสร้าง Order Service พร้อมฉีด TaskDistributor
func NewService(repo domain.OrderRepository, distributor worker.TaskDistributor) domain.OrderService {
	return &service{
		repo:        repo,
		distributor: distributor,
	}
}

// Checkout ตรวจสอบความถูกต้องและสร้างคำสั่งซื้อ พร้อมส่งงานเข้า Background Task
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
	createdOrder, err := s.repo.CreateOrder(ctx, userID, input.Items)
	if err != nil {
		return nil, err
	}

	// ==========================================
	// 🚀 Asynchronous Tasks: ส่งงานเข้า Redis Queue ทันทีหลัง Commit
	// ==========================================
	if s.distributor != nil && createdOrder != nil {
		// 📧 ดึง Email จริงของผู้ใช้จาก JWT Claims ที่อยู่ใน Context (ถ้ามี)
		userEmail := "customer@example.com"
		if claims, ok := ctx.Value(middleware.UserContextKey).(*security.CustomClaims); ok && claims != nil {
			userEmail = claims.Email
		}

		// 1. ส่ง Email ยืนยันการสั่งซื้อทันที
		emailPayload := worker.OrderCreatedEmailPayload{
			OrderID:     createdOrder.ID,
			UserID:      userID,
			UserEmail:   userEmail, // 👈 ใช้ Email จริงของผู้ใช้งาน
			TotalAmount: createdOrder.TotalAmount,
		}
		_ = s.distributor.DistributeTaskOrderCreatedEmail(ctx, emailPayload)

		// 2. ตั้งเวลาตรวจบิลหมดอายุล่วงหน้า (สำหรับ Demo ตั้งไว้ 1 นาที! เพื่อดูผลลัพธ์ได้ทันที)
		timeoutPayload := worker.OrderTimeoutPayload{
			OrderID: createdOrder.ID,
		}
		_ = s.distributor.DistributeTaskOrderTimeoutCheck(
			ctx,
			timeoutPayload,
			asynq.ProcessIn(1*time.Minute), // หน่วงเวลา 1 นาที
		)
	}

	return createdOrder, nil
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
