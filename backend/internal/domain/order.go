package domain

import (
	"context"
	"errors"
	"time"
)

// ==========================================
// 1. Order Domain Errors
// ==========================================

var (
	ErrInsufficientStock = errors.New("insufficient stock for product")
	ErrOrderNotFound     = errors.New("order not found")
	ErrEmptyCart         = errors.New("cart items cannot be empty")
	ErrInvalidQuantity   = errors.New("item quantity must be greater than 0")
)

// ==========================================
// 2. Entities & DTOs
// ==========================================

// OrderItem ข้อมูลรายการสินค้าในคำสั่งซื้อ
type OrderItem struct {
	ID          int       `json:"id"`
	OrderID     int       `json:"order_id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"` // สำหรับแสดงชื่อสินค้าเวลา Query
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"` // ราคา ณ วันที่สั่งซื้อ
	CreatedAt   time.Time `json:"created_at"`
}

// Order ข้อมูลคำสั่งซื้อ
type Order struct {
	ID          int         `json:"id"`
	UserID      int         `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"` // pending, paid, cancelled
	Items       []OrderItem `json:"items,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// CheckoutItemInput ข้อมูลสินค้าแต่ละชิ้นที่ส่งมาตอน Checkout
type CheckoutItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CheckoutInput ข้อมูลรวมที่ส่งมาจากหน้าบ้านตอนกดสั่งซื้อ
type CheckoutInput struct {
	Items []CheckoutItemInput `json:"items"`
}

// ==========================================
// 3. Interfaces
// ==========================================

// OrderRepository สัญญาการทำงานของ Data Access Layer สำหรับ Order
type OrderRepository interface {
	CreateOrder(ctx context.Context, userID int, items []CheckoutItemInput) (*Order, error)
	FindOrderByID(ctx context.Context, orderID, userID int) (*Order, error)
	FindOrdersByUserID(ctx context.Context, userID int) ([]Order, error)
	FindAllOrders(ctx context.Context) ([]Order, error)
	CancelOrderAndRestoreStock(ctx context.Context, orderID int) error
}

// OrderService สัญญาการทำงานของ Business Logic สำหรับ Order
type OrderService interface {
	Checkout(ctx context.Context, userID int, input CheckoutInput) (*Order, error)
	GetOrderByID(ctx context.Context, orderID, userID int, userRole string) (*Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]Order, error)
	GetAllOrders(ctx context.Context) ([]Order, error)
}
