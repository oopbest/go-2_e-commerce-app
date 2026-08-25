# Walkthrough 10: Interactive API Documentation ด้วย Swagger & OpenAPI

> **สรุปการเรียนรู้: การสร้างเอกสาร API แบบโต้ตอบได้ (Interactive API Documentation) ใน Go**  
> มุ่งเน้นการติดตั้ง **`swaggo/swag`**, การเขียน **Declarative Doc Annotations**, การสร้างสเปก **OpenAPI / Swagger 2.0 (`docs/`)**, และการติดตั้ง **Swagger UI** ให้ทีมงานสามารถทดลองยิง API ผ่านหน้าเว็บเบราว์เซอร์ได้ทันที

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 10

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # [ปรับปรุง] เพิ่ม General Swagger Metadata และ Route /swagger/
├── docs/                                   # [ใหม่] โฟลเดอร์สร้างโดย swag init
│   ├── docs.go                             # Go code ที่เก็บ OpenAPI Spec
│   ├── swagger.json                        # Swagger Spec ในรูปแบบ JSON
│   └── swagger.yaml                        # Swagger Spec ในรูปแบบ YAML
├── internal/
│   ├── order/
│   │   └── handler.go                      # [ปรับปรุง] เพิ่ม Doc Annotations ให้ Order Endpoints
│   ├── product/
│   │   └── handler.go                      # [ปรับปรุง] เพิ่ม Doc Annotations ให้ Product Endpoints
│   └── user/
│       └── handler.go                      # [ปรับปรุง] เพิ่ม Doc Annotations ให้ Auth Endpoints
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-10/
│       └── walkthrough-10.md
├── .env
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังการทำงานของ Swagger Documentation Generator

```
[ 1. เขียน Go Handlers + Doc Comments ]
  // @Summary      User Login
  // @Tags         Auth
  // @Param        input body domain.LoginInput true "..."
  // @Success      200  {object} domain.AuthResponse
  // @Router       /api/auth/login [post]

                      │
                      ▼
[ 2. สั่งรันคำสั่ง: swag init -g cmd/api/main.go ]
  ├── 🔍 สแกน Go Structs และ Comment ทั้งหมด
  └── 📦 สร้างไฟล์ใน docs/ (docs.go, swagger.json, swagger.yaml)

                      │
                      ▼
[ 3. เสิร์ฟหน้าเว็บผ่าน http-swagger ]
  mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

                      │
                      ▼
[ 4. ผู้ใช้งานเปิด Web Browser ]
  👉 http://localhost:8080/swagger/index.html
  - กดดู Request / Response Schema
  - ทดสอบ "Try it out" -> "Execute" ได้ทันที
  - รองรับการใส่ Bearer Token ผ่านปุ่ม Authorize 🔓
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. General API Annotations (`cmd/api/main.go`)
```go
// @title           Go E-Commerce REST API
// @version         1.0
// @description     High-performance production-ready E-Commerce Backend with JWT Auth, PostgreSQL, and Redis Cache.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
```
- กำหนดข้อมูลทั่วไปของ API ระบบ Authentication แบบ `BearerAuth` ซึ่งทำให้ปุ่ม **Authorize 🔓** ใน Swagger UI สามารถรับ JWT Token เพื่อนำไปทดสอบ Endpoint ที่ต้องการสิทธิ์ได้

### 2. Method Annotations บนแต่ละ Handler
- **`@Summary` & `@Description`**: อธิบายว่าฟังก์ชันนี้ทำอะไร
- **`@Tags`**: จัดกลุ่มของ Endpoint (เช่น `Auth`, `Products`, `Orders`)
- **`@Accept` & `@Produce`**: รูปแบบข้อมูลที่รับและส่งออก (`json`)
- **`@Security BearerAuth`**: กำหนดว่า Endpoint นี้ต้องใช้ JWT Token
- **`@Param`**: ระบุ Parameter ไม่ว่าจะเป็น `body`, `path`, หรือ `query`
- **`@Success` & `@Failure`**: ระบุ Status Code และ Model Struct ตอบกลับ

### 3. การรัน Swagger บน Containerized Production Stack
- เมื่อแก้ไขหรือสร้าง Annotation ใหม่ เพียงสั่ง:
  1. `swag init -g cmd/api/main.go`
  2. `docker compose up --build -d`
- ตัวคอนเทนเนอร์จะรัน Swagger UI พร้อมให้ทีมงานเรียกใช้งานได้ทันที

---

## 📊 4. สรุป Endpoint บน Swagger UI (`/swagger/index.html`)

| หมวดหมู่ (Tag) | Method | Endpoint | สิทธิ์การเข้าถึง |
| :--- | :---: | :--- | :---: |
| **Auth** | `POST` | `/api/auth/register` | 🌐 Public |
| **Auth** | `POST` | `/api/auth/login` | 🌐 Public |
| **Products** | `GET` | `/api/products` | 🌐 Public (Redis Cached) |
| **Products** | `GET` | `/api/products/{id}` | 🌐 Public (Redis Cached) |
| **Products** | `POST` | `/api/products` | 🔒 Admin (BearerAuth) |
| **Products** | `PUT` | `/api/products/{id}` | 🔒 Admin (BearerAuth) |
| **Products** | `DELETE`| `/api/products/{id}` | 🔒 Admin (BearerAuth) |
| **Orders** | `POST` | `/api/orders/checkout` | 🔒 User (BearerAuth) |
| **Orders** | `GET` | `/api/orders` | 🔒 User (BearerAuth) |
| **Orders** | `GET` | `/api/orders/{id}` | 🔒 User (BearerAuth) |

---

## 🏆 บทสรุป 10 Phases สมบูรณ์แบบ (Zero to Hero)

โปรเจกต์นี้กลายเป็น **Production-Grade Go Backend** เต็มรูปแบบที่พร้อมสำหรับการทำงานระดับ Enterprise แล้วครับ!
