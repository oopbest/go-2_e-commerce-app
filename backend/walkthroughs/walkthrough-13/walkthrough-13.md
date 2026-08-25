# Walkthrough 13: Cloud Observability & Real-Time Metrics (Prometheus + Grafana Dashboard)

> **สรุปการเรียนรู้: การสร้างระบบ Real-Time Metrics & Observability ในระดับ Production ด้วย Go**  
> มุ่งเน้นการใช้งาน **`prometheus/client_golang`**, การออกแบบ **Counter**, **Gauge**, และ **Histogram**, การทำ **Path Normalization** ป้องกัน High Cardinality, การดึง **Go Runtime Stats (Goroutines, Memory)**, การวัด **Business Metrics (Orders & Revenue)**, และการแสดงผลบน **Grafana Live Dashboard**

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 13

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # [ปรับปรุง] ลงทะเบียน /metrics และครอบ MetricsMiddleware
├── deploy/
│   └── prometheus/
│       └── prometheus.yml                  # [ใหม่] Scrape Config ดึง Metrics จาก api:8080 ทุก 5s
├── internal/
│   ├── metrics/
│   │   └── metrics.go                      # [ใหม่] ประกาศ Prometheus Metric Collectors (RPS, Latency, Orders, Revenue)
│   ├── middleware/
│   │   └── metrics.go                      # [ใหม่] MetricsMiddleware ดักจับ Request, วัดเวลา & Normalize Path
│   └── order/
│       └── service.go                      # [ปรับปรุง] บันทึก OrdersTotal และ RevenueTotal เมื่อ Checkout สำเร็จ
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-13/
│       └── walkthrough-13.md
├── docker-compose.yml                      # [ปรับปรุง] เพิ่ม service prometheus (พอร์ต 9090) และ grafana (พอร์ต 3000)
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังสถาปัตยกรรม Observability Pipeline

```
[ Client Traffic (Web / Mobile / Swagger UI) ]
                     │
                     ▼ (HTTP Requests)
┌─────────────────────────────────────────────────────────────┐
│ 1. Go REST API (cmd/api)                                    │
│    ├── Middleware Pipeline: Recovery -> Metrics -> Logger   │
│    │   - นับ In-Flight Requests (Gauge)                     │
│    │   - บันทึก Latency Duration (Histogram)                │
│    │   - บันทึก Requests Count แยกตาม Method, Path, Status  │
│    │   - ทำ Path Normalization (/api/products/{id})         │
│    ├── Business Service: นับยอดขาย (Revenue) และ Orders รวม │
│    └── Endpoint: GET /metrics (promhttp.Handler)            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼ (Scrape every 5s)
┌─────────────────────────────────────────────────────────────┐
│ 2. Prometheus Server (port: 9090)                           │
│    - Time-Series Database จัดเก็บและประมวลผลข้อมูลสถิติ     │
│    - รองรับการ Query ด้วยภาษา PromQL                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼ (PromQL Queries)
┌─────────────────────────────────────────────────────────────┐
│ 3. Grafana Live Dashboard (port: 3000)                      │
│    ├── 📊 Requests Per Second (RPS) by Route & Status       │
│    ├── 🛒 Total Orders Placed Counter                       │
│    ├── 💰 Total Revenue Generated (THB)                     │
│    └── 🧵 Live Active Goroutines Count                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. การเลือกใช้ Prometheus Data Types ที่ถูกต้อง
- **Counter (`ecommerce_http_requests_total`, `ecommerce_orders_total`, `ecommerce_revenue_thb_total`)**: ตัวเลขที่เพิ่มขึ้นเรื่อยๆ ห้ามลดลง
- **Gauge (`ecommerce_http_requests_in_flight`, `go_goroutines`)**: ตัวเลขที่ขึ้น-ลงได้ตามสภาวะจริง
- **Histogram (`ecommerce_http_request_duration_seconds`)**: บันทึกการกระจายตัวของ Latency เป็น Buckets เพื่อนำมาคำนวณ **P95 / P99 Percentiles**

### 2. การป้องกันปัญหา High Cardinality ด้วย Path Normalization
```go
var idRegex = regexp.MustCompile(`/\d+`)

func normalizePath(path string) string {
    return idRegex.ReplaceAllString(path, "/{id}")
}
```
- แปลง `/api/products/1` และ `/api/products/2` ให้เป็น `/api/products/{id}` เสมอ เพื่อไม่ให้สร้าง Label นับล้านตัวจนล้น Memory ของ Prometheus

### 3. Business & Infrastructure Metrics Co-location
- นอกจากการวัดระบบพื้นฐาน (RPS, Latency, Goroutines) ยังสามารถบันทึก Metrics ทางธุรกิจ เช่น ยอดขายรวม (Revenue in THB) และจำนวนคำสั่งซื้อ ทำให้ Dashboard มีคุณค่าทั้งต่อฝ่าย Tech และฝ่าย Business!

---

## 📊 4. สรุปผลการทดสอบบน Grafana Live Dashboard

| Panel Name | คำสั่ง PromQL | ผลการทดสอบจริง |
| :--- | :--- | :---: |
| **Requests Per Second (RPS)** | `sum(rate(ecommerce_http_requests_total[1m])) by (path, status)` | 📈 แสดงเส้นแยก `/api/auth/login`, `/api/orders/checkout`, `/api/products/{id}` ตามจริง |
| **Total Orders Placed** | `ecommerce_orders_total` | 🛒 ตัวเลขอัปเดตเพิ่มขึ้นทันทีเมื่อมี Order เข้า |
| **Total Revenue (บาท)** | `ecommerce_revenue_thb_total` | 💰 กราฟขั้นบันไดเพิ่มขึ้นตามยอดเงินจริง (> 20,000 THB) |
| **Go Goroutines Count** | `go_goroutines` | 🧵 อยู่ในช่วง 10-13 Goroutines ไม่มี Memory/Goroutine Leak |
