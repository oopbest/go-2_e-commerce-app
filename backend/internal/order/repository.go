package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type repository struct {
	db *sql.DB
}

// NewRepository Constructor สำหรับสร้าง Order Repository
func NewRepository(db *sql.DB) domain.OrderRepository {
	return &repository{
		db: db,
	}
}

// CreateOrder สร้างคำสั่งซื้อพร้อมตัดสต็อกอย่างปลอดภัย (Transaction + Pessimistic Lock)
func (r *repository) CreateOrder(ctx context.Context, userID int, items []domain.CheckoutItemInput) (*domain.Order, error) {
	// 1. เริ่ม Database Transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // ถ้าเกิด error หรือ panic จะ rollback อัตโนมัติ (ถ้า commit แล้ว rollback จะไม่มีผล)

	var totalAmount float64
	var orderItems []domain.OrderItem

	// 2. ตรวจสอบและตัดสต็อกสินค้าทีละรายการ
	for _, item := range items {
		var (
			productName  string
			currentPrice float64
			currentStock int
		)

		// 🔒 SELECT ... FOR UPDATE: ล็อกแถวสินค้านี้ไว้ในระดับ Database ป้องกัน Race Condition
		queryLock := `
			SELECT name, price, stock 
			FROM products 
			WHERE id = $1 
			FOR UPDATE
		`
		err := tx.QueryRowContext(ctx, queryLock, item.ProductID).Scan(&productName, &currentPrice, &currentStock)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("product ID %d not found", item.ProductID)
			}
			return nil, fmt.Errorf("failed to lock product %d: %w", item.ProductID, err)
		}

		// ตรวจสอบว่าสินค้ามีพอหรือไม่
		if currentStock < item.Quantity {
			return nil, fmt.Errorf("%w: product '%s' (available: %d, requested: %d)",
				domain.ErrInsufficientStock, productName, currentStock, item.Quantity)
		}

		// ตัดสต็อกใน Transaction
		queryDeduct := `UPDATE products SET stock = stock - $1 WHERE id = $2`
		if _, err := tx.ExecContext(ctx, queryDeduct, item.Quantity, item.ProductID); err != nil {
			return nil, fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, err)
		}

		subtotal := currentPrice * float64(item.Quantity)
		totalAmount += subtotal

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Price:       currentPrice,
		})
	}

	// 3. บันทึกหัวบิลคำสั่งซื้อ (ตาราง orders)
	queryOrder := `
		INSERT INTO orders (user_id, total_amount, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, created_at, status
	`
	var newOrder domain.Order
	newOrder.UserID = userID
	newOrder.TotalAmount = totalAmount

	err = tx.QueryRowContext(ctx, queryOrder, userID, totalAmount).Scan(
		&newOrder.ID, &newOrder.CreatedAt, &newOrder.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// 4. บันทึกรายการสินค้าในคำสั่งซื้อ (ตาราง order_items)
	queryOrderItem := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	for i := range orderItems {
		orderItems[i].OrderID = newOrder.ID
		err := tx.QueryRowContext(ctx, queryOrderItem,
			newOrder.ID,
			orderItems[i].ProductID,
			orderItems[i].Quantity,
			orderItems[i].Price,
		).Scan(&orderItems[i].ID, &orderItems[i].CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// 5. Commit Transaction (ยืนยันการเปลี่ยนแปลงทั้งหมดลง Database)
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit order transaction: %w", err)
	}

	newOrder.Items = orderItems
	return &newOrder, nil
}

// FindOrderByID ค้นหาคำสั่งซื้อตาม ID
func (r *repository) FindOrderByID(ctx context.Context, orderID, userID int) (*domain.Order, error) {
	queryOrder := `
		SELECT id, user_id, total_amount, status, created_at
		FROM orders
		WHERE id = $1 AND (user_id = $2 OR $2 = 0)
	`
	var o domain.Order
	err := r.db.QueryRowContext(ctx, queryOrder, orderID, userID).Scan(
		&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	// Query รายการสินค้าใน Order นั้นพร้อมชื่อสินค้า
	queryItems := `
		SELECT oi.id, oi.order_id, oi.product_id, p.name, oi.quantity, oi.price, oi.created_at
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.id ASC
	`
	rows, err := r.db.QueryContext(ctx, queryItems, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.Price, &item.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

// FindOrdersByUserID ดึงประวัติคำสั่งซื้อของ User
func (r *repository) FindOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// FindAllOrders ดึงคำสั่งซื้อทั้งหมด (สำหรับ Admin)
func (r *repository) FindAllOrders(ctx context.Context) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, created_at
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// CancelOrderAndRestoreStock ยกเลิกบิลที่หมดอายุและคืนสต็อกสินค้าใน Database Transaction
func (r *repository) CancelOrderAndRestoreStock(ctx context.Context, orderID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. ตรวจสอบสถานะ Order
	var status string
	err = tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1 FOR UPDATE", orderID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	// ถ้า Order ไม่ได้อยู่ในสถานะ pending ไม่ต้องทำอะไร
	if status != "pending" {
		return nil
	}

	// 2. ปรับสถานะ Order เป็น cancelled
	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = 'cancelled' WHERE id = $1", orderID)
	if err != nil {
		return err
	}

	// 3. ดึงรายการสินค้าใน Order เพื่อเตรียมคืนสต็อก
	rows, err := tx.QueryContext(ctx, "SELECT product_id, quantity FROM order_items WHERE order_id = $1", orderID)
	if err != nil {
		return err
	}

	type refundItem struct {
		productID int
		quantity  int
	}
	var items []refundItem
	for rows.Next() {
		var item refundItem
		if err := rows.Scan(&item.productID, &item.quantity); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	_ = rows.Close() // 👈 ปิด rows ทันทีที่อ่านเสร็จ ก่อนเริ่มรัน UPDATE

	if err := rows.Err(); err != nil {
		return err
	}

	// 4. คืนจำนวนสต็อกกลับเข้าสินค้าแต่ละรายการ
	for _, item := range items {
		_, err = tx.ExecContext(ctx, "UPDATE products SET stock = stock + $1 WHERE id = $2", item.quantity, item.productID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
