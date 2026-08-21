# Walkthrough 4: ระบบ Authentication, การเข้ารหัสด้วย bcrypt, JWT Security และ Role-Based Access Control (RBAC)

> **สรุปการเรียนรู้: การสร้างระบบความปลอดภัยมาตรฐานระดับสากลใน Go Backend**  
> มุ่งเน้นการจัดการสิทธิ์ผู้ใช้งาน, การแฮชรหัสผ่านด้วย `bcrypt`, การออกและตรวจสอบ `JWT Token`, การออกแบบ `Auth Middleware`, และการฝังข้อมูล User ลงใน `context.Context`

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 4

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # รวมร่าง Auth & Product Modules พร้อม Middleware
├── internal/
│   ├── database/
│   │   └── postgres.go
│   ├── domain/
│   │   ├── product.go
│   │   └── user.go                         # [ใหม่] User Entity, DTOs, Errors, และ Interfaces
│   ├── middleware/
│   │   └── auth.go                         # [ใหม่] AuthMiddleware & RequireRole
│   ├── product/
│   │   ├── handler.go                      # [ปรับปรุง] ป้องกัน Endpoint ด้วย Middleware
│   │   ├── repository_postgres.go
│   │   └── service.go
│   └── user/                               # [ใหม่] โมดูล User & Auth
│       ├── handler.go                      # HTTP: /api/auth/register, /api/auth/login
│       ├── repository.go                   # Data: จัดการตาราง users ใน PostgreSQL
│       └── service.go                      # Logic: ตรวจสอบข้อมูล, แฮชรหัสผ่าน, ออก JWT
├── pkg/
│   └── security/                           # [ใหม่] Generic Security Tools
│       ├── jwt.go                          # สร้างและตรวจสอบ JSON Web Token (HS256)
│       └── password.go                     # เข้ารหัสและเปรียบเทียบรหัสผ่านด้วย bcrypt
├── migrations/
│   └── init.sql                            # เพิ่มคำสั่ง CREATE TABLE users
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── walkthrough-2/
│   ├── walkthrough-3/
│   └── walkthrough-4/
│       └── walkthrough-4.md
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังระบบความปลอดภัยและวงจรการทำงาน (Auth Lifecycle)

