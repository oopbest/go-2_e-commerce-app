-- 1. สร้างตาราง promotions (คูปองส่วนลด)
CREATE TABLE IF NOT EXISTS promotions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('fixed', 'percentage')),
    discount_value NUMERIC(10, 2) NOT NULL CHECK (discount_value > 0),
    min_spend NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    max_discount NUMERIC(10, 2),
    total_quota INT NOT NULL DEFAULT 100 CHECK (total_quota > 0),
    used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    starts_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. สร้างตาราง flash_sales (แคมเปญแฟลชเซลล์)
CREATE TABLE IF NOT EXISTS flash_sales (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. สร้างตาราง flash_sale_items (สินค้าที่เข้าร่วมแฟลชเซลล์)
CREATE TABLE IF NOT EXISTS flash_sale_items (
    id SERIAL PRIMARY KEY,
    flash_sale_id INT NOT NULL REFERENCES flash_sales(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    flash_price NUMERIC(10, 2) NOT NULL CHECK (flash_price > 0),
    flash_stock INT NOT NULL CHECK (flash_stock > 0),
    sold_count INT NOT NULL DEFAULT 0 CHECK (sold_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_flash_sale_product UNIQUE(flash_sale_id, product_id)
);

-- 4. ขยายตาราง orders ให้บันทึกการใช้คูปองและยอดส่วนลด
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS coupon_id INT REFERENCES promotions(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(10, 2) NOT NULL DEFAULT 0.00;

-- 5. สร้าง Index เพื่อประสิทธิภาพในการค้นหาโค้ดและแคมเปญ
CREATE INDEX IF NOT EXISTS idx_promotions_code ON promotions(code);
CREATE INDEX IF NOT EXISTS idx_promotions_active ON promotions(is_active, expires_at);
CREATE INDEX IF NOT EXISTS idx_flash_sales_active ON flash_sales(is_active, starts_at, expires_at);
CREATE INDEX IF NOT EXISTS idx_flash_sale_items_product ON flash_sale_items(product_id);

-- 6. ใส่ข้อมูลเริ่มต้น (Seed Promotions & Flash Sale)
-- คูปอง 1: ลด 100 บาท เมื่อซื้อครบ 1,000 บาท
-- คูปอง 2: ลด 15% (สูงสุด 300 บาท) เมื่อซื้อครบ 1,500 บาท
INSERT INTO promotions (id, code, title, description, discount_type, discount_value, min_spend, max_discount, total_quota, starts_at, expires_at) VALUES
(1, 'WELCOME100', 'ยินดีต้อนรับ สมาชิกใหม่ลด 100 บาท', 'ลดทันที 100 บาท เมื่อสั่งซื้อสินค้าครบ 1,000 บาทขึ้นไป', 'fixed', 100.00, 1000.00, NULL, 500, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days'),
(2, 'NEXTGEN15', 'NextGen Mega Sale ลด 15%', 'รับส่วนลด 15% สูงสุด 300 บาท เมื่อสั่งซื้อครบ 1,500 บาท', 'percentage', 15.00, 1500.00, 300.00, 200, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '14 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('promotions_id_seq', (SELECT MAX(id) FROM promotions));

-- แคมเปญ Flash Sale: มีผล 48 ชั่วโมงนับจากปัจจุบัน
INSERT INTO flash_sales (id, title, description, starts_at, expires_at, is_active) VALUES
(1, 'Midnight Flash Deals ⚡', 'ลดราคาสินค้าเกมมิ่งเกียร์ยอดฮิตแบบจำกัดเวลาและจำนวนชิ้น!', CURRENT_TIMESTAMP - INTERVAL '1 hour', CURRENT_TIMESTAMP + INTERVAL '48 hours', TRUE)
ON CONFLICT (id) DO NOTHING;

SELECT setval('flash_sales_id_seq', (SELECT MAX(id) FROM flash_sales));

-- นำ Mechanical Keyboard (ID: 1) และ Gaming Headset (ID: 3) เข้าร่วม Flash Sale
-- สินค้า ID 1: ปกติ 2,590 -> Flash Price 1,990 (โควตา 5 ชิ้น)
-- สินค้า ID 3: ปกติ 1,990 -> Flash Price 1,490 (โควตา 5 ชิ้น)
INSERT INTO flash_sale_items (flash_sale_id, product_id, flash_price, flash_stock, sold_count) VALUES
(1, 1, 1990.00, 5, 0),
(1, 3, 1490.00, 5, 0)
ON CONFLICT (flash_sale_id, product_id) DO NOTHING;
