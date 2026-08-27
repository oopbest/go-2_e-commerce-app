-- 1. สร้างตาราง brands
CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    logo_url TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. ใส่ข้อมูลแบรนด์เริ่มต้น (Seed Brands)
INSERT INTO brands (id, name, logo_url, description) VALUES
(1, 'Keychron', 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=100', 'Premium Custom Mechanical Keyboards'),
(2, 'Logitech G', 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=100', 'High-Performance Gaming Gear & Peripherals'),
(3, 'HyperX', 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=100', 'Professional Gaming Headsets & Esports Audio'),
(4, 'Alienware', 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=100', 'Ultra-Premium Gaming Displays & Monitors')
ON CONFLICT (id) DO NOTHING;

SELECT setval('brands_id_seq', (SELECT MAX(id) FROM brands));

-- 3. ขยายโครงสร้างตาราง products
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS brand_id INT REFERENCES brands(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS image_url TEXT,
ADD COLUMN IF NOT EXISTS sku VARCHAR(100),
ADD COLUMN IF NOT EXISTS specs JSONB DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS rating NUMERIC(2, 1) DEFAULT 5.0,
ADD COLUMN IF NOT EXISTS reviews_count INT DEFAULT 0;

-- 4. สร้าง Index สำหรับการค้นหาและกรองด้วย Brand และ Specs
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_products_specs ON products USING gin(specs);

-- 5. อัปเดตข้อมูลสินค้าเดิมให้มีรูปภาพสวยๆ แบรนด์ และสเปกครบถ้วน
UPDATE products 
SET 
    brand_id = 1,
    sku = 'KB-KEY-001',
    image_url = 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80',
    rating = 4.9,
    reviews_count = 142,
    specs = '{"layout": "75% (84 Keys)", "switch": "Gateron G Pro Red", "connectivity": "Bluetooth 5.1 / Type-C", "battery": "4000 mAh", "warranty": "2 Years"}'::jsonb
WHERE id = 1;

UPDATE products 
SET 
    brand_id = 2,
    sku = 'MS-LOGI-002',
    image_url = 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80',
    rating = 4.8,
    reviews_count = 98,
    specs = '{"sensor": "HERO 25K", "dpi": "100 - 25,600 DPI", "connectivity": "LIGHTSPEED Wireless", "weight": "63 grams", "battery": "70 Hours"}'::jsonb
WHERE id = 2;

UPDATE products 
SET 
    brand_id = 3,
    sku = 'HS-HYPER-003',
    image_url = 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=800&q=80',
    rating = 4.7,
    reviews_count = 76,
    specs = '{"driver": "Dynamic 53mm with Neodymium Magnets", "surround": "DTS Headphone:X Spatial Audio", "frame": "Aluminum", "microphone": "Detachable Noise-cancelling"}'::jsonb
WHERE id = 3;

UPDATE products 
SET 
    brand_id = 4,
    sku = 'MN-ALIEN-004',
    image_url = 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80',
    rating = 5.0,
    reviews_count = 54,
    specs = '{"size": "34-inch QD-OLED Curved 1800R", "resolution": "WQHD 3440 x 1440", "refresh_rate": "175Hz", "response_time": "0.1ms GtG", "hdr": "VESA DisplayHDR TrueBlack 400"}'::jsonb
WHERE id = 4;
