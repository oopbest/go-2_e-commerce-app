# Walkthrough 12: Enterprise Database Migrations & Schema Versioning (golang-migrate + embed.FS)

> **สรุปการเรียนรู้: การจัดการโครงสร้างฐานข้อมูลแบบ Version Control ใน Go**  
> มุ่งเน้นการใช้งาน **`golang-migrate/migrate/v4`**, การฝังไฟล์ SQL ลงในไบนารีด้วย **`embed.FS`**, การเขียนสคริปต์ **UP (อัปเกรด)** และ **DOWN (ย้อนกลับ/Rollback)**, การทำ **Auto-Migration** อัตโนมัติเมื่อแอปพลิเคชันสตาร์ท, และการจัดการ Zero-Downtime Schema Evolution

---

## 🏗️ 1. โครงสร้างโปรเจกต์ที่เพิ่มขึ้นใน Phase 12

```text
ecommerce-app/
├── cmd/
│   ├── api/
│   │   └── main.go                         # [ปรับปรุง] เรียก database.RunMigrations(db) ตอนสตาร์ท
│   └── worker/
│       └── main.go
├── internal/
│   └── database/
│       ├── migration.go                    # [ใหม่] Auto-Migration Engine (iofs driver + postgres driver)
│       ├── postgres.go
│       └── redis.go
├── migrations/                             # [ใหม่] Versioned SQL Files พร้อม embed.FS
│   ├── 000001_init_schema.up.sql           # สร้างตาราง products, users, orders, order_items & Seed Data
│   ├── 000001_init_schema.down.sql         # Rollback ลบตารางย้อนกลับตามลำดับ FK
│   ├── 000002_add_categories_table.up.sql   # สร้างตาราง categories และ ALTER TABLE products ADD category_id
│   ├── 000002_add_categories_table.down.sql # Rollback ตาราง categories
│   └── migrations.go                       # [ใหม่] ฝังไฟล์ SQL เข้าไบนารีด้วย //go:embed *.sql
├── walkthroughs/
│   ├── walkthrough-1/
│   ├── ...
│   └── walkthrough-12/
│       └── walkthrough-12.md
├── docker-compose.yml                      # [ปรับปรุง] ใช้พอร์ต 15432:5432 และลบ volume mount เก่าออก
├── go.mod
└── go.sum
```

---

## 🔄 2. แผนผังสถาปัตยกรรม Database Versioning & Auto-Migration

```
[ Go Binary Compilation (go build / Dockerfile) ]
  migrations/*.sql ──( //go:embed *.sql )──> ฝังลงใน Binary โดยตรง! 📦
                                                    │
                                                    ▼
[ Server Startup (cmd/api/main.go) ]
  1. เชื่อมต่อ PostgreSQL Pool (db)
  2. เรียก database.RunMigrations(db)
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. ตรวจสอบตาราง schema_migrations ใน PostgreSQL              │
│    - อ่าน Version ล่าสุดในฐานข้อมูล                            │
│    - เปรียบเทียบกับ Version สูงสุดที่มีใน embed.FS            │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼ (Apply UP Migrations)
┌─────────────────────────────────────────────────────────────┐
│ 4. ทำการรันสคริปต์ UP ทีละเวอร์ชันตามลำดับ                     │
│    ├── 000001: สร้างโครงสร้างตารางหลัก & Seed Data          │
│    ├── 000002: สร้างตาราง categories & เพิ่ม category_id     │
│    └── อัปเดต schema_migrations.version = 2                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧠 3. หัวใจและ Concept สำคัญที่ได้เรียนรู้ใน Phase นี้

### 1. การฝังไฟล์ SQL เข้าสู่ Binary ด้วย `embed.FS` (`migrations/migrations.go`)
```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```
- ทำให้ไฟล์ Migration เดินทางไปพร้อมกับ Go Binary เสมอ ไม่ต้องคัดลอกไฟล์ SQL แยกไปบน Production Server

### 2. Auto-Migration Engine (`internal/database/migration.go`)
```go
driver, _ := iofs.New(migrations.FS, ".")
dbDriver, _ := postgres.WithInstance(db, &postgres.Config{})
m, _ := migrate.NewWithInstance("iofs", driver, "postgres", dbDriver)

if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
    return err
}
```
- รัน UP อัตโนมัติทุกครั้งที่ Deploy เวอร์ชันใหม่ และไม่ Error หาก Schema เป็นเวอร์ชันล่าสุดอยู่แล้ว (`migrate.ErrNoChange`)

### 3. Zero-Downtime Schema Evolution & Backward Compatibility
- เมื่อใช้คำสั่ง `ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id INT ...`:
  - สินค้าเดิมที่มีอยู่จะได้รับค่า `category_id = NULL` โดยอัตโนมัติ เพื่อไม่ให้กระทบกับระบบเดิมที่กำลังรันอยู่
  - เราสามารถเขียนคำสั่ง **Data Migration** เพื่ออัปเดตหมวดหมู่ให้สินค้าเดิมได้โดยที่ระบบไม่ต้องหยุดทำงาน

### 4. การจัดการ Port Exclusion บน Windows Hyper-V / WSL2
- ตรวจสอบช่วงพอร์ตที่ถูกจองด้วย `netsh interface ipv4 show excludedportrange protocol=tcp`
- แก้ปัญหาการแย่งพอร์ต `5432` บน Windows ด้วยการเปลี่ยนพอร์ตภายนอกของ Postgres ใน Docker เป็น **`15432:5432`**

---

## 📊 4. ตารางบันทึกสถานะ Database Schema

| Version | Migration Name | สิ่งที่เปลี่ยนแปลง | สถานะ |
| :---: | :--- | :--- | :---: |
| **1** | `000001_init_schema` | สร้างตาราง `products`, `users`, `orders`, `order_items` | ✅ Applied |
| **2** | `000002_add_categories_table` | สร้างตาราง `categories` และเพิ่ม `products.category_id` | ✅ Applied |
