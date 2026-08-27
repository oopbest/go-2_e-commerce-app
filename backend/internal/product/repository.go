package product

import (
	"sync"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

// inMemoryRepository struct สำหรับเก็บข้อมูลสินค้าใน RAM
type inMemoryRepository struct {
	mu       sync.RWMutex
	products []domain.Product
	nextID   int
}

// NewInMemoryRepository Constructor function
// คืนค่าเป็น domain.ProductRepository (Interface)
func NewInMemoryRepository() domain.ProductRepository {
	return &inMemoryRepository{
		products: []domain.Product{
			{ID: 1, Name: "Mechanical Keyboard", Description: "RGB Hot-swappable", Price: 2590.00, Stock: 15, CreatedAt: time.Now()},
			{ID: 2, Name: "Wireless Mouse", Description: "Ergonomic 2.4GHz", Price: 1290.00, Stock: 30, CreatedAt: time.Now()},
		},
		nextID: 3,
	}
}

// FindAll ดึงสินค้าทั้งหมด
func (r *inMemoryRepository) FindAll() ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// ทำ slice copy เพื่อไม่ให้ภายนอกแก้ไข slice ต้นฉบับใน memory โดยตรง
	copied := make([]domain.Product, len(r.products))
	copy(copied, r.products)
	return copied, nil
}

// FindByID ค้นหาสินค้าตาม ID
func (r *inMemoryRepository) FindByID(id int) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, domain.ErrProductNotFound
}

// Create เพิ่มสินค้าใหม่
func (r *inMemoryRepository) Create(input domain.CreateProductInput) (*domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	newProduct := domain.Product{
		ID:          r.nextID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		CreatedAt:   time.Now(),
	}
	r.nextID++
	r.products = append(r.products, newProduct)

	return &newProduct, nil
}

// Update แก้ไขข้อมูลสินค้า
func (r *inMemoryRepository) Update(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.products {
		if p.ID == id {
			r.products[i].Name = input.Name
			r.products[i].Description = input.Description
			r.products[i].Price = input.Price
			r.products[i].Stock = input.Stock
			return &r.products[i], nil
		}
	}
	return nil, domain.ErrProductNotFound
}

// Delete ลบสินค้า
func (r *inMemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.products {
		if p.ID == id {
			copy(r.products[i:], r.products[i+1:])
			r.products[len(r.products)-1] = domain.Product{}
			r.products = r.products[:len(r.products)-1]
			return nil
		}
	}
	return domain.ErrProductNotFound
}

// FindAllBrands จำลองข้อมูลแบรนด์สำหรับ Unit Test
func (r *inMemoryRepository) FindAllBrands() ([]domain.Brand, error) {
	return []domain.Brand{
		{ID: 1, Name: "Keychron"},
		{ID: 2, Name: "Logitech G"},
		{ID: 3, Name: "HyperX"},
		{ID: 4, Name: "Alienware"},
	}, nil
}
