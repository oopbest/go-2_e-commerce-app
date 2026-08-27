package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

type repository struct {
	db  *sql.DB
	rdb *redis.Client
}

// NewRepository Constructor สำหรับสร้าง Order Repository
func NewRepository(db *sql.DB, rdb *redis.Client) domain.OrderRepository {
	return &repository{
		db:  db,
		rdb: rdb,
	}
}

// CreateOrder สร้างคำสั่งซื้อพร้อมตัดสต็อก ตรวจสอบ Flash Sale และหักส่วนลดคูปอง (Transaction + Pessimistic Lock)
func (r *repository) CreateOrder(ctx context.Context, userID int, input domain.CheckoutInput) (*domain.Order, error) {
	// 1. เริ่ม Database Transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var subtotalAmount float64
	var orderItems []domain.OrderItem
	hasFlashSaleDeduction := false

	// 2. ตรวจสอบและตัดสต็อกสินค้าทีละรายการ
	for _, item := range input.Items {
		var (
			productName  string
			currentPrice float64
			currentStock int
		)

		// 🔒 SELECT ... FOR UPDATE ล็อกแถวสินค้าในตาราง products
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

		if currentStock < item.Quantity {
			return nil, fmt.Errorf("%w: product '%s' (available: %d, requested: %d)",
				domain.ErrInsufficientStock, productName, currentStock, item.Quantity)
		}

		// ⚡ ตรวจสอบว่าสินค้าชิ้นนี้กำลังจัด Flash Sale และมีโควตาเหลืออยู่หรือไม่
		queryFlash := `
			SELECT fsi.id, fsi.flash_price, fsi.flash_stock, fsi.sold_count
			FROM flash_sale_items fsi
			JOIN flash_sales fs ON fsi.flash_sale_id = fs.id
			WHERE fsi.product_id = $1
			  AND fs.is_active = TRUE
			  AND fs.starts_at <= CURRENT_TIMESTAMP
			  AND fs.expires_at > CURRENT_TIMESTAMP
			  AND (fsi.flash_stock - fsi.sold_count) >= $2
			FOR UPDATE
		`
		var fsiID int
		var flashPrice float64
		var flashStock, soldCount int
		err = tx.QueryRowContext(ctx, queryFlash, item.ProductID, item.Quantity).Scan(&fsiID, &flashPrice, &flashStock, &soldCount)
		if err == nil {
			// สินค้าเข้าเงื่อนไข Flash Sale -> ปรับราคาเป็น Flash Price และตัดโควตา Flash Sale
			currentPrice = flashPrice
			_, err = tx.ExecContext(ctx, "UPDATE flash_sale_items SET sold_count = sold_count + $1 WHERE id = $2", item.Quantity, fsiID)
			if err != nil {
				return nil, fmt.Errorf("failed to update flash sale quota: %w", err)
			}
			hasFlashSaleDeduction = true
		}

		// ตัดสต็อกสินค้าหลัก
		queryDeduct := `UPDATE products SET stock = stock - $1 WHERE id = $2`
		if _, err := tx.ExecContext(ctx, queryDeduct, item.Quantity, item.ProductID); err != nil {
			return nil, fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, err)
		}

		subtotal := currentPrice * float64(item.Quantity)
		subtotalAmount += subtotal

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Price:       currentPrice,
		})
	}

	// 🎟️ 3. ตรวจสอบและคำนวณส่วนลดจากคูปอง (ถ้ามีการกรอกโค้ดมา)
	var discountAmount float64
	var couponID *int
	cleanCode := strings.TrimSpace(input.CouponCode)

	if cleanCode != "" {
		var promo domain.Promotion
		var desc sql.NullString
		queryPromo := `
			SELECT id, code, title, description, discount_type, discount_value,
			       min_spend, max_discount, total_quota, used_count,
			       starts_at, expires_at, is_active, created_at
			FROM promotions
			WHERE UPPER(code) = UPPER($1)
			FOR UPDATE
		`
		err := tx.QueryRowContext(ctx, queryPromo, cleanCode).Scan(
			&promo.ID, &promo.Code, &promo.Title, &desc, &promo.DiscountType, &promo.DiscountValue,
			&promo.MinSpend, &promo.MaxDiscount, &promo.TotalQuota, &promo.UsedCount,
			&promo.StartsAt, &promo.ExpiresAt, &promo.IsActive, &promo.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrPromotionNotFound
			}
			return nil, fmt.Errorf("failed to lock promotion: %w", err)
		}

		// คำนวณส่วนลดตามกฎธุรกิจ
		discount, err := promo.CalculateDiscount(subtotalAmount)
		if err != nil {
			return nil, err
		}

		// ตัดโควตาคูปอง
		queryIncPromo := `UPDATE promotions SET used_count = used_count + 1 WHERE id = $1`
		if _, err := tx.ExecContext(ctx, queryIncPromo, promo.ID); err != nil {
			return nil, fmt.Errorf("failed to increment promotion usage: %w", err)
		}

		discountAmount = discount
		couponID = &promo.ID
	}

	// คำนวณยอดสุทธิ (ยอดรวม - ส่วนลด) ไม่ให้ติดลบ
	finalTotal := math.Max(0, subtotalAmount-discountAmount)
	finalTotal = math.Round(finalTotal*100) / 100

	// 4. บันทึกหัวบิลคำสั่งซื้อ (ตาราง orders) พร้อมบันทึก coupon_id และ discount_amount
	queryOrder := `
		INSERT INTO orders (user_id, total_amount, discount_amount, coupon_id, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, created_at, status
	`
	var newOrder domain.Order
	newOrder.UserID = userID
	newOrder.TotalAmount = finalTotal
	newOrder.DiscountAmount = discountAmount
	newOrder.CouponID = couponID

	err = tx.QueryRowContext(ctx, queryOrder, userID, finalTotal, discountAmount, couponID).Scan(
		&newOrder.ID, &newOrder.CreatedAt, &newOrder.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// 5. บันทึกรายการสินค้าในคำสั่งซื้อ (ตาราง order_items)
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

	// 6. Commit Transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit order transaction: %w", err)
	}

	// 7. เคลียร์ Redis Cache สินค้า (และ Flash Sale ถ้ามีการตัดโควตา)
	if r.rdb != nil {
		_ = r.rdb.Del(ctx, "products:all").Err()
		if hasFlashSaleDeduction {
			_ = r.rdb.Del(ctx, "flashsale:active").Err()
		}
	}

	newOrder.Items = orderItems
	return &newOrder, nil
}

