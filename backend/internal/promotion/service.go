package promotion

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type service struct {
	repo domain.PromotionRepository
}

// NewService Constructor สำหรับสร้าง Promotion Service
func NewService(repo domain.PromotionRepository) domain.PromotionService {
	return &service{repo: repo}
}

// GetActivePromotions ดึงคูปองทั้งหมดที่ใช้งานได้ในปัจจุบัน
func (s *service) GetActivePromotions(ctx context.Context) ([]domain.Promotion, error) {
	return s.repo.FindAllActive(ctx)
}

// ValidateCoupon ตรวจสอบความถูกต้องของโค้ดคูปอง และคำนวณยอดส่วนลดสุทธิ
func (s *service) ValidateCoupon(ctx context.Context, code string, subtotal float64) (*domain.ValidationResult, error) {
	cleanCode := strings.TrimSpace(code)
	if cleanCode == "" {
		return nil, errors.New("coupon code is required")
	}
	if subtotal <= 0 {
		return nil, errors.New("subtotal must be greater than 0")
	}

	// 1. ค้นหาคูปองจาก Database
	promo, err := s.repo.FindByCode(ctx, cleanCode)
	if err != nil {
		return nil, err
	}

	// 2. เรียกใช้ Pure Domain Method เพื่อคำนวณส่วนลดตามกฎธุรกิจ
	discount, err := promo.CalculateDiscount(subtotal)
	if err != nil {
		return nil, err
	}

	finalTotal := math.Max(0, subtotal-discount)
	finalTotal = math.Round(finalTotal*100) / 100

	return &domain.ValidationResult{
		Promotion:      promo,
		Subtotal:       subtotal,
		DiscountAmount: discount,
		FinalTotal:     finalTotal,
	}, nil
}
