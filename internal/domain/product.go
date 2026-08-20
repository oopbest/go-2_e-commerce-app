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

// 2. Entity & DTOs
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type UpdateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// 3. Interfaces
type ProductRepository interface {
	FindAll() ([]Product, error)
	FindByID(id int) (*Product, error)
	Create(input CreateProductInput) (*Product, error)
	Update(id int, input UpdateProductInput) (*Product, error)
	Delete(id int) error
}

type ProductService interface {
	GetAllProducts() ([]Product, error)
	GetProductByID(id int) (*Product, error)
	CreateProduct(input CreateProductInput) (*Product, error)
	UpdateProduct(id int, input UpdateProductInput) (*Product, error)
	DeleteProduct(id int) error
}
