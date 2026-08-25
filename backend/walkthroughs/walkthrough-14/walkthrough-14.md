# Walkthrough 14: Payment Lifecycle & Mock Gateway Architecture (Clean Architecture & Webhooks)

> **สรุปการเรียนรู้: การออกแบบและพัฒนาระบบชำระเงิน (Payment System) และการยืนยันคำสั่งซื้อ (Order Confirmation)**  
> มุ่งเน้นการใช้ **Interface-Driven Design** สำหรับ **Payment Gateway (Mock Gateway)**, การออก **Payment Intent (PromptPay QR Code / Credit Card URL)**, การทำ **Atomic Status Transition (`pending` ──► `paid`)**, การป้องกัน **Double-Payment & Idempotency**, และการบันทึกใบเสร็จรับเงิน (Payment Receipt)

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 14

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # [ปรับปรุง] Wiring Payment Module (Repo, Gateway, Service, Handler)
├── internal/
│   ├── domain/
│   │   └── payment.go                      # [ใหม่] Payment Entity, DTOs, PaymentGateway & PaymentRepository Interfaces
│   ├── order/
│   │   └── service.go                      # [ปรับปรุง] ปรับเวลา Asynq Timeout เป็น 5 นาที (5*time.Minute)
│   └── payment/                            # [ใหม่] โมดูลระบบชำระเงิน
│       ├── gateway.go                      # Mock Payment Gateway (สร้าง QR Code / Checkout URL จำลอง)
│       ├── handler.go                      # HTTP Handlers (POST /intent, POST /confirm, GET /orders/{id})
│       ├── repository.go                   # Database Transaction (Atomic Lock & Status Update)
│       └── service.go                      # Business Logic & Gateway Verification
├── migrations/
│   ├── 000004_create_payments_table.up.sql   # [ใหม่] สร้างตาราง payments พร้อม Indexes
│   └── 000004_create_payments_table.down.sql # [ใหม่] Rollback ตาราง payments
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-14/
│       └── walkthrough-14.md
├── docs/                                   # [ปรับปรุง] อัปเดต OpenAPI / Swagger Specs สำหรับ Payments
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังสถาปัตยกรรม Complete Payment Lifecycle

```
[ 1. User Checkout ] ──► POST /api/orders/checkout
                           │
                           ▼
               [ Order Created (Status: "pending") ]
               [ Reserve Stock (ตัดสต็อกชั่วคราว) ]
               [ Asynq Timer: รอจ่ายเงิน 5 นาที ]
                           │
                           ▼
[ 2. Create Intent ] ──► POST /api/payments/intent
                           │
                           ▼
               [ Mock Payment Gateway ]
               - สร้าง Transaction Ref (TXN_MOCK_...)
               - ออก PromptPay QR Code หรือ URL บัตรเครดิต
                           │
                           ▼
[ 3. Confirm Pay ]   ──► POST /api/payments/confirm (หรือ Webhook)
                           │
                           ▼ (Atomic Database Transaction)
┌─────────────────────────────────────────────────────────────┐
│ 4. Repository.ConfirmOrderPaymentTx                         │
│    ├── ล็อกแถว Order (SELECT ... FOR UPDATE)                │
│    ├── ตรวจสอบว่ายังไม่ถูกยกเลิก (not cancelled)            │
│    ├── ตรวจสอบยอดเงินตรงกับบิล (amount match)               │
│    ├── เปลี่ยน orders.status = 'paid'                       │
│    └── บันทึกประวัติใบเสร็จลงตาราง `payments`               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
[ 5. Timeout Worker ] ──► เมื่อครบ 5 นาที: ตรวจพบ status == 'paid'
                          ==> 🛡️ ไม่ยกเลิกบิล และไม่คืนสต็อก!
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. Interface-Driven Payment Gateway (Clean Architecture)
```go
type PaymentGateway interface {
    GenerateIntent(ctx context.Context, orderID int, amount float64, method string) (*PaymentIntentResponse, error)
    VerifyTransaction(ctx context.Context, transactionRef string, expectedAmount float64) (bool, error)
}
```
- ด้วยการใช้ Interface ทำให้เราสามารถสลับระหว่าง **`MockGateway`**, **`StripeGateway`**, หรือ **`OmiseGateway`** ได้อย่างง่ายดายโดยไม่ต้องแก้โค้ด Service เลย (Open/Closed Principle)

### 2. Atomic Status Transition & Double-Payment Protection
- ใช้ Pessimistic Locking (`SELECT status FROM orders WHERE id = $1 FOR UPDATE`):
  - ป้องกันการกดยืนยันจ่ายเงินซ้ำ (Idempotency)
  - ป้องกันการจ่ายเงินให้กับบิลที่หมดเวลาและถูกยกเลิกไปแล้ว (`cancelled`)

### 3. Asynq Timeout & Payment State Coordination
- เมื่อลูกค้ายืนยันการจ่ายเงินสำเร็จ (`status = 'paid'`) เมื่อ Worker ของ Asynq ตื่นขึ้นมาตรวจบิลในนาทีที่ 5 จะพบว่าบิลจ่ายเงินเรียบร้อยแล้ว จึงปล่อยให้บิลสมบูรณ์และไม่คืนสต็อกสินค้า

---

## 📊 4. สรุปผลการทดสอบผ่าน Postman & Swagger

| Endpoint | Method | คำอธิบาย | ผลการทดสอบจริง |
| :--- | :---: | :--- | :---: |
| `/api/orders/checkout` | `POST` | สั่งซื้อสินค้า Monitor (14,900 THB) | ✅ บิล Order ID 11 (`status: pending`) |
| `/api/payments/intent` | `POST` | ขอ PromptPay QR Code | ✅ ได้รับ `TXN_MOCK_1787641907_11` และ QR Data |
| `/api/payments/confirm` | `POST` | ยืนยันการชำระเงิน 14,900 THB | ✅ สำเร็จ! Order เปลี่ยนสถานะเป็น `paid` |
| `/api/payments/orders/11` | `GET` | ตรวจสอบใบเสร็จรับเงิน | 🧾 ได้ข้อมูล Payment ID 1 พร้อม Timestamp ถูกต้อง |
