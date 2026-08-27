package product_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ==========================================
// 1. สร้าง Mock Repository จำลอง (Implement domain.ProductRepository)
// ==========================================

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) FindAll() ([]domain.Product, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductRepository) FindByID(id int) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Create(input domain.CreateProductInput) (*domain.Product, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) FindAllBrands() ([]domain.Brand, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Brand), args.Error(1)
}

func (m *MockProductRepository) FindWithFilter(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductListResponse), args.Error(1)
}

// ==========================================
// 2. Unit Tests สำหรับ Product Service
// ==========================================

func TestCreateProduct(t *testing.T) {
	t.Run("Success: Create Valid Product", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		input := domain.CreateProductInput{
			Name:        "Wireless Keyboard",
			Description: "Compact 65%",
			Price:       2490.00,
			Stock:       10,
		}

		expectedProduct := &domain.Product{
			ID:          1,
			Name:        input.Name,
			Description: input.Description,
			Price:       input.Price,
			Stock:       input.Stock,
			CreatedAt:   time.Now(),
		}

		// กำหนดพฤติกรรม Mock: เมื่อเรียก Create ด้วย input นี้ ให้คืน expectedProduct กลับไป
		mockRepo.On("Create", input).Return(expectedProduct, nil)

		// สั่งรันฟังก์ชันจริง
		result, err := service.CreateProduct(input)

		require.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, result.ID)
		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Price, result.Price)

		// ยืนยันว่า Mock ถูกเรียกใช้งานจริงตามที่ตั้งไว้
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation Error: Empty Product Name", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		input := domain.CreateProductInput{
			Name:  "   ", // ชื่อว่างเปล่า
			Price: 100,
			Stock: 5,
		}

		result, err := service.CreateProduct(input)

		assert.Error(t, err)
		assert.Equal(t, "product name is required", err.Error())
		assert.Nil(t, result)

		// มั่นใจว่าไม่ได้แอบเรียก Repo ไปบันทึกใน Database
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Validation Error: Invalid Price", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		input := domain.CreateProductInput{
			Name:  "Gaming Mouse",
			Price: -10, // ราคาติดลบ
			Stock: 5,
		}

		result, err := service.CreateProduct(input)

		assert.Error(t, err)
		assert.Equal(t, "price must be greater than 0", err.Error())
		assert.Nil(t, result)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Validation Error: Negative Stock", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		input := domain.CreateProductInput{
			Name:  "Gaming Mouse",
			Price: 990,
			Stock: -5, // สต็อกติดลบ
		}

		result, err := service.CreateProduct(input)

		assert.Error(t, err)
		assert.Equal(t, "stock cannot be negative", err.Error())
		assert.Nil(t, result)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestGetProductByID(t *testing.T) {
	t.Run("Success: Find Product by ID", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		expectedProduct := &domain.Product{
			ID:    1,
			Name:  "Mechanical Keyboard",
			Price: 2590,
			Stock: 15,
		}

		mockRepo.On("FindByID", 1).Return(expectedProduct, nil)

		result, err := service.GetProductByID(1)

		require.NoError(t, err)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Mechanical Keyboard", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Validation Error: Invalid ID (Zero or Negative)", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		result, err := service.GetProductByID(0)

		assert.Error(t, err)
		assert.Equal(t, "invalid product ID", err.Error())
		assert.Nil(t, result)
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("Error: Product Not Found in Repository", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		service := product.NewService(mockRepo)

		mockRepo.On("FindByID", 999).Return(nil, errors.New("product not found"))

		result, err := service.GetProductByID(999)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
