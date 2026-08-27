# Walkthrough 16: Marketing & Deals Engine (Promotion / Coupon + Flash Sale) - Backend Phase B

## 📌 1. ภาพรวมและสถาปัตยกรรมระบบ (Overview & Architecture)

ใน Milestone **Phase B** เราได้พัฒนา **Marketing & Deals Engine** ซึ่งเป็นกลไกขับเคลื่อนการส่งเสริมการขายระดับ Enterprise ประกอบด้วย 2 ฟีเจอร์หลักที่ทำงานร่วมกับระบบ Checkout Transaction และ In-Memory Cache:

```
                                  ┌─────────────────────────────────────────┐
                                  │        Phase B: Marketing Engine        │
                                  └────────────────────┬────────────────────┘
                                                       │
                 ┌─────────────────────────────────────┴─────────────────────────────────────┐
                 ▼                                                                           ▼
   [ 1. Promotion & Coupon Engine ]                                            [ 2. Flash Sale Engine ]
 • คูปองส่วนลด: Fixed (บาท) / Percentage (%)                                  • แคมเปญลดราคาแบบจำกัดเวลา
 • เงื่อนไขขั้นต่ำ (Min Spend) & เพดานลดสูงสุด (Max Discount)                   • ราคาพิเศษ (Flash Price)
 • จำกัดโควตาสิทธิ์การใช้ (Total Quota vs Used Count)                         • โควตาสต็อกแยกพิเศษ (Flash Stock)
 • ป้องกันการใช้เกินสิทธิ์ด้วย Database Atomic Guard                            • ดึงข้อมูลเร็วระดับ <1ms ด้วย Redis Cache
                 │                                                                           │
                 └─────────────────────────────────────┬─────────────────────────────────────┘
                                                       │
                                                       ▼
                       ┌─────────────────────────────────────────────────────────────┐
                       │          Atomic Order Checkout (Database Transaction)       │
                       │  1. SELECT ... FOR UPDATE ล็อกสต็อกสินค้าหลัก               │
                       │  2. ตรวจสอบและตัดโควตา Flash Sale (ปรับเป็น flash_price)    │
                       │  3. ตรวจสอบและล็อกคูปอง (คำนวณส่วนลด & ตัด used_count)      │
                       │  4. บันทึก orders พร้อม discount_amount และ coupon_id       │
                       │  5. เคลียร์ Redis Cache (products:all, flashsale:active)    │
                       └─────────────────────────────────────────────────────────────┘
```

---

## 🗄️ 2. Database Migration (`000006_create_promotions_and_flash_sales_tables`)

### 2.1 ไฟล์ Migration ขาขึ้น (`up.sql`)
- **`promotions`**: ตารางคูปองส่วนลด รองรับทั้ง `fixed` และ `percentage` พร้อมเงื่อนไข `min_spend`, `max_discount`, `total_quota`, `used_count`
- **`flash_sales`**: ตารางแคมเปญแฟลชเซลล์พร้อมช่วงเวลาเริ่มต้น-สิ้นสุด (`starts_at`, `expires_at`)
- **`flash_sale_items`**: ตารางเชื่อมสินค้าเข้ากับแฟลชเซลล์ พร้อมกำหนด `flash_price`, `flash_stock`, และ `sold_count`
- **`orders`**: เพิ่มคอลัมน์ `coupon_id INT REFERENCES promotions(id)` และ `discount_amount NUMERIC(10, 2)`
- **Indexes**: สร้าง B-Tree Index บน `promotions(code)`, `promotions(is_active, expires_at)`, `flash_sales(is_active, starts_at, expires_at)`, และ `flash_sale_items(product_id)`
- **Seed Data**:
  - `WELCOME100`: ลด 100 บาท เมื่อซื้อครบ 1,000 บาท (โควตา 500 สิทธิ์)
  - `NEXTGEN15`: ลด 15% สูงสุด 300 บาท เมื่อซื้อครบ 1,500 บาท (โควตา 200 สิทธิ์)
  - `Midnight Flash Deals ⚡`: Mechanical Keyboard (ID 1: ปกติ 2,590 -> 1,990) และ Gaming Headset (ID 3: ปกติ 1,990 -> 1,490) โควตาอย่างละ 5 ชิ้น

---

## 🏗️ 3. การพัฒนา Clean Architecture ในแต่ละ Layer

