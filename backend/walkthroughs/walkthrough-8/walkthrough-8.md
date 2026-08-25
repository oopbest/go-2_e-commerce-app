# Walkthrough 8: Automated Testing & Mocking in Go (Table-Driven Tests, Testify, httptest, และ Code Coverage)

> **สรุปการเรียนรู้: การทดสอบซอฟต์แวร์แบบอัตโนมัติตามมาตรฐานสากลใน Go**  
> มุ่งเน้นการเขียน **Table-Driven Tests**, การจำลอง **Mock Interfaces** ด้วย `testify/mock`, การทดสอบ **HTTP Presentation Layer** ด้วย `net/http/httptest`, และการวิเคราะห์ **Code Coverage Report (HTML)**

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 8

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── middleware/
│   ├── order/
│   ├── product/
│   │   ├── handler.go
│   │   ├── handler_test.go                 # [ใหม่] Unit Test สำหรับ HTTP Handlers ด้วย httptest
│   │   ├── repository_cached.go
│   │   ├── repository_postgres.go
│   │   ├── service.go
│   │   └── service_test.go                 # [ใหม่] Unit Test สำหรับ Business Logic ด้วย Mock Repo
│   └── user/
├── pkg/
│   └── security/
│       ├── jwt.go
│       ├── jwt_test.go                     # [ใหม่] Unit Test สำหรับการสร้าง/ตรวจสอบ Token
│       ├── password.go
│       └── password_test.go              # [ใหม่] Table-Driven Unit Test สำหรับ bcrypt
├── migrations/
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-8/
│       └── walkthrough-8.md
├── coverage.out                            # [ใหม่] ไฟล์รายงานผล Code Coverage
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังวงจรการทดสอบในแต่ละเลเยอร์ (Testing Pyramid)

```
[ 1. Security & Helpers Layer ]
  password_test.go & jwt_test.go ──> ทดสอบฟังก์ชันคณิตศาสตร์/การเข้ารหัส (Fastest: ~0.00s)

──────────────────────────────────────────────────────────────────────────────────────────

[ 2. Service Layer (Business Logic) ]
  service_test.go ──> NewService(MockProductRepository)
                            │
                            ├── 🧪 ทดสอบ: ตรวจสอบชื่อห้ามว่าง, ราคา > 0, สต็อก >= 0
                            └── 🚫 MockRepo.AssertNotCalled("Create") มั่นใจว่าไม่เซฟลง DB เมื่อข้อมูลผิด

──────────────────────────────────────────────────────────────────────────────────────────

[ 3. HTTP Presentation Layer ]
  handler_test.go ──> httptest.NewRequest + httptest.NewRecorder
                            │
                            ├── 🧪 ทดสอบ: Status Code 200, 201, 400, 404, 403
                            ├── 🔒 ทดสอบ: RBAC Role Check ผ่าน Context Injection
                            └── ⚡ ทำงานทั้งหมดใน Memory โดยไม่ต้องเปิด Port Network จริง
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. Table-Driven Tests ใน Go (`pkg/security/password_test.go`)
```go
testCases := []struct {
    name          string
    password      string
    checkPassword string
    shouldMatch   bool
}{
    {"Valid Password Match", "secret123", "secret123", true},
    {"Wrong Password Mismatch", "secret123", "wrong", false},
}

for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        ...
    })
}
```
- การนิยาม Test Cases เป็นตาราง (Anonymous Struct Slice) แล้ววนลูปด้วย `t.Run()` ช่วยให้อ่านง่าย เพิ่มเคสใหม่ได้สะดวก และจำแนกผลลัพธ์ได้อย่างชัดเจน

### 2. Mocking Interface ด้วย `testify/mock` (`internal/product/service_test.go`)
- สร้าง `MockProductRepository` จำลองการทำงานของ Database:
  ```go
  mockRepo.On("FindByID", 1).Return(expectedProduct, nil)
  result, err := service.GetProductByID(1)
  mockRepo.AssertExpectations(t) // ตรวจสอบว่าฟังก์ชันถูกเรียกตามเงื่อนไขเป๊ะหรือไม่
  ```

### 3. การทดสอบ HTTP API ด้วย `net/http/httptest` (`internal/product/handler_test.go`)
- **`httptest.NewRequest` & `httptest.NewRecorder`**: จำลอง HTTP Request/Response ทั้งหมดใน RAM ทดสอบ Route และ JSON Response ได้อย่างแม่นยำ
- **Context Injection ในการทดสอบสิทธิ์ RBAC**:
  ```go
  adminClaims := &security.CustomClaims{UserID: 1, Role: "admin"}
  ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminClaims)
  req = req.WithContext(ctx)
  ```

### 4. Interactive Code Coverage Analysis
- สร้างและดูรายงานความครอบคลุมของโค้ด:
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
  ```
- แสดงโค้ดแบบไฮไลต์สี: 🟢 เขียว (ทดสอบแล้ว) / 🔴 แดง (ยังไม่ได้ทดสอบ)

---

## 📊 4. สรุปผลการรัน Test Suite

```text
=== RUN   TestJWTTokenGenerationAndValidation (3 Subtests)  --- PASS
=== RUN   TestPasswordHashing (3 Subtests)                  --- PASS
=== RUN   TestCreateProduct (4 Subtests)                    --- PASS
=== RUN   TestGetProductByID (3 Subtests)                   --- PASS
=== RUN   TestHandleGetProducts (1 Subtest)                 --- PASS
=== RUN   TestHandleGetProductByID (3 Subtests)             --- PASS
=== RUN   TestHandleCreateProduct (2 Subtests)              --- PASS
---------------------------------------------------------------------
PASS: 100% of all Unit Tests Passed successfully!
```

---

## 🚀 5. ก้าวต่อไป: สู่ระบบ Multi-Stage Containerization (Phase 9)

ในบทเรียนถัดไป เราจะแพ็กเกจ Go Backend ของเราให้พร้อม Deploy ขึ้น Cloud ด้วย **Multi-Stage Dockerfile**:
1. **Builder Stage**: ใช้ `golang:alpine` ในการคอมไพล์ Binary
2. **Final Lean Stage**: คัดลอกเฉพาะ Binary ตัวเล็กๆ ไปรันบน `alpine` (ขนาด Image เล็กกว่า **20 MB**!)
3. **Docker Compose All-in-One**: รันทั้ง `api`, `postgres`, และ `redis` ขึ้นมาพร้อมกันในคำสั่งเดียว
