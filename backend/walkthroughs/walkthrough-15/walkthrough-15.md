# Walkthrough 15: Product Detail & Brand Entity (Phase A - Go Backend)

## 📌 บทนำและเป้าหมาย (Overview & Objectives)

ใน Milestone นี้ เราได้ทำการยกระดับระบบ **Product Catalog** ของ Go Clean Architecture จากเดิมที่มีเฉพาะข้อมูลพื้นฐาน (`name`, `description`, `price`, `stock`) ให้กลายเป็น **Enterprise-Grade E-Commerce Data Model** โดยเพิ่ม:
1. **Brand Entity**: รองรับข้อมูลแบรนด์สินค้า โลโก้ และคำอธิบาย
2. **Product Enrichment**: รูปภาพสินค้าความละเอียดสูง, รหัสสินค้าประจำร้าน (SKU), คะแนนรีวิว, และจำนวนรีวิว
3. **Technical Specifications (PostgreSQL JSONB)**: รองรับสเปกสินค้าเชิงลึกแบบยืดหยุ่น (เช่น layout, switch, connectivity, battery, warranty)
4. **Performance Indexing**: GIN Index สำหรับ JSONB และ B-Tree Index สำหรับ Foreign Key
5. **In-Memory Caching (Redis)**: แคชรายชื่อแบรนด์ทั้งหมด (`brands:all`)

---

## 🏛️ สถาปัตยกรรมระบบ (Architecture Flow)

```
[ Client / Frontend ]
         │
         ▼  HTTP GET /api/products/1  |  GET /api/brands
┌────────────────────────────────────────────────────────┐
│                   Go API Handler                       │
│  - GET /api/brands                                     │
│  - GET /api/products/{id}                              │
└──────────────────────────┬─────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────┐
│                Product Service Layer                   │
│  - GetAllBrands()                                      │
│  - GetProductByID(id)                                  │
└──────────────────────────┬─────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────┐
│            Redis Cached Repository Layer               │
│  - Check Redis ("brands:all", "products:all")          │
│  - Cache HIT ──► Return instant JSON                   │
│  - Cache MISS ──► Forward to Postgres Repository       │
└──────────────────────────┬─────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────┐
│           PostgreSQL Database (Migration 000005)       │
│  - Table: brands                                       │
│  - Table: products (with specs JSONB, brand_id FK)    │
│  - LEFT JOIN categories & brands                       │
└────────────────────────────────────────────────────────┘
```

---

## 🗄️ 1. Database Migration `000005`

### 1.1 ไฟล์ Migration ขาขึ้น (Up Migration)
- **ตำแหน่งไฟล์**: `backend/migrations/000005_add_brands_and_product_details.up.sql`

