package domain

import (
	"context"
	"errors"
	"time"
)

// 1. Domain Errors
var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidInput    = errors.New("invalid product input")
)

// Brand ข้อมูลแบรนด์สินค้า
type Brand struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	LogoURL     string    `json:"logo_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// 2. Entity & DTOs
type Product struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Stock        int            `json:"stock"`
	CategoryID   *int           `json:"category_id,omitempty"`
	CategoryName string         `json:"category_name,omitempty"`
	BrandID      *int           `json:"brand_id,omitempty"`
	BrandName    string         `json:"brand_name,omitempty"`
	ImageURL     string         `json:"image_url"`
	SKU          string         `json:"sku"`
	Specs        map[string]any `json:"specs,omitempty"`
	Rating       float64        `json:"rating"`
	ReviewsCount int            `json:"reviews_count"`
	CreatedAt    time.Time      `json:"created_at"`
}

type CreateProductInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	CategoryID  *int           `json:"category_id,omitempty"`
	BrandID     *int           `json:"brand_id,omitempty"`
	ImageURL    string         `json:"image_url"`
	SKU         string         `json:"sku"`
	Specs       map[string]any `json:"specs,omitempty"`
}

type UpdateProductInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	CategoryID  *int           `json:"category_id,omitempty"`
	BrandID     *int           `json:"brand_id,omitempty"`
	ImageURL    string         `json:"image_url"`
	SKU         string         `json:"sku"`
	Specs       map[string]any `json:"specs,omitempty"`
}

// ProductFilter เงื่อนไขค้นหา กรอง และเรียงลำดับสินค้า (Phase C)
type ProductFilter struct {
	Search      string   `json:"search"`        // ค้นหาจากชื่อหรือคำอธิบายสินค้า
	CategoryID  *int     `json:"category_id"`   // กรองตามหมวดหมู่
	BrandID     *int     `json:"brand_id"`      // กรองตามแบรนด์
	MinPrice    *float64 `json:"min_price"`     // กรองราคาต่ำสุด
	MaxPrice    *float64 `json:"max_price"`     // กรองราคาสูงสุด
	InStockOnly bool     `json:"in_stock_only"` // กรองเฉพาะสินค้าที่มีสต็อก (> 0)
	SortBy      string   `json:"sort_by"`       // "price_asc", "price_desc", "rating", "newest"
	Page        int      `json:"page"`          // หน้าที่ต้องการดึง (default: 1)
	Limit       int      `json:"limit"`         // จำนวนต่อหน้า (default: 20)
}

// ProductListResponse ผลลัพธ์รายการสินค้าพร้อมข้อมูลการแบ่งหน้า
type ProductListResponse struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

// 3. Interfaces
type ProductRepository interface {
	FindAll() ([]Product, error)
	FindWithFilter(ctx context.Context, filter ProductFilter) (*ProductListResponse, error) // 👈 เพิ่มเมธอดค้นหาพร้อมตัวกรอง
	FindByID(id int) (*Product, error)
	FindAllBrands() ([]Brand, error)
	Create(input CreateProductInput) (*Product, error)
	Update(id int, input UpdateProductInput) (*Product, error)
	Delete(id int) error
}

type ProductService interface {
	GetAllProducts() ([]Product, error)
	GetProductsWithFilter(ctx context.Context, filter ProductFilter) (*ProductListResponse, error) // 👈 เพิ่มเมธอดค้นหาพร้อมตัวกรอง
	GetProductByID(id int) (*Product, error)
	GetAllBrands() ([]Brand, error)
	CreateProduct(input CreateProductInput) (*Product, error)
	UpdateProduct(id int, input UpdateProductInput) (*Product, error)
	DeleteProduct(id int) error
}
