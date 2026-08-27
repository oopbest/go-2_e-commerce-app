# Walkthrough 17: Search, Multi-Filter & Dynamic Sorting Engine - Backend Phase C

## 📌 1. ภาพรวมและสถาปัตยกรรมระบบ (Overview & Architecture)

ใน Milestone **Phase C (Backend)** เราได้พัฒนา **Search, Multi-Filter & Sorting Engine** สำหรับการค้นหาและคัดกรองสินค้าหลายมิติระดับ Enterprise ขับเคลื่อนด้วย Clean Architecture, Dynamic Parameterized SQL (ป้องกัน SQL Injection 100%), และ Intelligent Dynamic In-Memory Cache บน Redis:

```
                                    Client Request (URL Query Params)
               [ ?search=keyboard&brand_id=1&min_price=1000&sort_by=price_asc&page=1 ]
                                                   │
                                                   ▼
                                     ┌───────────────────────────┐
                                     │    HTTP Handler Layer     │
                                     │  • Parse & Sanitize Query │
                                     │  • Build ProductFilter DTO│
                                     └─────────────┬─────────────┘
                                                   │
                                                   ▼
                                     ┌───────────────────────────┐
                                     │   Product Service Layer   │
                                     │  • Orchestration & Context│
                                     └─────────────┬─────────────┘
                                                   │
                                                   ▼
                                     ┌───────────────────────────┐
                                     │ Redis Cache (Decorator)   │
                                     │  Key: products:filter:... │
                                     └──────┬─────────────▲──────┘
                             Cache MISS     │             │ Cache Set (TTL: 1m)
                                            ▼             │
                               ┌──────────────────────────┴────┐
                               │   PostgreSQL Repository       │
                               │  1. Dynamic WHERE Slice       │
                               │  2. Parameterized ($1, $2...) │
                               │  3. Total Count Query         │
                               │  4. Dynamic ORDER BY & LIMIT  │
                               └───────────────────────────────┘
```

---

## 🛠️ 2. สรุปการพัฒนา Code แต่ละส่วนอย่างละเอียด

### 2.1 Domain Layer (`backend/internal/domain/product.go`)
นิยาม DTO สำหรับรับค่าตัวกรอง และผลลัพธ์การค้นหาพร้อมข้อมูลการแบ่งหน้า (Pagination):

```go
type ProductFilter struct {
	Search      string   `json:"search"`        // ค้นหา Keyword จากชื่อหรือคำอธิบายสินค้า
	CategoryID  *int     `json:"category_id"`   // กรองตามหมวดหมู่
	BrandID     *int     `json:"brand_id"`      // กรองตามแบรนด์
	MinPrice    *float64 `json:"min_price"`     // กรองราคาต่ำสุด
	MaxPrice    *float64 `json:"max_price"`     // กรองราคาสูงสุด
	InStockOnly bool     `json:"in_stock_only"` // กรองเฉพาะสินค้าที่มีสต็อก (> 0)
	SortBy      string   `json:"sort_by"`       // "price_asc", "price_desc", "rating", "newest"
	Page        int      `json:"page"`          // หน้าที่ต้องการดึง (default: 1)
	Limit       int      `json:"limit"`         // จำนวนต่อหน้า (default: 20, max: 100)
}

type ProductListResponse struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
```

---

### 2.2 PostgreSQL Dynamic SQL Builder (`backend/internal/product/repository_postgres.go`)
- **Dynamic WHERE Clause**: ต่อเงื่อนไขผ่าน `conditions = append(conditions, ...)` โดยใช้ Parameterized Index `$1, $2, $3...` อย่างปลอดภัย ป้องกัน SQL Injection 100%
- **Total Count Query**: นับจำนวนสินค้าทั้งหมดที่ตรงตามเงื่อนไขเพื่อคำนวณ `totalPages`
- **Dynamic Sorting**: รองรับ `price_asc`, `price_desc`, `rating`, `newest` (มี fallback เป็น `p.id ASC`)
- **Safe Pagination**: กำหนด `LIMIT` และ `OFFSET` ตามสูตร `offset = (page - 1) * limit`