```sql
-- 1. สร้างตาราง brands
CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    logo_url TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. ใส่ข้อมูลแบรนด์เริ่มต้น (Seed Brands)
INSERT INTO brands (id, name, logo_url, description) VALUES
(1, 'Keychron', 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=100', 'Premium Custom Mechanical Keyboards'),
(2, 'Logitech G', 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=100', 'High-Performance Gaming Gear & Peripherals'),
(3, 'HyperX', 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=100', 'Professional Gaming Headsets & Esports Audio'),
(4, 'Alienware', 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=100', 'Ultra-Premium Gaming Displays & Monitors')
ON CONFLICT (id) DO NOTHING;

SELECT setval('brands_id_seq', (SELECT MAX(id) FROM brands));

-- 3. ขยายโครงสร้างตาราง products
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS brand_id INT REFERENCES brands(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS image_url TEXT,
ADD COLUMN IF NOT EXISTS sku VARCHAR(100),
ADD COLUMN IF NOT EXISTS specs JSONB DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS rating NUMERIC(2, 1) DEFAULT 5.0,
ADD COLUMN IF NOT EXISTS reviews_count INT DEFAULT 0;

-- 4. สร้าง Index สำหรับการค้นหาและกรองด้วย Brand และ Specs
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_products_specs ON products USING gin(specs);

-- 5. อัปเดตข้อมูลสินค้าเดิมให้มีรูปภาพสวยๆ แบรนด์ และสเปกครบถ้วน
UPDATE products 
SET 
    brand_id = 1,
    sku = 'KB-KEY-001',
    image_url = 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80',
    rating = 4.9,
    reviews_count = 142,
    specs = '{"layout": "75% (84 Keys)", "switch": "Gateron G Pro Red", "connectivity": "Bluetooth 5.1 / Type-C", "battery": "4000 mAh", "warranty": "2 Years"}'::jsonb
WHERE id = 1;

UPDATE products 
SET 
    brand_id = 2,
    sku = 'MS-LOGI-002',
    image_url = 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80',
    rating = 4.8,
    reviews_count = 98,
    specs = '{"sensor": "HERO 25K", "dpi": "100 - 25,600 DPI", "connectivity": "LIGHTSPEED Wireless", "weight": "63 grams", "battery": "70 Hours"}'::jsonb
WHERE id = 2;

UPDATE products 
SET 
    brand_id = 3,
    sku = 'HS-HYPER-003',
    image_url = 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=800&q=80',
    rating = 4.7,
    reviews_count = 76,
    specs = '{"driver": "Dynamic 53mm with Neodymium Magnets", "surround": "DTS Headphone:X Spatial Audio", "frame": "Aluminum", "microphone": "Detachable Noise-cancelling"}'::jsonb
WHERE id = 3;

UPDATE products 
SET 
    brand_id = 4,
    sku = 'MN-ALIEN-004',
    image_url = 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80',
    rating = 5.0,
    reviews_count = 54,
    specs = '{"size": "34-inch QD-OLED Curved 1800R", "resolution": "WQHD 3440 x 1440", "refresh_rate": "175Hz", "response_time": "0.1ms GtG", "hdr": "VESA DisplayHDR TrueBlack 400"}'::jsonb
WHERE id = 4;
```

### 1.2 ไฟล์ Rollback ขาลง (Down Migration)
- **ตำแหน่งไฟล์**: `backend/migrations/000005_add_brands_and_product_details.down.sql`

```sql
DROP INDEX IF EXISTS idx_products_specs;
DROP INDEX IF EXISTS idx_products_brand_id;

ALTER TABLE products 
DROP COLUMN IF EXISTS reviews_count,
DROP COLUMN IF EXISTS rating,
DROP COLUMN IF EXISTS specs,
DROP COLUMN IF EXISTS sku,
DROP COLUMN IF EXISTS image_url,
DROP COLUMN IF EXISTS brand_id;

DROP TABLE IF EXISTS brands;
```

---

## 💻 2. Go Clean Architecture Implementation

### 2.1 Domain Layer (`internal/domain/product.go`)
- สร้าง Entity `Brand`
- ขยาย `Product` struct พร้อม tag `json:",omitempty"`
- ขยาย `ProductRepository` และ `ProductService` interfaces ด้วย `FindAllBrands()` และ `GetAllBrands()`

```go
type Brand struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	LogoURL     string    `json:"logo_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Price        float64        `json:"price"`
	Stock        int            `json:"stock"`
	CategoryID   *int           `json:"category_id,omitempty"`
	CategoryName string         `json:"category_name,omitempty"`
	BrandID      *int           `json:"brand_id,omitempty"`
	BrandName    string         `json:"brand_name,omitempty"`
	ImageURL     string         `json:"image_url"`
	SKU          string         `json:"sku"`
	Specs        map[string]any `json:"specs,omitempty"`
	Rating       float64        `json:"rating"`
	ReviewsCount int            `json:"reviews_count"`
	CreatedAt    time.Time      `json:"created_at"`
}
```

### 2.2 PostgreSQL Repository (`internal/product/repository_postgres.go`)
- Query แบบ `LEFT JOIN` เข้ากับทั้งตาราง `categories` และ `brands`
- ป้องกันค่า SQL `NULL` ด้วย `COALESCE(c.name, '')`, `COALESCE(b.name, '')`
- Scan PostgreSQL JSONB ด้วยการอ่านไบต์ `var specsBytes []byte` แล้ว `json.Unmarshal`:

```go
var specsBytes []byte
if err := rows.Scan(
    &p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
    &p.CategoryID, &p.CategoryName,
    &p.BrandID, &p.BrandName,
    &p.ImageURL, &p.SKU,
    &specsBytes,
    &p.Rating, &p.ReviewsCount, &p.CreatedAt,
); err != nil {
    return nil, err
}