### 3.1 Domain Layer (`internal/domain/`)
- **`promotion.go`**:
  - นิยาม `Promotion`, `ValidationResult`, และ Domain Errors (`ErrPromotionNotFound`, `ErrPromotionExpired`, `ErrPromotionQuotaExceeded`, `ErrPromotionMinSpendNotMet`)
  - **Pure Domain Method `(p *Promotion) CalculateDiscount(subtotal float64)`**: คำนวณส่วนลดตามกฎธุรกิจโดยไม่พึ่งพา I/O หรือ Database ทำให้ทดสอบแบบ Unit Test ได้ง่ายและปลอดภัย
- **`flash_sale.go`**:
  - นิยาม `FlashSale`, `FlashSaleItem`, และ Interfaces สำหรับ Data Access
- **`order.go`**:
  - เพิ่มฟิลด์ `DiscountAmount float64` และ `CouponID *int` ใน `Order` Entity
  - เพิ่มฟิลด์ `CouponCode string` ใน `CheckoutInput` DTO

---

### 3.2 Promotion Module (`internal/promotion/`)
- **`repository.go`**:
  - `FindAllActive`: ดึงคูปองที่เปิดใช้งานและยังไม่หมดอายุ
  - `FindByCode`: ค้นหาคูปองแบบไม่สนใจตัวพิมพ์เล็ก-ใหญ่ (`UPPER(code) = UPPER($1)`)
  - `IncrementUsageTx`: ใช้ Atomic Guard: `UPDATE promotions SET used_count = used_count + 1 WHERE id = $1 AND used_count < total_quota` ป้องกัน Race Condition
- **`service.go`**:
  - `GetActivePromotions`: ดึงคูปองสำหรับหน้าบ้าน
  - `ValidateCoupon`: ตรวจสอบความถูกต้องและคำนวณยอดส่วนลดสุทธิแบบ Real-time
- **`handler.go`**:
  - `GET /api/promotions`: ดึงรายการคูปองส่วนลดที่ใช้งานได้
  - `POST /api/promotions/validate`: รับ Payload `{ code, subtotal }` แล้วส่งผลลัพธ์คำนวณกลับไปให้ผู้ใช้พรีวิวก่อนชำระเงิน

---

### 3.3 Flash Sale Module (`internal/flashsale/`)
- **`repository.go`**:
  - ใช้ **Redis In-Memory Caching (`flashsale:active`)** ดักหน้า PostgreSQL เพื่อรองรับ High-Concurrency Read ในช่วงแคมเปญ โดยมี TTL 1 นาที
  - หาก Cache Miss จะ Query ข้อมูลแคมเปญและรายการสินค้าจาก PostgreSQL แล้วบันทึกลง Redis อัตโนมัติ
- **`service.go` & `handler.go`**:
  - `GET /api/flash-sales/active`: เสิร์ฟข้อมูลแคมเปญที่กำลังจัดอยู่ปัจจุบัน

---

### 3.4 Order Checkout Integration (`internal/order/repository.go`)
ฟังก์ชัน **`CreateOrder`** ได้รับการอัปเกรดให้รองรับ Transaction ระดับ Enterprise:
1. **Flash Sale Price Override**: ตรวจสอบว่าสินค้าที่สั่งซื้ออยู่ใน Flash Sale ที่ยังไม่หมดอายุและมีโควตาคงเหลือหรือไม่ ถ้ามีจะปรับราคาเป็น `flash_price` และตัด `sold_count` ทันที
2. **Coupon Locking & Deduction**: ล็อกแถวคูปองด้วย `SELECT ... FOR UPDATE`, คำนวณส่วนลดตาม `subtotalAmount`, และบวก `used_count + 1` ใน Transaction
3. **Persist Order Header**: บันทึกคำสั่งซื้อลงตาราง `orders` พร้อมเก็บ `discount_amount` และ `coupon_id`
4. **Cache Invalidation**: ล้างแคช `products:all` และ `flashsale:active` ทันทีหลัง Commit สำเร็จ
5. **Quota Restoration on Cancel**: ในฟังก์ชัน `CancelOrderAndRestoreStock` หากคำสั่งซื้อหมดอายุ (5 นาที) ระบบจะคืนสต็อกสินค้าพร้อมกับคืนสิทธิ์คูปอง (`used_count - 1`) ให้ลูกค้าอัตโนมัติ