---

### 2.3 Intelligent Redis Caching (`backend/internal/product/repository_cached.go`)
สร้าง Cache Key แบบ Dynamic ตามค่าจริงของพารามิเตอร์:
```go
cacheKey := fmt.Sprintf(
	"products:filter:s=%s:c=%v:b=%v:min=%v:max=%v:stk=%t:sort=%s:p=%d:l=%d",
	filter.Search, derefInt(filter.CategoryID), derefInt(filter.BrandID),
	derefFloat(filter.MinPrice), derefFloat(filter.MaxPrice),
	filter.InStockOnly, filter.SortBy, filter.Page, filter.Limit,
)
```
- แคชผลลัพธ์ลง Redis เป็นเวลา **1 นาที** เพื่อให้ค้นหาได้เร็วระดับ **<1ms** และสต็อกไม่อัพเดทช้าเกินไป

---

### 2.4 HTTP Handler & Backward Compatibility (`backend/internal/product/handler.go`)
- หาก Client ไม่ส่ง Query Params ใดๆ มาเลย (`len(q) == 0`): จะคืน `[]Product` ตามเดิม ทำให้หน้าเว็บเดิมและ Unit Test ไม่พัง
- หากมี Query Params: จะแปลงค่าและส่งกลับเป็น `ProductListResponse`

---

## 🧪 3. ผลการทดสอบและการตรวจสอบ (Verification)

### 3.1 Unit Testing
รันชุดทดสอบทุกแพ็กเกจ:
```bash
go test ./...
```
**ผลลัพธ์**: ผ่าน 100% ทุกแพ็กเกจ (รวมถึง `product_test` ที่อัปเดต Mock เรียบร้อย)

### 3.2 Live Endpoint Testing บน Docker (`:8080`)

1. **ทดสอบ Backward Compatibility (ไม่ส่ง Params)**:
   ```bash
   curl -s http://localhost:8080/api/products
   ```
   *ผลลัพธ์*: คืน Array สินค้า `[...]` 4 ชิ้นปกติ ไม่ส่งผลกระทบต่อ Storefront เดิม

2. **ทดสอบ Keyword Search (`?search=keyboard`)**:
   ```bash
   curl -s "http://localhost:8080/api/products?search=keyboard"
   ```
   *ผลลัพธ์*: คืนเฉพาะ `Mechanical Keyboard` พร้อม `total_count: 1, total_pages: 1`

3. **ทดสอบ Sorting & Pagination (`?sort_by=price_desc&limit=2`)**:
   ```bash
   curl -s "http://localhost:8080/api/products?sort_by=price_desc&limit=2"
   ```
   *ผลลัพธ์*: เรียงสินค้าแพงสุดขึ้นก่อน:
   - ชิ้นที่ 1: `UltraWide Curved Monitor 34` (฿14,900)
   - ชิ้นที่ 2: `Mechanical Keyboard` (฿2,590)
   - ข้อมูลเพจ: `total_count: 4, page: 1, limit: 2, total_pages: 2`

4. **ทดสอบ Brand Filter (`?brand_id=2`)**:
   ```bash
   curl -s "http://localhost:8080/api/products?brand_id=2"
   ```
   *ผลลัพธ์*: คืนเฉพาะสินค้าของ `Logitech G` (`Wireless Mouse`, ฿1,290)

---

## 📦 4. คำสั่ง Git Commit สำหรับ Backend Phase C

```bash
cd backend
git add .
git commit -m "feat(catalog): implement search, multi-filter, dynamic sorting, and pagination engine" -m "- Add ProductFilter and ProductListResponse DTOs in domain
- Build safe parameterized dynamic SQL query builder in Postgres repository
- Implement query-based intelligent dynamic Redis caching in cached repository
- Upgrade HTTP handler with comprehensive query parameter parsing and backward compatibility
- Regenerate Swagger API documentation with query parameter annotations
- Document complete Phase C backend architecture in walkthrough-17.md"
git push
```
