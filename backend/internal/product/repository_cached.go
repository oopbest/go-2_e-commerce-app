package product

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/redis/go-redis/v9"
)

type cachedRepository struct {
	next domain.ProductRepository // หุ้ม Postgres Repository ไว้ข้างใน
	rdb  *redis.Client
	ttl  time.Duration
}

// NewCachedRepository Constructor สร้าง Caching Layer หุ้ม Repository จริง
func NewCachedRepository(next domain.ProductRepository, rdb *redis.Client, ttl time.Duration) domain.ProductRepository {
	return &cachedRepository{
		next: next,
		rdb:  rdb,
		ttl:  ttl,
	}
}

// FindAll ดึงสินค้าทั้งหมด (ตรวจแคชก่อนเสมอ)
func (r *cachedRepository) FindAll() ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := "products:all"

	// 1. ตรวจสอบใน Redis
	cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []domain.Product
		if err := json.Unmarshal([]byte(cachedJSON), &products); err == nil {
			slog.Debug("⚡ Cache HIT: products:all")
			return products, nil
		}
	}

	// 2. Cache Miss: ไปดึงจาก Database จริง (Postgres)
	slog.Debug("🗄️ Cache MISS: products:all -> fetching from Postgres")
	products, err := r.next.FindAll()
	if err != nil {
		return nil, err
	}

	// 3. เซฟผลลัพธ์ลง Redis พร้อมกำหนดเวลาหมดอายุ (TTL)
	if data, err := json.Marshal(products); err == nil {
		_ = r.rdb.Set(ctx, cacheKey, data, r.ttl).Err()
	}

	return products, nil
}

// FindByID ดึงสินค้าตาม ID
func (r *cachedRepository) FindByID(id int) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("product:%d", id)

	// 1. ตรวจสอบใน Redis
	cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var p domain.Product
		if err := json.Unmarshal([]byte(cachedJSON), &p); err == nil {
			slog.Debug("⚡ Cache HIT", "key", cacheKey)
			return &p, nil
		}
	}

	// 2. Cache Miss: ดึงจาก Postgres
	slog.Debug("🗄️ Cache MISS -> fetching from Postgres", "key", cacheKey)
	product, err := r.next.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 3. เซฟลง Redis
	if data, err := json.Marshal(product); err == nil {
		_ = r.rdb.Set(ctx, cacheKey, data, r.ttl).Err()
	}

	return product, nil
}

// Create สร้างสินค้าใหม่ และทำการล้างแคช (Cache Invalidation)
func (r *cachedRepository) Create(input domain.CreateProductInput) (*domain.Product, error) {
	product, err := r.next.Create(input)
	if err != nil {
		return nil, err
	}

	// ล้างแคชรายการสินค้าทั้งหมด เพื่อให้คนต่อไปได้รายการที่อัปเดต
	r.invalidateCache("products:all")
	return product, nil
}

// Update แก้ไขสินค้า และล้างแคชทั้งรายชิ้นและรายการรวม
func (r *cachedRepository) Update(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	product, err := r.next.Update(id, input)
	if err != nil {
		return nil, err
	}

	// ล้างแคชทั้ง products:all และ product:{id}
	r.invalidateCache("products:all", fmt.Sprintf("product:%d", id))
	return product, nil
}

// Delete ลบสินค้า และล้างแคช
func (r *cachedRepository) Delete(id int) error {
	if err := r.next.Delete(id); err != nil {
		return err
	}

	r.invalidateCache("products:all", fmt.Sprintf("product:%d", id))
	return nil
}

// invalidateCache Helper สำหรับลบ Key ใน Redis
func (r *cachedRepository) invalidateCache(keys ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, k := range keys {
		_ = r.rdb.Del(ctx, k).Err()
		slog.Debug("🧹 Cache INVALIDATED", "key", k)
	}
}

// FindAllBrands ดึงรายชื่อแบรนด์ทั้งหมด (พร้อมแคชใน Redis ด้วยคีย์ brands:all)
func (r *cachedRepository) FindAllBrands() ([]domain.Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := "brands:all"

	// 1. ตรวจสอบใน Redis ก่อน
	cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var brands []domain.Brand
		if err := json.Unmarshal([]byte(cachedJSON), &brands); err == nil {
			slog.Debug("⚡ Cache HIT: brands:all")
			return brands, nil
		}
	}

	// 2. Cache Miss: ดึงจาก Postgres
	slog.Debug("🗄️ Cache MISS: brands:all -> fetching from Postgres")
	brands, err := r.next.FindAllBrands()
	if err != nil {
		return nil, err
	}

	// 3. เซฟลง Redis
	if data, err := json.Marshal(brands); err == nil {
		_ = r.rdb.Set(ctx, cacheKey, data, r.ttl).Err()
	}

	return brands, nil
}

// FindWithFilter ดึงรายการสินค้าพร้อมตัวกรอง (ตรวจแคชตาม Query Parameters เสมอ)
func (r *cachedRepository) FindWithFilter(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResponse, error) {
	// สร้าง Cache Key เฉพาะเจาะจงตามเงื่อนไขที่ส่งมา
	cacheKey := fmt.Sprintf(
		"products:filter:s=%s:c=%v:b=%v:min=%v:max=%v:stk=%t:sort=%s:p=%d:l=%d",
		filter.Search,
		derefInt(filter.CategoryID),
		derefInt(filter.BrandID),
		derefFloat(filter.MinPrice),
		derefFloat(filter.MaxPrice),
		filter.InStockOnly,
		filter.SortBy,
		filter.Page,
		filter.Limit,
	)

	// 1. ตรวจสอบใน Redis
	if r.rdb != nil {
		cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp domain.ProductListResponse
			if err := json.Unmarshal([]byte(cachedJSON), &resp); err == nil {
				slog.Debug("⚡ Cache HIT: " + cacheKey)
				return &resp, nil
			}
		}
	}

	// 2. Cache Miss: ดึงจาก Postgres
	resp, err := r.next.FindWithFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 3. บันทึกลง Redis (TTL 1 นาที สำหรับ Dynamic Queries)
	if r.rdb != nil {
		if data, err := json.Marshal(resp); err == nil {
			_ = r.rdb.Set(ctx, cacheKey, data, 1*time.Minute).Err()
		}
	}

	return resp, nil
}

func derefInt(i *int) any {
	if i == nil {
		return "nil"
	}
	return *i
}

func derefFloat(f *float64) any {
	if f == nil {
		return "nil"
	}
	return *f
}
