package flashsale

import (
	"context"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type service struct {
	repo domain.FlashSaleRepository
}

// NewService Constructor สำหรับสร้าง Flash Sale Service
func NewService(repo domain.FlashSaleRepository) domain.FlashSaleService {
	return &service{repo: repo}
}

// GetCurrentActive ดึงแคมเปญแฟลชเซลล์ที่กำลังจัดอยู่ในปัจจุบัน
func (s *service) GetCurrentActive(ctx context.Context) (*domain.FlashSale, error) {
	return s.repo.FindCurrentActive(ctx)
}
