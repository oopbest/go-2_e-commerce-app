package product_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/middleware"
	"github.com/oopbest/ecommerce-app/internal/product"
	"github.com/oopbest/ecommerce-app/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ==========================================
// 1. Mock Product Service (จำลอง Service Layer)
// ==========================================

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) GetAllProducts() ([]domain.Product, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(id int) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) CreateProduct(input domain.CreateProductInput) (*domain.Product, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductService) GetAllBrands() ([]domain.Brand, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Brand), args.Error(1)
}

func (m *MockProductService) GetProductsWithFilter(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductListResponse), args.Error(1)
}

// ==========================================
// 2. HTTP Handler Tests
// ==========================================

func TestHandleGetProducts(t *testing.T) {
	t.Run("Status 200 OK: Return Product List", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		expectedProducts := []domain.Product{
			{ID: 1, Name: "Mechanical Keyboard", Price: 2590, Stock: 10},
			{ID: 2, Name: "Wireless Mouse", Price: 1290, Stock: 20},
		}

		mockService.On("GetAllProducts").Return(expectedProducts, nil)

		// 1. ตั้งค่า Router
		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		// 2. จำลอง HTTP Request & Response Recorder
		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		rec := httptest.NewRecorder()

		// 3. ยิง Request เข้า Router
		mux.ServeHTTP(rec, req)

		// 4. ตรวจสอบผลลัพธ์
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var responseProducts []domain.Product
		err := json.NewDecoder(rec.Body).Decode(&responseProducts)
		require.NoError(t, err)
		assert.Len(t, responseProducts, 2)
		assert.Equal(t, "Mechanical Keyboard", responseProducts[0].Name)

		mockService.AssertExpectations(t)
	})
}

func TestHandleGetProductByID(t *testing.T) {
	t.Run("Status 200 OK: Return Product by ID", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		expectedProduct := &domain.Product{
			ID:    1,
			Name:  "Mechanical Keyboard",
			Price: 2590,
			Stock: 10,
		}

		mockService.On("GetProductByID", 1).Return(expectedProduct, nil)

		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var res domain.Product
		err := json.NewDecoder(rec.Body).Decode(&res)
		require.NoError(t, err)
		assert.Equal(t, 1, res.ID)
		assert.Equal(t, "Mechanical Keyboard", res.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Status 404 Not Found", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		mockService.On("GetProductByID", 999).Return(nil, domain.ErrProductNotFound)

		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		req := httptest.NewRequest(http.MethodGet, "/api/products/999", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Status 400 Bad Request: Invalid ID Format", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		// ส่ง id เป็นตัวหนังสือ "abc" แทนตัวเลข
		req := httptest.NewRequest(http.MethodGet, "/api/products/abc", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		mockService.AssertNotCalled(t, "GetProductByID")
	})
}

func TestHandleCreateProduct(t *testing.T) {
	t.Run("Status 201 Created: Valid Payload by Admin", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		input := domain.CreateProductInput{
			Name:        "Gaming Mouse",
			Description: "Wireless 2.4GHz",
			Price:       1290,
			Stock:       20,
		}

		createdProduct := &domain.Product{
			ID:          10,
			Name:        input.Name,
			Description: input.Description,
			Price:       input.Price,
			Stock:       input.Stock,
		}

		mockService.On("CreateProduct", input).Return(createdProduct, nil)

		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		bodyBytes, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// 🔒 ฝัง Claims จำลองของ Admin เข้าไปใน Context เพื่อให้ผ่าน RequireRole("admin")
		adminClaims := &security.CustomClaims{
			UserID: 1,
			Email:  "admin@store.com",
			Role:   "admin",
		}
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminClaims)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var res domain.Product
		_ = json.NewDecoder(rec.Body).Decode(&res)
		assert.Equal(t, 10, res.ID)
		assert.Equal(t, "Gaming Mouse", res.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Status 403 Forbidden: Customer Role Cannot Create Product", func(t *testing.T) {
		mockService := new(MockProductService)
		handler := product.NewHandler(mockService)

		mux := http.NewServeMux()
		noAuthMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
		handler.RegisterRoutes(mux, noAuthMiddleware)

		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer([]byte(`{}`)))

		// 🔒 ฝัง Customer Claims (ไม่ใช่ Admin)
		customerClaims := &security.CustomClaims{
			UserID: 2,
			Email:  "buyer@gmail.com",
			Role:   "customer",
		}
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, customerClaims)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		// ต้องได้ 403 Forbidden และห้ามเรียก Service
		assert.Equal(t, http.StatusForbidden, rec.Code)
		mockService.AssertNotCalled(t, "CreateProduct")
	})
}
