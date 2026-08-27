package domain

import (
	"context"
	"time"
)

// FlashSaleItem รายการสินค้าที่เข้าร่วม Flash Sale
type FlashSaleItem struct {
	ID            int       `json:"id"`
	FlashSaleID   int       `json:"flash_sale_id"`
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ImageURL      string    `json:"image_url"`
	OriginalPrice float64   `json:"original_price"`
	FlashPrice    float64   `json:"flash_price"`
	FlashStock    int       `json:"flash_stock"`
	SoldCount     int       `json:"sold_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// FlashSale แคมเปญแฟลชเซลล์
type FlashSale struct {
	ID          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	StartsAt    time.Time       `json:"starts_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	IsActive    bool            `json:"is_active"`
	Items       []FlashSaleItem `json:"items,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// Interfaces
type FlashSaleRepository interface {
	FindCurrentActive(ctx context.Context) (*FlashSale, error)
}

type FlashSaleService interface {
	GetCurrentActive(ctx context.Context) (*FlashSale, error)
}
