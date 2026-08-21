# Walkthrough 6: ยกระดับสู่ Production-Ready (Graceful Shutdown, Structured Logging ด้วย `log/slog`, Panic Recovery, และ Config Management)

> **สรุปการเรียนรู้: เสาหลักความเสถียรและความปลอดภัยระดับ Production ใน Go Backend**  
> มุ่งเน้นการจัดการ Configuration ตามหลัก 12-Factor App, การบันทึก Log แบบ **Structured Logging (`log/slog`)**, การสร้าง Middleware ดักจับ Panic และคำนวณ Latency, และการปิดระบบอย่างนุ่มนวล (**Graceful Shutdown**)

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 6

```text
ecommerce-app/
├── cmd/
│   └── api/
│       └── main.go                         # [ปรับปรุง] Structured Logger, Server Timeouts, และ Graceful Shutdown
├── internal/
│   ├── config/                             # [ใหม่] โหลดการตั้งค่าจาก Environment Variables
│   │   └── config.go
│   ├── database/
│   │   └── postgres.go
│   ├── domain/
│   │   ├── order.go
│   │   ├── product.go
│   │   └── user.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go                       # [ใหม่] Request Logger ดักจับ Status Code & Latency ด้วย slog
│   │   └── recovery.go                     # [ใหม่] Panic Recovery ป้องกัน Server พัง
│   ├── order/
│   ├── product/
│   └── user/
├── pkg/
│   └── security/
├── migrations/
│   └── init.sql
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── walkthrough-2/
│   ├── walkthrough-3/
│   ├── walkthrough-4/
│   ├── walkthrough-5/
│   └── walkthrough-6/
│       └── walkthrough-6.md
├── .env                                    # [ใหม่] ไฟล์เก็บค่า Configuration และ Secret
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังวงจรการทำงานของ Server Lifecycle & Middleware Pipeline

```
[ 1. เริ่มต้นระบบ (Startup) ]
.env ──> config.Load() ──> slog.New() ──> database.NewPostgresDB() ──> DI Modules ──> go server.ListenAndServe()

────────────────────────────────────────────────────────────────────────────────────────────────────────────

[ 2. เมื่อมี HTTP Request เข้ามา (Middleware Chain) ]
Client Request 
     │
     ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. middleware.Recovery                                      │
│    - ครอบการทำงานทั้งหมดด้วย defer recover()                 │
│    - ถ้าเกิด Panic -> ดักจับ บันทึก Stack Trace และส่ง 500   │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. middleware.RequestLogger                                 │
│    - จับเวลาเริ่มต้น: start := time.Now()                    │
│    - ครอบ ResponseWriter เพื่อดักจับ HTTP Status Code        │
│    - เมื่อเสร็จสิ้น -> บันทึก Structured Log ด้วย slog       │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. ServeMux & Auth Middleware & Feature Handlers            │
└─────────────────────────────────────────────────────────────┘

────────────────────────────────────────────────────────────────────────────────────────────────────────────

[ 3. เมื่อได้รับสัญญาณปิดระบบ (Graceful Shutdown) ]
Signal (Ctrl+C / SIGTERM) 
     │
     ▼
quit <- channel ──> server.Shutdown(10s Timeout) ──> รอ Request เก่าเสร็จ ──> db.Close() ──> Exit Cleanly
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. Configuration Management (12-Factor App)
```go
func Load() *Config {
    _ = godotenv.Load() // โหลดจาก .env ในเครื่อง Dev (ถ้ามี)
    return &Config{
        Port:      getEnv("SERVER_PORT", "8080"),
        JWTSecret: getEnv("JWT_SECRET", "default-fallback"),
        ...
    }
}
```
- แยกการตั้งค่า (Config) และรหัสผ่านลับ (Secrets) ออกจาก Source Code ทำให้โค้ดสามารถรันได้ทุก Environment (Dev, Staging, Production) โดยไม่ต้องแก้โค้ด

### 2. Structured Logging ด้วย `log/slog` (มาตรฐาน Go 1.21+)
- **Development**: ใช้ `slog.NewTextHandler` เพื่อให้อ่านง่ายบน Terminal
- **Production**: ใช้ `slog.NewJSONHandler` เพื่อให้ระบบ Cloud Monitoring (เช่น Loki, Datadog) สามารถ Query และ Filter ข้อมูลตาม Key ต่างๆ ได้ทันที

### 3. เทคนิค ResponseWriter Wrapping เพื่อดักจับ Status Code
```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```
- ใน Go Standard Library `http.ResponseWriter` ดั้งเดิมไม่มี Getter สำหรับดู Status Code การใช้ **Struct Embedding** ดัก Override `WriteHeader` ทำให้เราสามารถคำนวณและบันทึก Status Code ลงใน Logger Middleware ได้อย่างสมบูรณ์แบบ

### 4. Panic Recovery Middleware (`internal/middleware/recovery.go`)
- ดักจับ Runtime Panic (เช่น Null Pointer Exception) ด้วย `recover()` และดึง Stack Trace ผ่าน `debug.Stack()` ช่วยให้ระบบ **ไม่พังทั้ง Process** เมื่อมี Error หลุดรอด

### 5. Server Timeouts (ป้องกัน Slowloris DoS Attacks)
```go
server := &http.Server{
    Addr:         ":" + cfg.Port,
    Handler:      handlerWithMiddlewares,
    ReadTimeout:  10 * time.Second, // ตัดการเชื่อมต่อถ้า Client ส่ง Header/Body ช้าเกินไป
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### 6. Graceful Shutdown ด้วย Channels & OS Signals
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
<-quit // บล็อกรอสัญญาณ Ctrl+C หรือคำสั่งปิดจาก Docker/K8s

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = server.Shutdown(shutdownCtx) // รอให้ Request ที่กำลังทำงานอยู่เสร็จสิ้น
```

---

## 🚀 4. ก้าวต่อไป: สรุปภาพรวมและต่อยอดสู่ระดับ Senior

ยินดีด้วยครับ! ตอนนี้โปรเจกต์ **Go E-Commerce Backend** ของคุณมีคุณสมบัติครบถ้วนตามมาตรฐานสากล:
- ✅ **Clean Architecture & Standard Go Layout**
- ✅ **Database Transactions & Concurrency Stock Locking**
- ✅ **User Authentication, bcrypt & JWT Security**
- ✅ **Role-Based Access Control (RBAC)**
- ✅ **PostgreSQL Database & Connection Pooling**
- ✅ **Structured Logging & Panic Recovery**
- ✅ **Graceful Shutdown & Environment Management**
