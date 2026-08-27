package product

import (
	"errors"
	"strings"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

// service struct เก็บ dependency คือ repo ผ่าน Interface
type service struct {
	repo domain.ProductRepository
}

// NewService Constructor function สำหรับสร้าง Service
func NewService(repo domain.ProductRepository) domain.ProductService {
	return &service{
		repo: repo,
	}
}

// GetAllProducts ดึงสินค้าทั้งหมด
func (s *service) GetAllProducts() ([]domain.Product, error) {
	return s.repo.FindAll()
}

// GetProductByID ดึงสินค้าตาม ID พร้อมตรวจสอบความถูกต้องของ ID
func (s *service) GetProductByID(id int) (*domain.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}
	return s.repo.FindByID(id)
}

// CreateProduct สร้างสินค้าใหม่พร้อมตรวจสอบเงื่อนไข (Business Validation)
func (s *service) CreateProduct(input domain.CreateProductInput) (*domain.Product, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("product name is required")
	}
	if input.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	if input.Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	return s.repo.Create(input)
}

// UpdateProduct แก้ไขสินค้าพร้อมตรวจสอบเงื่อนไข
func (s *service) UpdateProduct(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("product name is required")
	}
	if input.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	if input.Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	return s.repo.Update(id, input)
}

// DeleteProduct ลบสินค้า
func (s *service) DeleteProduct(id int) error {
	if id <= 0 {
		return errors.New("invalid product ID")
	}
	return s.repo.Delete(id)
}

// GetAllBrands ดึงรายชื่อแบรนด์ทั้งหมด
func (s *service) GetAllBrands() ([]domain.Brand, error) {
	return s.repo.FindAllBrands()
}