```
[ 1. สมัครสมาชิก / เข้าสู่ระบบ ]
Client ──(POST /api/auth/login)──> user/handler ──> user/service ──> user/repository (PostgreSQL)
                                         │                   │
                                         │                   ├──> security.CheckPasswordHash()
                                         │                   └──> security.GenerateToken()
                                         ▼
                               [ รับ JWT Token กลับไป ]

──────────────────────────────────────────────────────────────────────────────────────────

[ 2. เรียกดูสินค้า (Public) ]
Client ──(GET /api/products)──> product/handler ──> คืนรายการสินค้าทันที (ไม่ต้องมี Token)

──────────────────────────────────────────────────────────────────────────────────────────

[ 3. สร้าง/แก้ไข/ลบสินค้า (Protected: Admin Only) ]
Client ──(POST /api/products + Header: Bearer <Token>)
                 │
                 ▼
┌───────────────────────────────────────────────────────────────┐
│ 1. middleware.AuthMiddleware                                  │
│    - ตรวจสอบความถูกต้องและวันหมดอายุของ JWT Token             │
│    - ดึง Claims (UserID, Email, Role) ฝังลงใน r.Context()     │
└──────────────────────────────┬────────────────────────────────┘
                               │
                               ▼
┌───────────────────────────────────────────────────────────────┐
│ 2. middleware.RequireRole("admin")                            │
│    - อ่าน Role จาก Context: ถ้าเป็น 'admin' -> ให้ผ่าน        │
│    - ถ้าเป็น 'customer' -> ตัดจบด้วย 403 Forbidden            │
└──────────────────────────────┬────────────────────────────────┘
                               │
                               ▼
┌───────────────────────────────────────────────────────────────┐
│ 3. product/handler.handleCreateProduct                        │
│    - ดำเนินการสร้างสินค้าลง Database                          │
└───────────────────────────────────────────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. `json:"-"` Struct Tag (Data Sanitization)
```go
type User struct {
    ID           int       `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // ป้องกันไม่ให้ส่ง Hash ออกไปใน JSON Response เด็ดขาด
    Role         string    `json:"role"`
}
```
- ใน Go การใส่ tag `json:"-"` จะทำให้ตัวแปลง JSON ข้ามฟิลด์นี้เสมอ ช่วยการันตีว่าจะไม่มีทางที่ Password Hash จะหลุดออกไปใน Response แม้จะส่ง Struct `User` ทั้งตัวกลับไป

### 2. Password Hashing ด้วย `bcrypt` (`pkg/security/password.go`)
- **ห้ามเก็บ Plain Text Password เด็ดขาด**: `bcrypt` จะทำการสุ่ม Salt และ Hash รหัสผ่านด้วย Cost Factor (Default = 10)
- `bcrypt.CompareHashAndPassword()` ทำการเปรียบเทียบรหัสผ่านแบบป้องกัน Timing Attack

### 3. JWT Claims & Signing (`pkg/security/jwt.go`)
- สร้าง `CustomClaims` ที่ฝัง `UserID`, `Email`, `Role` ไว้ใน Payload ของ Token
- ลงลายมือชื่อดิจิทัลด้วยอัลกอริทึม **HMAC-SHA256 (HS256)** พร้อมกำหนดวันหมดอายุ (`ExpiresAt`)

### 4. การจัดการ Database Error ใน PostgreSQL (`internal/user/repository.go`)
- ดักจับกรณีผู้ใช้กรอก Email ซ้ำด้วยการตรวจสอบ PostgreSQL Error Code **`23505`** (`unique_violation`):
  ```go
  var pqErr *pq.Error
  if errors.As(err, &pqErr) && pqErr.Code == "23505" {
      return nil, domain.ErrUserAlreadyExists // แมปเป็น HTTP 409 Conflict
  }
  ```

### 5. Middleware Pattern ใน Go (`internal/middleware/auth.go`)
- **Higher-Order Functions**: การนำฟังก์ชันมา Wrap ครอบ `http.HandlerFunc` ตัวถัดไป
- **Context Injection**: ใช้ `context.WithValue(r.Context(), UserContextKey, claims)` เพื่อส่งต่อข้อมูลผู้ใช้ไปยัง Handler ตัวถัดไปอย่างปลอดภัย
- **Type-safe Context Key**: ประกาศ `type contextKey string` เพื่อป้องกันไม่ให้ Key ของแพ็กเกจชนกับแพ็กเกจอื่นในระบบ

---

## 📊 4. สรุป Route & สิทธิ์การเข้าถึง (API Security Matrix)

| Endpoint | Method | สิทธิ์การเข้าถึง (Access Level) | รายละเอียด |
| :--- | :---: | :---: | :--- |
| `/api/auth/register` | `POST` | 🌐 **Public** | สมัครสมาชิกใหม่ (Default Role: `customer`) |
| `/api/auth/login` | `POST` | 🌐 **Public** | เข้าสู่ระบบและรับ JWT Token |
| `/api/products` | `GET` | 🌐 **Public** | ดึงรายการสินค้าทั้งหมด |
| `/api/products/{id}` | `GET` | 🌐 **Public** | ดึงข้อมูลสินค้าตาม ID |
| `/api/products` | `POST` | 🔒 **Admin Only** | สร้างสินค้าใหม่ (ต้องมี Bearer Token สิทธิ์ admin) |
| `/api/products/{id}` | `PUT` | 🔒 **Admin Only** | แก้ไขข้อมูลสินค้า |
| `/api/products/{id}` | `DELETE` | 🔒 **Admin Only** | ลบสินค้าออกจากระบบ |

---

## 🚀 5. ก้าวต่อไป: สู่ระบบ Cart & Order Checkout (Phase 5)

ในบทเรียนถัดไป เราจะเข้าสู่แก่นแท้ทางวิศวกรรมของระบบ E-Commerce:
1. **ระบบตะกร้าสินค้า (Cart & Cart Items)**
2. **ระบบสั่งซื้อ (Order & Order Items)**
3. **Database Transaction (`db.BeginTx`)**: รวมการสร้าง Order, หักเงิน, และตัดสต็อกไว้ใน Transaction เดียวกัน (All-or-Nothing)
4. **Concurrency & Race Condition Prevention**: จัดการตัดสต็อกสินค้าชิ้นเดียวกันที่มีคนกดซื้อพร้อมกัน 100 คน ด้วยเทคนิค **Pessimistic Locking (`SELECT ... FOR UPDATE`)**
