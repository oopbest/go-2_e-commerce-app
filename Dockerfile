# ==========================================
# Stage 1: Build Stage (คอมไพล์ Go Binary)
# ==========================================
FROM golang:alpine AS builder

WORKDIR /app

# ติดตั้ง git และ ca-certificates เผื่อกรณีใช้ดึง dependencies
RUN apk add --no-cache git ca-certificates tzdata

# คัดลอกเฉพาะ go.mod และ go.sum ก่อน เพื่อให้ Docker แคช Dependency Layer
COPY go.mod go.sum ./
RUN go mod download

# คัดลอก Source Code ทั้งหมด
COPY . .

# คอมไพล์เป็น Static Binary
# - CGO_ENABLED=0: ตัดการผูกมัดกับ C library ทั้งหมด
# - GOOS=linux: คอมไพล์สำหรับระบบ Linux
# - -ldflags="-w -s": ลบ Debug Information และ DWARF Symbol Tables ทำให้ขนาด Binary เล็กลง 30-40%!
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/api

# ==========================================
# Stage 2: Final Production Stage (ขนาดจิ๋ว < 20MB)
# ==========================================
FROM alpine:3.21

WORKDIR /app

# ติดตั้ง Root Certificates (สำหรับยิง HTTPS) และ Timezone Data
RUN apk --no-cache add ca-certificates tzdata

# 🔒 Security Best Practice: สร้าง Non-Root User เพื่อความปลอดภัย (ไม่รันด้วยสิทธิ์ root บน Production)
RUN adduser -D -g '' appuser

# คัดลอกเฉพาะไฟล์ Binary ที่คอมไพล์เสร็จแล้วมาจาก Stage 1
COPY --from=builder /app/bin/server /app/server

# ปรับสิทธิ์การเข้าถึงให้เป็นของ appuser
USER appuser

# เปิดพอร์ตที่แอปทำงาน
EXPOSE 8080

# คำสั่งเริ่มต้นการทำงาน
ENTRYPOINT ["/app/server"]
