package domain

import (
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

// 3. Interfaces
type ProductRepository interface {
	FindAll() ([]Product, error)
	FindByID(id int) (*Product, error)
	FindAllBrands() ([]Brand, error) // 👈 ดึงรายชื่อแบรนด์ทั้งหมด
	Create(input CreateProductInput) (*Product, error)
	Update(id int, input UpdateProductInput) (*Product, error)
	Delete(id int) error
}

type ProductService interface {
	GetAllProducts() ([]Product, error)
	GetProductByID(id int) (*Product, error)
	GetAllBrands() ([]Brand, error) // 👈 ดึงรายชื่อแบรนด์ทั้งหมด
	CreateProduct(input CreateProductInput) (*Product, error)
	UpdateProduct(id int, input UpdateProductInput) (*Product, error)
	DeleteProduct(id int) error
}