if len(specsBytes) > 0 {
    _ = json.Unmarshal(specsBytes, &p.Specs)
}
```

### 2.3 Redis Caching Layer (`internal/product/repository_cached.go`)
- เพิ่มการแคชรายชื่อแบรนด์ลง Redis ด้วยคีย์ `brands:all` เพื่อลดภาระของ Database:

```go
func (r *cachedRepository) FindAllBrands() ([]domain.Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := "brands:all"
	cachedJSON, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var brands []domain.Brand
		if err := json.Unmarshal([]byte(cachedJSON), &brands); err == nil {
			slog.Debug("⚡ Cache HIT: brands:all")
			return brands, nil
		}
	}

	brands, err := r.next.FindAllBrands()
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(brands); err == nil {
		_ = r.rdb.Set(ctx, cacheKey, data, r.ttl).Err()
	}
	return brands, nil
}
```

### 2.4 HTTP Handler (`internal/product/handler.go`)
- ลงทะเบียนเส้นทางสาธารณะ: `mux.HandleFunc("GET /api/brands", h.handleGetBrands)`
- เพิ่ม Swagger Doc Annotations สำหรับ Documenting API

---

## 🧪 3. การทดสอบและผลลัพธ์ (Verification & Test Results)

### 3.1 Unit Tests & Code Quality
```bash
go vet ./...
go test ./...
```
**ผลลัพธ์**: ผ่าน 100% (Pass ทุกแพ็กเกจ)

### 3.2 Live Endpoint Responses

#### 🔹 `GET /api/brands`
```json
[
  {
    "id": 1,
    "name": "Keychron",
    "logo_url": "https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=100",
    "description": "Premium Custom Mechanical Keyboards",
    "created_at": "2026-08-27T00:41:08.780436Z"
  },
  {
    "id": 2,
    "name": "Logitech G",
    "logo_url": "https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=100",
    "description": "High-Performance Gaming Gear & Peripherals",
    "created_at": "2026-08-27T00:41:08.780436Z"
  }
]
```

#### 🔹 `GET /api/products/1`
```json
{
  "id": 1,
  "name": "Mechanical Keyboard",
  "description": "RGB Hot-swappable",
  "price": 2590,
  "stock": 13,
  "category_id": 1,
  "category_name": "Gaming Gear",
  "brand_id": 1,
  "brand_name": "Keychron",
  "image_url": "https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80",
  "sku": "KB-KEY-001",
  "specs": {
    "battery": "4000 mAh",
    "connectivity": "Bluetooth 5.1 / Type-C",
    "layout": "75% (84 Keys)",
    "switch": "Gateron G Pro Red",
    "warranty": "2 Years"
  },
  "rating": 4.9,
  "reviews_count": 142,
  "created_at": "2026-08-20T23:47:01.375501Z"
}
```

---

## 📦 4. Git Commit Message

```bash
git add .
git commit -m "feat(catalog): add Brand entity and enrich Product with specs JSONB, SKU, and ratings" -m "- Add migration 000005 creating brands table with seed data and extending products table
- Add GIN index for specs JSONB and foreign key index for brand_id
- Enrich Product domain entity with BrandID, BrandName, CategoryName, SKU, Specs, Rating, and ReviewsCount
- Update PostgreSQL repository with category and brand LEFT JOINs and safe JSONB unmarshaling
- Implement FindAllBrands with Redis caching under brands:all key
- Register GET /api/brands endpoint with Swagger annotations
- Update unit tests and mocks in service_test.go and handler_test.go"
git push
```
