# Walkthrough 2: การจัดโครงสร้างแบบ Clean Architecture & Dependency Injection ใน Go

> **สรุปการเรียนรู้: จาก Single File (`main.go`) สู่สถาปัตยกรรมระดับ Production**  
> มุ่งเน้นการแยกหน้าที่ (Separation of Concerns), Go Interfaces, และ Dependency Injection

---

## 🏗️ 1. โครงสร้างโปรเจกต์ (Standard Go Project Layout)

เราได้แยกโค้ดจากไฟล์เดียวออกเป็นชั้นๆ ตามมาตรฐานของ Go:

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go          # Composition Root (จุดประกอบชิ้นส่วน & รันเซิร์ฟเวอร์)
├── internal/
│   ├── domain/              # แกนกลางของระบบ (Entities, DTOs, Errors, Interfaces)
│   │   └── product.go
│   └── product/             # ฟีเจอร์สินค้า
│       ├── handler.go       # HTTP Presentation Layer (Controller)
│       ├── service.go       # Business Logic & Validation Layer
│       └── repository.go    # Data Access Layer (In-Memory Store)
├── walkthroughs/
│   ├── walkthrough-1.md
│   └── walkthrough-2/
│       └── walkthrough-2.md
├── go.mod
```

---

## 🔄 2. แผนผังการทำงานและการส่งต่อข้อมูล (Data Flow)

```
[ HTTP Request: GET /api/products/1 ]
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Handler Layer (internal/product/handler.go)              │
│    - รับ HTTP Request / Path Parameters                     │
│    - เรียก: service.GetProductByID(id)                      │
│    - แปลง Error -> HTTP Status (เช่น 404 Not Found)         │
│    - ส่ง HTTP Response เป็น JSON                            │
└──────────────────────────────┬──────────────────────────────┘
                               │ เรียกผ่าน domain.ProductService (Interface)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Service Layer (internal/product/service.go)              │
│    - ตรวจสอบความถูกต้องทางธุรกิจ (Business Validation)       │
│      (เช่น id > 0, ชื่อห้ามว่าง, ราคา > 0, สต็อก >= 0)       │
│    - เรียก: repo.FindByID(id)                               │
└──────────────────────────────┬──────────────────────────────┘
                               │ เรียกผ่าน domain.ProductRepository (Interface)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Repository Layer (internal/product/repository.go)        │
│    - จัดการอ่าน/เขียนข้อมูลจริง (ปัจจุบัน: In-Memory + Mutex)│
│    - คืนค่า Entity หรือ domain.ErrProductNotFound           │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญของ Go ที่ได้เรียนรู้ใน Phase นี้

### 1. ทำไมต้องมีโฟลเดอร์ `internal/`?
- ในภาษา Go โค้ดที่อยู่ในโฟลเดอร์ชื่อ `internal/` จะถูก **ป้องกันโดย Go Compiler** ไม่ให้โปรเจกต์ภายนอกสามารถ `import` ไปใช้งานได้
- ช่วยรักษาความปลอดภัยและความเป็นส่วนตัวของ Business Logic ภายในระบบ

### 2. ทำไมต้องเริ่มที่ `internal/domain`? (Domain-Centric)
- **Domain คือหัวใจของระบบ**: ไม่ควรมี Dependency กับ HTTP Framework หรือ Database ใดๆ
- เก็บ **Entity Struct**, **DTOs Input**, **Domain Errors**, และ **Interfaces (สัญญาข้อตกลง)**

### 3. Implicit Interfaces (หัวใจของความยืดหยุ่นใน Go)
- ใน Go **ไม่ต้องมีคีย์เวิร์ด `implements`**
- ตราบใดที่ Struct มีฟังก์ชันครบตามที่ Interface กำหนด Go จะถือว่า Struct นั้นผ่านสัญญาทันที:
  ```go
  // ใน domain/product.go
  type ProductRepository interface {
      FindAll() ([]Product, error)
      ...
  }

  // ใน product/repository.go
  type inMemoryRepository struct { ... }
  func (r *inMemoryRepository) FindAll() ([]domain.Product, error) { ... }
  // -> inMemoryRepository กลายเป็น ProductRepository โดยอัตโนมัติ!
  ```

### 4. Dependency Injection (DI) ผ่าน Constructor Functions
- แทนที่จะสร้าง Repository ภายใน Service โดยตรง เราส่ง (Inject) เข้ามาทาง Parameter:
  ```go
  func NewService(repo domain.ProductRepository) domain.ProductService {
      return &service{repo: repo}
  }
  ```
- **ข้อดี:** ทำให้เราสามารถสลับ In-Memory Repo เป็น PostgreSQL Repo หรือ Mock Repo สำหรับทำ Unit Test ได้อย่างอิสระ

### 5. Error Sentinel & `errors.Is()`
- ใช้ `errors.Is(err, domain.ErrProductNotFound)` ใน Handler เพื่อตรวจจับ Error ชนิดเฉพาะจาก Domain และแมปเป็น `404 Not Found` ได้อย่างแม่นยำ

### 6. Composition Root ใน `cmd/api/main.go`
- ทำหน้าที่เป็นจุดต่อสายไฟเพียงจุดเดียว:
  ```go
  repo := product.NewInMemoryRepository()
  service := product.NewService(repo)
  handler := product.NewHandler(service)
  handler.RegisterRoutes(mux)
  ```

---

## 📊 4. สรุปหน้าที่ของแต่ละ Layer

| Layer | ไฟล์ | หน้าที่หลัก | สิ่งที่เรียกใช้ (Dependency) |
| :--- | :--- | :--- | :--- |
| **Domain** | `internal/domain/product.go` | นิยาม Entity, DTO, Error, Interface | ไม่มี (Pure Go) |
| **Repository** | `internal/product/repository.go` | คุยกับ Data Store (RAM / DB) | `domain` |
| **Service** | `internal/product/service.go` | ตรวจสอบ Business Rules / Logic | `domain.ProductRepository` |
| **Handler** | `internal/product/handler.go` | แปลง HTTP $\leftrightarrow$ Go Struct | `domain.ProductService` |
| **Entry Point** | `cmd/api/main.go` | ประกอบชิ้นส่วนและเปิดพอร์ต | `internal/product` |

---

## 🚀 5. ก้าวต่อไป: สู่ Database จริง (PostgreSQL + Docker Compose)

ในขั้นตอนถัดไป เราจะนำสิ่งที่วางรากฐานไว้มาใช้งานจริง:
1. เปิด **PostgreSQL** ผ่าน `docker-compose.yml`
2. สร้างตาราง `products` ด้วยคำสั่ง SQL
3. สร้าง `repository_postgres.go` ที่ implement `domain.ProductRepository`
4. สลับใน `cmd/api/main.go` เพียงบรรทัดเดียวเพื่อต่อ Database จริง!