---

### 3.5 Dependency Injection & Routing (`cmd/api/main.go`)
- ทำการ Wire-up Promotion และ Flash Sale Modules เข้ากับ Database, Redis, และ HTTP Router อย่างเป็นระเบียบตามหลัก Clean Architecture

---

## 🧪 4. การทดสอบจริงบน Docker Environment (Verification Results)

### 4.1 ตรวจสอบความถูกต้องของซอร์สโค้ด Go
```bash
go vet ./...
go test ./...
```
**ผลลัพธ์**: ผ่าน 100% ทุกแพ็กเกจ ไม่มี Error

### 4.2 ทดสอบ API Endpoints จริง

#### 1) ตรวจสอบรายการคูปอง (`GET /api/promotions`)
```json
[
  {
    "id": 1,
    "code": "WELCOME100",
    "discount_type": "fixed",
    "discount_value": 100,
    "min_spend": 1000,
    "total_quota": 500,
    "used_count": 0
  },
  {
    "id": 2,
    "code": "NEXTGEN15",
    "discount_type": "percentage",
    "discount_value": 15,
    "min_spend": 1500,
    "max_discount": 300,
    "total_quota": 200,
    "used_count": 0
  }
]
```

#### 2) ทดสอบคำนวณคูปอง (`POST /api/promotions/validate`)
- **เคสโค้ด `WELCOME100` (ยอด 2,590)**:
  - ได้รับส่วนลด `100` $\to$ ยอดสุทธิ `2,490` บาท
- **เคสโค้ด `NEXTGEN15` (15% ของ 2,590 = 388.50 แต่ติดเพดาน 300)**:
  - ได้รับส่วนลดตามเพดาน `300` $\to$ ยอดสุทธิ `2,290` บาท

#### 3) ตรวจสอบแคมเปญ Flash Sale (`GET /api/flash-sales/active`)
```json
{
  "id": 1,
  "title": "Midnight Flash Deals ⚡",
  "items": [
    {
      "product_id": 1,
      "product_name": "Mechanical Keyboard",
      "original_price": 2590,
      "flash_price": 1990,
      "flash_stock": 5,
      "sold_count": 0
    },
    {
      "product_id": 3,
      "product_name": "Gaming Headset",
      "original_price": 1990,
      "flash_price": 1490,
      "flash_stock": 5,
      "sold_count": 0
    }
  ]
}
```

#### 4) ทดสอบ Checkout Order รวมทั้ง Flash Price และ Coupon
- สั่งซื้อสินค้า ID 1 (ปกติ 2,590) พร้อมระบุโค้ด `WELCOME100`:
  ```json
  {
    "id": 22,
    "user_id": 5,
    "total_amount": 1890,
    "discount_amount": 100,
    "coupon_id": 1,
    "status": "pending",
    "items": [
      {
        "product_id": 1,
        "product_name": "Mechanical Keyboard",
        "quantity": 1,
        "price": 1990
      }
    ]
  }
  ```
  - ✅ สินค้าคิดราคา Flash Price: `1,990` บาท (แทนราคาปกติ 2,590)
  - ✅ ส่วนลดคูปองหักเพิ่มอีก: `100` บาท
  - ✅ ยอดชำระสุทธิ: `1,890` บาท
  - ✅ โควตา Flash Sale ถูกตัด: `sold_count = 1/5`
  - ✅ โควตาคูปองถูกตัด: `used_count = 1/500`

---

## 📦 5. Git Commit Reference

```bash
cd backend
git add .
git commit -m "feat(marketing): implement Promotion coupon system and Flash Sale engine with transactional checkout" -m "- Add migration 000006 creating promotions, flash_sales, and flash_sale_items tables
- Enrich Order domain entity with DiscountAmount and CouponID, and extend CheckoutInput with CouponCode
- Implement Promotion repository, service, and HTTP handler for GET /api/promotions and POST /api/promotions/validate
- Implement Flash Sale repository with Redis caching, service, and HTTP handler for GET /api/flash-sales/active
- Integrate atomic Flash Sale pricing and coupon quota validation into order CreateOrder transaction
- Add coupon quota restoration on expired order cancellation in CancelOrderAndRestoreStock
- Update Swagger API documentation for all marketing engine endpoints"
git push
```
