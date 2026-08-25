# Walkthrough 11: Event-Driven & Asynchronous Background Workers (Redis + Asynq)

> **สรุปการเรียนรู้: การออกแบบสถาปัตยกรรม Event-Driven และ Asynchronous Background Task Queue ใน Go**  
> มุ่งเน้นการใช้งาน **`hibiken/asynq`**, การแยก **Task Distributor (Producer)** และ **Task Processor (Consumer)**, การส่ง **Immediate Background Jobs (Email Notification)** และ **Delayed Tasks (Order Timeout Auto-Cancellation & Stock Restoral)**

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 11

```text
ecommerce-app/
├── cmd/
│   ├── api/
│   │   └── main.go                         # [ปรับปรุง] เชื่อมต่อ Task Distributor ส่งงานเข้า Redis Queue
│   └── worker/
│       └── main.go                         # [ใหม่] Worker Server แบบแยก Process อิสระ (Worker Pool)
├── internal/
│   ├── domain/
│   │   └── order.go                        # [ปรับปรุง] เพิ่มฟังก์ชัน CancelOrderAndRestoreStock ใน Interface
│   ├── order/
│   │   ├── repository.go                   # [ปรับปรุง] Logic ยกเลิกคำสั่งซื้อและคืนสต็อกสินค้าใน Transaction
│   │   └── service.go                      # [ปรับปรุง] ส่ง Task เข้า Queue เมื่อ Checkout สำเร็จ
│   └── worker/                             # [ใหม่] โมดูลจัดการ Background Tasks
│       ├── distributor.go                  # Task Producer (ตัวส่งงานเข้าคิว Redis)
│       ├── processor.go                    # Task Consumer (ตัวประมวลผลงาน Email และ Auto-Cancel)
│       └── task.go                         # นิยาม Task Types และ Payloads
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-11/
│       └── walkthrough-11.md
├── docs/
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังสถาปัตยกรรม Event-Driven & Task Queue

```
[ Client (Checkout POST /api/orders/checkout) ]
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. API Server (cmd/api)                                     │
│    - บันทึก Order และตัดสต็อกใน Database Transaction        │
│    - เรียก TaskDistributor.DistributeTask...                │
│    - ตอบกลับ 201 Created ทันทีใน 0.05 วิ! ⚡                │
└──────────────────────┬──────────────────────────────────────┘
                       │ (Enqueue JSON Tasks)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Redis Task Queue (asynq)                                 │
│    ├── Queue: "critical" (Immediate: Email Confirmation)    │
│    └── Queue: "default"  (Delayed: Timeout Check in 1 Min)  │
└──────────────────────┬──────────────────────────────────────┘
                       │ (Workers Dequeue Tasks)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Background Worker Server (cmd/worker - Concurrency: 10)  │
│    ├── 📧 [EMAIL WORKER] ส่ง Email ยืนยันถึงผู้ซื้อทันที     │
│    └── ⏰ [TIMEOUT WORKER] หลัง 1 นาที:                      │
│         - ตรวจพบ Order ยังคงสถานะ "pending"                 │
│         - ปรับสถานะเป็น "cancelled"                         │
│         - คืนจำนวนสต็อกกลับสู่ตาราง products ใน Transaction  │
│         - ล้างแคช Redis ของ Product อัตโนมัติ 🔄           │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. การแบ่งแยก Producer และ Consumer ด้วย Clean Architecture
- **Task Distributor (`distributor.go`)**: ทำหน้าที่สร้างและส่ง Task เข้า Redis Queue อยู่ในรูปแบบ Go Interface ทำให้ Service Layer เขียน Unit Test ได้ง่ายโดยไม่ต้องพึ่ง Redis จริง
- **Task Processor (`processor.go`)**: ทำหน้าที่รับ Task และประมวลผล Business Logic จริง

### 2. Immediate Async Jobs vs Delayed Scheduled Tasks
```go
// ส่ง Email ทันที
_ = s.distributor.DistributeTaskOrderCreatedEmail(ctx, emailPayload)

// หน่วงเวลาล่วงหน้า 1 นาทีเพื่อตรวจบิลหมดอายุ
_ = s.distributor.DistributeTaskOrderTimeoutCheck(
    ctx,
    timeoutPayload,
    asynq.ProcessIn(1*time.Minute),
)
```

### 3. Fault Tolerance & Automatic Retries
- `asynq` มีระบบ Retry อัตโนมัติ (Exponential Backoff) เมื่อเกิด Network Issue หรือ Database Lock ทำให้ไม่มี Task ใดตกหล่นหรือสูญหาย

### 4. Atomic Order Cancellation & Stock Refund
- หากลูกค้ายกเลิกหรือไม่ชำระเงินภายในเวลาที่กำหนด Worker จะเปิด Database Transaction:
  1. อัปเดต `orders.status = 'cancelled'`
  2. ดึงจำนวนสินค้าจาก `order_items` มาบวกคืนเข้า `products.stock`
  3. สั่งลบแคช `products:all` ใน Redis เพื่อให้ผู้ใช้งานคนอื่นเห็นสต็อกล่าสุดทันที

---

## 📊 4. สรุปผลการทดสอบการทำงานจริง

| เหตุการณ์ (Event) | เวลาที่เกิดขึ้น | สิ่งที่ระบบทำ | ผลลัพธ์ |
| :--- | :---: | :--- | :---: |
| **User Checkout** | `00:00` | บันทึก Order ID 3 & 4 + ส่ง 2 Tasks เข้าคิว | ⚡ ตอบกลับ 201 Created ทันที |
| **Email Worker** | `00:00` | ดึง Task ส่ง Email ถึง `buyer@gmail.com` | ✅ Email Delivered ในเสี้ยววินาที |
| **Timeout Worker** | `01:00` | ตรวจพบสถานะยังเป็น `pending` | 🚨 ยกเลิกบิล & คืนสต็อกสินค้าอัตโนมัติ |