// FindOrderByID ค้นหาคำสั่งซื้อตาม ID
func (r *repository) FindOrderByID(ctx context.Context, orderID, userID int) (*domain.Order, error) {
	queryOrder := `
		SELECT id, user_id, total_amount, discount_amount, coupon_id, status, created_at
		FROM orders
		WHERE id = $1 AND (user_id = $2 OR $2 = 0)
	`
	var o domain.Order
	var couponID sql.NullInt64
	err := r.db.QueryRowContext(ctx, queryOrder, orderID, userID).Scan(
		&o.ID, &o.UserID, &o.TotalAmount, &o.DiscountAmount, &couponID, &o.Status, &o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	if couponID.Valid {
		cID := int(couponID.Int64)
		o.CouponID = &cID
	}

	// Query รายการสินค้าใน Order
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
		SELECT id, user_id, total_amount, discount_amount, coupon_id, status, created_at
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
		var couponID sql.NullInt64
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.DiscountAmount, &couponID, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		if couponID.Valid {
			cID := int(couponID.Int64)
			o.CouponID = &cID
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// FindAllOrders ดึงคำสั่งซื้อทั้งหมด (สำหรับ Admin)
func (r *repository) FindAllOrders(ctx context.Context) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, total_amount, discount_amount, coupon_id, status, created_at
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
		var couponID sql.NullInt64
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.DiscountAmount, &couponID, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		if couponID.Valid {
			cID := int(couponID.Int64)
			o.CouponID = &cID
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// CancelOrderAndRestoreStock ยกเลิกบิลที่หมดอายุ คืนสต็อกสินค้า และคืนสิทธิ์คูปองใน Database Transaction
func (r *repository) CancelOrderAndRestoreStock(ctx context.Context, orderID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. ตรวจสอบสถานะ Order และ coupon_id
	var status string
	var couponID sql.NullInt64
	err = tx.QueryRowContext(ctx, "SELECT status, coupon_id FROM orders WHERE id = $1 FOR UPDATE", orderID).Scan(&status, &couponID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != "pending" {
		return nil
	}

	// 2. ปรับสถานะ Order เป็น cancelled
	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = 'cancelled' WHERE id = $1", orderID)
	if err != nil {
		return err
	}

	// 3. คืนโควตาสิทธิ์คูปอง (ถ้ามีการใช้คูปองในบิลนี้)
	if couponID.Valid {
		_, _ = tx.ExecContext(ctx, "UPDATE promotions SET used_count = GREATEST(0, used_count - 1) WHERE id = $1", couponID.Int64)
	}

	// 4. ดึงรายการสินค้าใน Order เพื่อคืนสต็อก
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
	_ = rows.Close()

	if err := rows.Err(); err != nil {
		return err
	}

	// 5. คืนจำนวนสต็อกกลับเข้าสินค้าแต่ละรายการ
	for _, item := range items {
		_, err = tx.ExecContext(ctx, "UPDATE products SET stock = stock + $1 WHERE id = $2", item.quantity, item.productID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if r.rdb != nil {
		_ = r.rdb.Del(ctx, "products:all").Err()
	}

	return nil
}
