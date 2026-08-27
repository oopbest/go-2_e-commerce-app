package domain

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"
)

// 1. Domain Errors สำหรับคูปองส่วนลด
var (
	ErrPromotionNotFound       = errors.New("promotion code not found")
	ErrPromotionExpired        = errors.New("promotion code has expired")
	ErrPromotionNotStarted     = errors.New("promotion campaign has not started yet")
	ErrPromotionQuotaExceeded  = errors.New("promotion quota has been completely used")
	ErrPromotionMinSpendNotMet = errors.New("order subtotal does not meet the minimum spend requirement")
	ErrPromotionInactive       = errors.New("promotion is currently inactive")
)

// 2. Promotion Entity
type Promotion struct {
	ID            int       `json:"id"`
	Code          string    `json:"code"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"`          // "fixed" หรือ "percentage"
	DiscountValue float64   `json:"discount_value"`         // เช่น 100 บาท หรือ 15%
	MinSpend      float64   `json:"min_spend"`              // ยอดซื้อขั้นต่ำ
	MaxDiscount   *float64  `json:"max_discount,omitempty"` // เพดานลดสูงสุด (กรณีคิดเป็น %)
	TotalQuota    int       `json:"total_quota"`
	UsedCount     int       `json:"used_count"`
	StartsAt      time.Time `json:"starts_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// ValidationResult ผลลัพธ์จากการคำนวณส่วนลด
type ValidationResult struct {
	Promotion      *Promotion `json:"promotion"`
	Subtotal       float64    `json:"subtotal"`
	DiscountAmount float64    `json:"discount_amount"`
	FinalTotal     float64    `json:"final_total"`
}

// CalculateDiscount ฟังก์ชันคำนวณส่วนลดตามกฎเกณฑ์ธุรกิจ (Pure Domain Method)
func (p *Promotion) CalculateDiscount(subtotal float64) (float64, error) {
	now := time.Now()

	// 1. ตรวจสอบสถานะการเปิดใช้งาน
	if !p.IsActive {
		return 0, ErrPromotionInactive
	}

	// 2. ตรวจสอบช่วงเวลา
	if now.Before(p.StartsAt) {
		return 0, ErrPromotionNotStarted
	}
	if now.After(p.ExpiresAt) {
		return 0, ErrPromotionExpired
	}

	// 3. ตรวจสอบโควตาสิทธิ์คงเหลือ
	if p.UsedCount >= p.TotalQuota {
		return 0, ErrPromotionQuotaExceeded
	}

	// 4. ตรวจสอบยอดซื้อขั้นต่ำ
	if subtotal < p.MinSpend {
		return 0, ErrPromotionMinSpendNotMet
	}

	// 5. คำนวณส่วนลดตามประเภท
	var discount float64
	if p.DiscountType == "fixed" {
		discount = p.DiscountValue
	} else if p.DiscountType == "percentage" {
		discount = subtotal * (p.DiscountValue / 100.0)
		// ถ้ามีการตั้งเพดานลดสูงสุด (MaxDiscount) ให้ปัดไม่เกินเพดาน
		if p.MaxDiscount != nil && discount > *p.MaxDiscount {
			discount = *p.MaxDiscount
		}
	}

	// ปัดเศษทศนิยม 2 ตำแหน่ง
	discount = math.Round(discount*100) / 100

	// ส่วนลดต้องไม่เกินยอดรวมสินค้า
	if discount > subtotal {
		discount = subtotal
	}

	return discount, nil
}

// 3. Interfaces
type PromotionRepository interface {
	FindAllActive(ctx context.Context) ([]Promotion, error)
	FindByCode(ctx context.Context, code string) (*Promotion, error)
	IncrementUsageTx(ctx context.Context, tx *sql.Tx, promotionID int) error
}

type PromotionService interface {
	GetActivePromotions(ctx context.Context) ([]Promotion, error)
	ValidateCoupon(ctx context.Context, code string, subtotal float64) (*ValidationResult, error)
}
