package flashsale

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

type repository struct {
	db  *sql.DB
	rdb *redis.Client
	ttl time.Duration
}

// NewRepository Constructor สำหรับ Flash Sale Repository พร้อม Redis Cache
func NewRepository(db *sql.DB, rdb *redis.Client) domain.FlashSaleRepository {
	return &repository{
		db:  db,
		rdb: rdb,
		ttl: 1 * time.Minute, // แคชแคมเปญไว้ 1 นาที
	}
}

// FindCurrentActive ดึงแคมเปญ Flash Sale ที่กำลังจัดอยู่ในปัจจุบันพร้อมรายการสินค้า
func (r *repository) FindCurrentActive(ctx context.Context) (*domain.FlashSale, error) {
	cacheKey := "flashsale:active"

	// 1. ตรวจสอบใน Redis Cache ก่อน
	if r.rdb != nil {
		cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var sale domain.FlashSale
			if err := json.Unmarshal([]byte(cachedJSON), &sale); err == nil {
				return &sale, nil
			}
		}
	}

	// 2. ค้นหาแคมเปญปัจจุบันจาก PostgreSQL
	querySale := `
		SELECT id, title, description, starts_at, expires_at, is_active, created_at
		FROM flash_sales
		WHERE is_active = TRUE 
		  AND starts_at <= CURRENT_TIMESTAMP 
		  AND expires_at > CURRENT_TIMESTAMP
		ORDER BY starts_at ASC
		LIMIT 1
	`
	var sale domain.FlashSale
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, querySale).Scan(
		&sale.ID, &sale.Title, &desc, &sale.StartsAt, &sale.ExpiresAt, &sale.IsActive, &sale.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // ไม่มีแคมเปญ Flash Sale ที่กำลังจัดอยู่
		}
		return nil, err
	}
	if desc.Valid {
		sale.Description = desc.String
	}

	// 3. ดึงรายการสินค้าในแคมเปญ Flash Sale พร้อมชื่อและรูปภาพ
	queryItems := `
		SELECT fsi.id, fsi.flash_sale_id, fsi.product_id, 
		       COALESCE(p.name, ''), COALESCE(p.image_url, ''),
		       p.price, fsi.flash_price, fsi.flash_stock, fsi.sold_count, fsi.created_at
		FROM flash_sale_items fsi
		JOIN products p ON fsi.product_id = p.id
		WHERE fsi.flash_sale_id = $1
		ORDER BY fsi.id ASC
	`
	rows, err := r.db.QueryContext(ctx, queryItems, sale.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sale.Items = []domain.FlashSaleItem{}
	for rows.Next() {
		var item domain.FlashSaleItem
		if err := rows.Scan(
			&item.ID, &item.FlashSaleID, &item.ProductID,
			&item.ProductName, &item.ImageURL,
			&item.OriginalPrice, &item.FlashPrice, &item.FlashStock, &item.SoldCount, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		sale.Items = append(sale.Items, item)
	}

	// 4. บันทึกลง Redis Cache
	if r.rdb != nil {
		if data, err := json.Marshal(sale); err == nil {
			_ = r.rdb.Set(ctx, cacheKey, data, r.ttl).Err()
		}
	}

	return &sale, nil
}
