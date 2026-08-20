# Walkthrough 1: พื้นฐาน Go RESTful API & In-Memory CRUD

> **สรุปการเรียนรู้ก่อนก้าวสู่ Database & Clean Architecture**  
> เปรียบเทียบมุมมองระหว่าง **Node.js (Express) / PHP** กับ **Go (Golang)**

---

## 📌 1. ภาพรวมสิ่งที่เราสร้างขึ้นใน `main.go`

เราได้สร้าง RESTful API สำหรับระบบจัดการสินค้า (E-Commerce Products) โดยใช้เฉพาะ **Go Standard Library (`net/http`)** แบบไม่พึ่งพา External Framework ใดๆ:

| Method | Endpoint | หน้าที่ | HTTP Status สำเร็จ |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | ตรวจสอบสถานะการทำงานของเซิร์ฟเวอร์ | `200 OK` |
| `GET` | `/api/products` | ดึงรายการสินค้าทั้งหมด | `200 OK` |
| `GET` | `/api/products/{id}` | ดึงข้อมูลสินค้าตาม ID (URL Path Parameter) | `200 OK` |
| `POST` | `/api/products` | สร้างสินค้าใหม่ (พร้อม Validation) | `201 Created` |
| `PUT` | `/api/products/{id}` | อัปเดตข้อมูลสินค้า (พร้อม Validation) | `200 OK` |
| `DELETE` | `/api/products/{id}` | ลบสินค้าออกจากระบบ | `204 No Content` |

---

## 🧠 2. สรุป Concept สำคัญที่ได้เรียนรู้

### 1. Data Contract & Visibility (Public vs Private)
- **Structs**: ทำหน้าที่คล้าย TypeScript Interface + Class
- **Visibility**: ดูที่ตัวอักษรแรกของชื่อฟิลด์/ฟังก์ชัน
  - ตัวพิมพ์ใหญ่ (`Name`, `Price`) = **Public / Exported** (แพ็กเกจอื่นและ JSON Encoder มองเห็น)
  - ตัวพิมพ์เล็ก (`name`, `price`) = **Private / Unexported** (ใช้ได้เฉพาะภายในแพ็กเกจ)
- **Struct Tags (`json:"name"`)**: กำหนดชื่อคีย์เวลาแปลงข้อมูลไป-กลับเป็น JSON

### 2. Pointer (`&`) ในการ Decode ข้อมูล
```go
var req CreateProductRequest
json.NewDecoder(r.Body).Decode(&req) // ส่ง Memory Address ผ่าน &
```
- Go ส่งค่าแบบ **Pass-by-Value (Copy ค่า)** เป็นค่าเริ่มต้น
- การใส่ `&req` (Pointer) คือการส่ง Address เพื่อให้ฟังก์ชัน `Decode` แก้ไขค่าลงในตัวแปรต้นทางได้โดยตรง (คล้าย `&$data` ใน PHP)

### 3. Concurrency Safety ด้วย `sync.RWMutex`
- **Node.js**: ทำงานบน Single Thread ไม่มีการแย่งเขียน Array พร้อมกัน
- **Go**: ทุกๆ HTTP Request ทำงานบน **Goroutine (Lightweight Thread)** อิสระ จึงต้องควบคุมการเข้าถึงข้อมูลร่วมกัน:
  - `mu.RLock()` / `mu.RUnlock()`: หลาย Goroutine อ่านข้อมูลพร้อมกันได้ (ห้ามเขียน)
  - `mu.Lock()` / `mu.Unlock()`: อนุญาตให้ Goroutine เดียวเข้าเขียนข้อมูลได้ (ผู้อื่นต้องรอ)

### 4. การจัดการ Cleanup ด้วย `defer`
```go
mu.Lock()
defer mu.Unlock() // จะถูกเรียกทำงานเสมอเมื่อฟังก์ชันจบ ไม่ว่าจะ return สำเร็จหรือเกิด error
```

### 5. Fail-Fast & Validation Pattern
- ตรวจสอบความถูกต้องของ Input (เช่น ID, Request Body, ราคา, สต็อก) **ก่อนเรียก `mu.Lock()`** เสมอ เพื่อไม่ให้คำขอที่ผิดพลาดไปบล็อก Goroutine ตัวอื่นโดยไม่จำเป็น

### 6. การจัดการ Slice (Array ใน Go)
1. **Value Copy ใน Loop**:
   - `for i, p := range products` ตัวแปร `p` เป็นแค่ Copy ถ้าต้องการแก้ค่าจริง ต้องแก้ผ่าน Index เช่น `products[i].Name = req.Name`
2. **การลบ Element ออกจาก Slice**:
   - **ท่า `copy` + Clear Zero Value**:
     ```go
     copy(products[i:], products[i+1:])
     products[len(products)-1] = Product{} // เคลียร์ช่องท้าย (ถ้าเป็น Pointer []*Product ให้ใช้ = nil)
     products = products[:len(products)-1] // หดขนาด slice
     ```
   - **ท่า `append` Slice Cut**:
     ```go
     products = append(products[:i], products[i+1:]...)
     ```

---

## 🗺️ 3. ตารางเทียบเปรียบเทียบ (Cheat Sheet)

| หน้าที่ | Node.js (Express) | PHP | Go (Golang) |
| :--- | :--- | :--- | :--- |
| **Data Type** | `interface Product {}` | `class Product {}` | `type Product struct {}` |
| **Path Param** | `req.params.id` | `$request->route('id')` | `r.PathValue("id")` (Go 1.22+) |
| **Decode Body**| `req.body` | `json_decode(...)` | `json.NewDecoder(r.Body).Decode(&req)` |
| **Response** | `res.status(200).json(...)` | `response()->json(..., 200)` | `json.NewEncoder(w).Encode(...)` |
| **Delete Status** | `res.status(204).send()` | `response()->noContent()` | `w.WriteHeader(http.StatusNoContent)` |
| **Error Handling** | `try / catch` | `try / catch` | `if err != nil { return }` |

---

## 🚀 4. ก้าวต่อไป: สู่ระบบจริง (Next Milestone)

ในบทเรียนถัดไป เราจะย้ายข้อมูลจาก RAM ไปเก็บใน **Database จริง (PostgreSQL)** และจัดระเบียบโค้ดตาม **Clean Architecture**:

```
[ HTTP Request ]
       │
       ▼
┌──────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌────────────┐
│   Handler    │ ──> │   Service    │ ──> │   Repository    │ ──> │ PostgreSQL │
│ (Controller) │     │(BusinessLogic│     │(Database Query) │     │ (Database) │
└──────────────┘     └──────────────┘     └─────────────────┘     └────────────┘
```

1. ออกแบบ Database Table: `users`, `categories`, `products`, `orders`
2. จัดโครงสร้างโฟลเดอร์ตามมาตรฐาน Go (`cmd/api`, `internal/domain`, `internal/product`)
3. เชื่อมต่อ Database Connection Pool ด้วย `database/sql` หรือ Driver สมัยใหม่
