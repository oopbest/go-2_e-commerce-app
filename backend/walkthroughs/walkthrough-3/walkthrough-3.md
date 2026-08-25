# Walkthrough 3: การเชื่อมต่อฐานข้อมูลจริงด้วย PostgreSQL + Docker Compose และ SQL Repository

> **สรุปการเรียนรู้: จาก In-Memory สู่ Persistent Database ด้วย PostgreSQL**  
> มุ่งเน้นการใช้งาน Docker Compose, Connection Pooling (`database/sql`), ไวยากรณ์ SQL ใน Go, และการสลับ Data Layer อย่างไร้รอยต่อผ่าน Interface

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 3

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # เชื่อมต่อ Database และสลับใช้ Postgres Repository
├── internal/
│   ├── database/
│   │   └── postgres.go                     # จัดการ Connection Pool ของ PostgreSQL
│   ├── domain/
│   │   └── product.go                      # Interfaces & Entities (ไม่ถูกแก้ไขเลย)
│   └── product/
│       ├── handler.go                      # HTTP Controller (ไม่ถูกแก้ไขเลย)
│       ├── service.go                      # Business Logic (ไม่ถูกแก้ไขเลย)
│       ├── repository.go                   # In-Memory Repository เดิม (ยังคงอยู่)
│       └── repository_postgres.go          # [ใหม่] PostgreSQL SQL Implementation
├── migrations/
│   └── init.sql                            # สคริปต์สร้างตาราง products และข้อมูลเริ่มต้น
├── docker-compose.yml                      # คอนฟิกเปิด PostgreSQL 16
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── walkthrough-2/
│   └── walkthrough-3/
│       └── walkthrough-3.md
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังการทำงานและการสลับ Data Layer (The Power of Interfaces)

```
                       [ HTTP Request ]
                              │
                              ▼
                ┌───────────────────────────┐
                │        Handler Layer      │ (ไม่มีการเปลี่ยนแปลง)
                └─────────────┬─────────────┘
                              │
                              ▼
                ┌───────────────────────────┐
                │        Service Layer      │ (ไม่มีการเปลี่ยนแปลง)
                └─────────────┬─────────────┘
                              │
             ┌────────────────┴────────────────┐
             │ สลับการทำงานผ่าน Interface       │
             ▼                                 ▼
┌───────────────────────────┐     ┌───────────────────────────┐
│ inMemoryRepository (เดิม)  │     │ postgresRepository (ใหม่) │
│ (เก็บใน RAM ด้วย Mutex)   │     │ (ติดต่อ PostgreSQL 16)    │
└───────────────────────────┘     └─────────────┬─────────────┘
                                                │
                                                ▼
                                  ┌───────────────────────────┐
                                  │ PostgreSQL Container (:5432│
                                  │ (ecommerce_db / Persistent│
                                  └───────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. Docker Compose สำหรับ Database
- ใช้ `postgres:16-alpine` พร้อมตั้งค่า `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- เชื่อมต่อ **Named Volume (`postgres_data`)** เพื่อให้ข้อมูลถูกบันทึกลงดิสก์อย่างถาวร (ไม่หายเมื่อ Container รีสตาร์ท)

### 2. Blank Identifier Import (`_ "github.com/lib/pq"`)
- ใน Go การ `import` โดยใส่ `_` นำหน้า หมายถึง **"ต้องการให้รันฟังก์ชัน `init()` ในแพ็กเกจนั้นโดยที่เราไม่ต้องเรียกใช้ชื่อแพ็กเกจตรงๆ"**
- สำหรับ Driver ของ Database จะทำหน้าที่ลงทะเบียน Driver ชื่อ `"postgres"` เข้าสู่ระบบ `database/sql` มาตรฐาน

### 3. Connection Pooling Best Practices (`internal/database/postgres.go`)
- `sql.Open()` จะยังไม่เชื่อมต่อ Network ทันที แต่เป็นการสร้าง Connection Pool
- การตั้งค่าที่จำเป็นสำหรับ Production:
  - `db.SetMaxOpenConns(25)`: จำกัดจำนวน Connection สูงสุด
  - `db.SetMaxIdleConns(25)`: จำนวน Connection ที่พักรอไว้ใน Pool
  - `db.SetConnMaxLifetime(5 * time.Minute)`: อายุสูงสุดของแต่ละ Connection ก่อนทำลายแล้วสร้างใหม่
  - `db.PingContext(ctx)`: ตรวจสอบความพร้อมของ Database ก่อนเริ่มรันแอป

### 4. การจัดการคำสั่ง SQL ใน Go (`internal/product/repository_postgres.go`)

| คำสั่ง / ฟังก์ชัน | การใช้งาน | ข้อควรระวัง / Best Practice |
| :--- | :--- | :--- |
| **Placeholder `$1, $2`** | ป้องกัน **SQL Injection** | PostgreSQL ใช้ `$1, $2` (แทนที่ `?` ใน MySQL) |
| **`QueryContext`** | ใช้กับคำสั่งที่ได้หลายแถว (`SELECT ...`) | ต้องมี **`defer rows.Close()`** เสมอเพื่อคืน Connection เข้า Pool |
| **`QueryRowContext`** | ใช้กับคำสั่งที่ได้แถวเดียว (`SELECT ... WHERE id = $1`) | ตรวจจับกรณีไม่พบข้อมูลด้วย `errors.Is(err, sql.ErrNoRows)` |
| **`RETURNING id, created_at`** | ดึงค่า Primary Key หรือฟิลด์ที่ DB เจนให้อัตโนมัติ | ช่วยลดการยิง Query ซ้ำซ้อนหลัง `INSERT` / `UPDATE` |
| **`ExecContext`** | ใช้กับคำสั่งที่ไม่คืนแถว (`DELETE`, `UPDATE`) | ตรวจสอบจำนวนแถวที่ได้รับผลกระทบด้วย `res.RowsAffected()` |
| **`context.WithTimeout`** | กำหนดระยะเวลา Timeout ในทุก Query (เช่น 3 วินาที) | ป้องกัน Goroutine แขวนค้างเมื่อ Database มีปัญหา |

---

## 💡 4. บทพิสูจน์พลังของ Clean Architecture

ใน `cmd/api/main.go` เราเพียงแค่:
```go
// 1. เปิด DB Connection
db, _ := database.NewPostgresDB(dbCfg)
defer db.Close()

// 2. สลับ Repository ส่งเข้าไปใน Service
productRepo := product.NewPostgresRepository(db)
productService := product.NewService(productRepo)
productHandler := product.NewHandler(productService)
```

> 🎯 **ผลลัพธ์:** โค้ดในส่วน **Business Rules (`service.go`)** และ **HTTP Handlers (`handler.go`)** ไม่ต้องเปลี่ยนแปลงแม้แต่ตัวอักษรเดียว!

---

## 🚀 5. ก้าวต่อไป: สู่ระบบ Authentication & Users (Phase 4)

ในบทเรียนถัดไป เราจะขยายระบบ E-Commerce สู่การจัดการผู้ใช้งาน:
1. ออกแบบตาราง `users` (id, email, password_hash, role, created_at)
2. ทำระบบ **User Registration** และเข้ารหัสผ่านด้วย **`bcrypt`**
3. ทำระบบ **User Login** และออก **`JWT (JSON Web Token)`**
4. สร้าง **Auth Middleware** เพื่อป้องกัน Endpoint ที่ต้องการสิทธิ์ (เช่น เฉพาะ Admin หรือ Seller เท่านั้นที่เพิ่ม/แก้ไข/ลบสินค้าได้)
