# Walkthrough 5: ระบบ Cart & Order Checkout (Database Transactions & Concurrency Locking)

> **สรุปการเรียนรู้: หัวใจสำคัญทางวิศวกรรมของ E-Commerce Backend**  
> มุ่งเน้นการจัดการธุรกรรมทางการเงิน (Database Transactions), การป้องกันการแย่งซื้อสินค้า (Race Condition) ด้วย **Pessimistic Locking (`SELECT ... FOR UPDATE`)**, และการบันทึกประวัติคำสั่งซื้อแบบหลายตาราง

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 5

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # เชื่อมต่อ Order Module เข้ากับ Auth Middleware
├── internal/
│   ├── database/
│   │   └── postgres.go
│   ├── domain/
│   │   ├── order.go                        # [ใหม่] Order & OrderItem Entities, DTOs, Interfaces
│   │   ├── product.go
│   │   └── user.go
│   ├── middleware/
│   │   └── auth.go
│   ├── order/                              # [ใหม่] โมดูลจัดการคำสั่งซื้อ
│   │   ├── handler.go                      # HTTP: /api/orders/checkout, /api/orders, /api/orders/{id}
│   │   ├── repository.go                   # Data: db.BeginTx + SELECT FOR UPDATE + ตาราง orders/order_items
│   │   └── service.go                      # Logic: Business Validation & Data Access Rules
│   ├── product/
│   └── user/
├── pkg/
│   └── security/
├── migrations/
│   └── init.sql                            # เพิ่มตาราง orders และ order_items
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── walkthrough-2/
│   ├── walkthrough-3/
│   ├── walkthrough-4/
│   └── walkthrough-5/
│       └── walkthrough-5.md
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังวงจรการทำงานของ Checkout Transaction (Concurrency Safe)

```
Client (Customer) ──(POST /api/orders/checkout + Bearer Token)
                                │
                                ▼
               ┌─────────────────────────────────┐
               │ 1. middleware.AuthMiddleware    │ ──> ดึง UserID จาก JWT Context
               └────────────────┬────────────────┘
                                │
                                ▼
               ┌─────────────────────────────────┐
               │ 2. order/handler & service      │ ──> Validate ตะกร้าไม่ว่าง, จำนวน > 0
               └────────────────┬────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. order/repository.CreateOrder (Database Transaction: BeginTx)             │
│                                                                             │
│   [ สินค้าชิ้นที่ 1, 2, ... ]                                               │
│       │                                                                     │
│       ├── 🔒 SELECT name, price, stock FROM products WHERE id = $1          │
│       │       FOR UPDATE (ล็อกแถวสินค้าใน PostgreSQL ทันที)                 │
│       │                                                                     │
│       ├── ❓ ตรวจสอบ: สต็อกพอหรือไม่?                                       │
│       │       ├── ❌ ไม่พอ: tx.Rollback() -> คืน ErrInsufficientStock       │
│       │       └── ✅ พอ: UPDATE products SET stock = stock - qty            │
│       │                                                                     │
│       └── ➕ รวมยอดเงิน: totalAmount += (price * qty)                       │
│                                                                             │
│   [ บันทึกคำสั่งซื้อ ]                                                      │
│       ├── 📝 INSERT INTO orders (user_id, total_amount, status)             │
│       └── 📝 INSERT INTO order_items (order_id, product_id, qty, price)     │
│                                                                             │
│   [ จบ Transaction ]                                                        │
│       └── 🚀 tx.Commit() (ยืนยันข้อมูลและปลดล็อกแถวสินค้า)                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                   [ คืน Order 201 Created ]
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. ปัญหา Race Condition ในระบบขายสินค้า
- หากมีผู้ใช้ 2 คนกดซื้อสินค้าชิ้นสุดท้ายพร้อมกัน ถ้าอ่านค่าสต็อกพร้อมกัน ทั้งคู่จะเห็น `stock = 1` และตัดสต็อกสำเร็จทั้งคู่ ส่งผลให้สต็อกติดลบ (`stock = -1`)
- **การแก้ปัญหาด้วย `SELECT ... FOR UPDATE`**: PostgreSQL จะทำการล็อกแถวของสินค้านั้นไว้ทันที ทำให้คำขอที่ 2 ต้องรอคิวจนกว่าคำขอแรกจะ Commit หรือ Rollback เสร็จสิ้น

### 2. รูปแบบ `defer tx.Rollback()` ใน Go
```go
tx, err := r.db.BeginTx(ctx, nil)
if err != nil { return nil, err }
defer tx.Rollback() // ปลอดภัย 100%: ถ้ามี error กลางทาง จะ rollback ทันที

// ... ทำงานต่างๆ ...

if err := tx.Commit(); err != nil {
    return nil, err
}
return order, nil
```
- ใน Go หากเราเรียก `tx.Commit()` สำเร็จ คำสั่ง `tx.Rollback()` ใน `defer` จะไม่ส่งผลเสียใดๆ แต่หากเกิด `return` กลางคันเพราะสต็อกหมดหรือ Error คำสั่ง `defer` จะยกเลิกการเปลี่ยนแปลงทั้งหมด คืนสภาพ Database ให้สมบูรณ์เสมอ

### 3. การ Snapshot ราคา ณ วันที่ซื้อ (`price_at_purchase`)
- ในตาราง `order_items` เราเก็บคอลัมน์ `price` แยกต่างหากจากตาราง `products` เพื่อให้ราคาในประวัติการสั่งซื้อไม่เปลี่ยนแปลง แม้ในอนาคตผู้ขายจะปรับราคาสินค้าขึ้นหรือลงก็ตาม

### 4. การดึง UserID จาก JWT Context (ป้องกัน ID Spoofing)
```go
claims, ok := middleware.GetUserClaims(r)
order, err := h.service.Checkout(r.Context(), claims.UserID, input)
```
- Client ไม่จำเป็นต้องส่ง `user_id` มาใน Request Body เลย เพราะเราดึงจาก Token ที่ยืนยันตัวตนแล้วโดยตรง

---

## 📊 4. สรุป Endpoint ของระบบสั่งซื้อ

| Endpoint | Method | สิทธิ์การเข้าถึง | รายละเอียด |
| :--- | :---: | :---: | :--- |
| `/api/orders/checkout` | `POST` | 🔒 **Logged-in User** | สั่งซื้อสินค้าในตะกร้า พร้อมตัดสต็อกแบบ Transaction |
| `/api/orders` | `GET` | 🔒 **Logged-in User** | ดูประวัติคำสั่งซื้อ (Customer เห็นเฉพาะของตนเอง / Admin เห็นทั้งหมด) |
| `/api/orders/{id}` | `GET` | 🔒 **Logged-in User** | ดูรายละเอียดบิลตาม ID พร้อมรายการสินค้าและราคา |

---

## 🚀 5. ก้าวต่อไป: สู่ระบบ Production-Ready (Phase 6)

ในบทเรียนถัดไป เราจะยกระดับ API ของเราสู่มาตรฐาน Production:
1. **Graceful Shutdown**: ปิด Server อย่างนุ่มนวล รอ Request ที่ค้างอยู่ทำงานเสร็จก่อน ไม่ตัด Connection กลางคัน
2. **Structured Logging ด้วย `log/slog`**: ระบบบันทึก Log แบบ JSON ตามมาตรฐานใหม่ของ Go 1.21+
3. **Configuration ด้วย Environment Variables**: แยก Secret และ Config ออกจาก Source Code
