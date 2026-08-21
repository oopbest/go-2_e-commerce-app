# Walkthrough 9: Multi-Stage Containerization & Production Docker Compose

> **สรุปการเรียนรู้: การแพ็กเกจ Go Backend สู่ Production Cloud Container ระดับ Enterprise**  
> มุ่งเน้นการสร้าง **Multi-Stage Dockerfile**, การคอมไพล์ **Static Binary (`CGO_ENABLED=0`, `-ldflags="-w -s"`)**, การรักษาความปลอดภัยด้วย **Non-Root User (`appuser`)**, การกำหนด **Health Checks**, และการรัน Full Stack All-in-One (`api`, `postgres`, `redis`) ด้วย Docker Compose

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 9

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
├── pkg/
├── migrations/
│   └── init.sql
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-9/
│       └── walkthrough-9.md
├── .dockerignore                            # [ใหม่] ป้องกันไม่ให้ส่งไฟล์ไม่จำเป็นเข้า Docker Build
├── Dockerfile                              # [ใหม่] Multi-Stage Build (ขนาด Content Size เพียง 8 MB!)
├── docker-compose.yml                      # [ปรับปรุง] Full Stack Service (API + PostgreSQL + Redis พร้อม Healthcheck)
├── .env
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังสถาปัตยกรรม Multi-Stage Build & Docker Network

```
[ Stage 1: Builder (golang:alpine) ]
  COPY go.mod go.sum ──> RUN go mod download ──> COPY Source Code
                                                        │
                                                        ▼
                        RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server
                                                        │
                                  ┌─────────────────────┘ (ส่งต่อเฉพาะไฟล์ Binary ตัวเดียว)
                                  │
                                  ▼
[ Stage 2: Production Runner (alpine:3.21 - ขนาดจิ๋ว < 20MB) ]
  ├── 🔒 RUN adduser -D -g '' appuser (Non-Root User)
  ├── 📦 COPY --from=builder /app/bin/server /app/server
  ├── 👤 USER appuser
  └── 🚀 ENTRYPOINT ["/app/server"]

──────────────────────────────────────────────────────────────────────────────────────────

[ Full Stack Orchestration (docker compose up -d) ]
                 ┌────────────────────────────┐
                 │       ecommerce_api        │ (Port: 8080)
                 └──────────────┬─────────────┘
                                │ (รอจนกว่า DB และ Cache จะ Healthy)
                ┌───────────────┴───────────────┐
                ▼                               ▼
  ┌───────────────────────────┐   ┌───────────────────────────┐
  │    ecommerce_postgres     │   │      ecommerce_redis      │
  │ (Port: 5432 / pg_isready) │   │ (Port: 6379 / ping check) │
  └───────────────────────────┘   └───────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. Multi-Stage Build Architecture
- **ลดขนาด Image จาก 1,000 MB $\to$ เหลือเพียง 8 MB (Content Size)** เพราะใน Stage สุดท้ายเราใช้ `alpine` เปล่าๆ และคัดลอกเฉพาะไฟล์ Binary ที่คอมไพล์เสร็จแล้วมาใส่เท่านั้น

### 2. Flags การคอมไพล์ระดับมืออาชีพ (`-ldflags="-w -s"`)
- `CGO_ENABLED=0`: บังคับให้ Go สร้าง Pure Static Binary ที่ไม่มีการพึ่งพา Dynamic C Libraries ทำให้รันได้ทุก Linux Distro
- `-ldflags="-w -s"`:
  - `-w`: ลบ DWARF Debugging Information ออก
  - `-s`: ลบ Symbol Table ออก
  - ช่วยลดขนาดไฟล์ Binary ลงถึง 30 - 40%

### 3. Container Security Best Practice (Non-Root User)
```dockerfile
RUN adduser -D -g '' appuser
USER appuser
```
- ในระบบ Production ห้ามรันแอปพลิเคชันด้วยสิทธิ์ `root` เด็ดขาด เพื่อป้องกันความเสียหายหากคอนเทนเนอร์ถูกเจาะระบบ

### 4. Docker Compose Health Checks & Service Dependency
```yaml
depends_on:
  postgres:
    condition: service_healthy
  redis:
    condition: service_healthy
```
- API จะรอจนกว่า PostgreSQL (`pg_isready`) และ Redis (`redis-cli ping`) พร้อมรับ Connection จริงๆ จึงจะเริ่มบูตระบบ ช่วยแก้ปัญหา API Crash ตอนสตาร์ทพร้อมกัน

### 5. Production JSON Logging อัตโนมัติ
- เมื่อรันผ่าน Docker Compose ด้วย `APP_ENV=production` ตัว Logger (`log/slog`) จะสลับโหมดเป็น **JSON Format** โดยอัตโนมัติ พร้อมส่งต่อไปยัง Cloud Log Aggregators (Datadog, Loki, AWS CloudWatch) ได้ทันที

---

## 📊 4. สรุปสถานะการรัน Full Stack

| Service Container | Image | Ports | Healthcheck Status | Role |
| :--- | :--- | :---: | :---: | :--- |
| `ecommerce_api` | `ecommerce-api:latest` (~8MB) | `8080:8080` | `Up` | Go Clean Architecture REST API |
| `ecommerce_postgres` | `postgres:16-alpine` | `5432:5432` | `Up (healthy)` | Relational DB + Transactions |
| `ecommerce_redis` | `redis:7-alpine` | `6379:6379` | `Up (healthy)` | In-Memory Cache (< 1ms Latency) |

---

## 🏆 บทสรุป: สู่การเป็น Go Backend Engineer ตัวจริง

ยินดีด้วยอย่างยิ่งครับ! ตั้งแต่วันแรกที่คุณเริ่มสร้าง CRUD ง่ายๆ จนถึงวันนี้ คุณได้สร้าง **ระบบ E-Commerce Backend ระดับ Production-Ready** ที่สมบูรณ์แบบในทุกมิติ:

* ✅ **Clean Architecture & SOLID Design**
* ✅ **PostgreSQL, Transactions & Stock Concurrency Locking**
* ✅ **Bcrypt Hashing, JWT Security & RBAC Middleware**
* ✅ **12-Factor App Configuration & Structured Logging (`log/slog`)**
* ✅ **Graceful Shutdown & Resilience**
* ✅ **Redis In-Memory Caching with Decorator Pattern**
* ✅ **Automated Testing Suite (`testify`, `httptest`, Code Coverage)**
* ✅ **Ultra-Lean Multi-Stage Docker Container (< 20MB)**
