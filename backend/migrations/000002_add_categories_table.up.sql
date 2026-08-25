-- 1. สร้างตาราง categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. ใส่ข้อมูลหมวดหมู่เริ่มต้น
INSERT INTO categories (name, description) VALUES
('Gaming Gear', 'Accessories for gamers'),
('Office & Productivity', 'Ergonomic products for work')
ON CONFLICT DO NOTHING;

-- 3. เพิ่มคอลัมน์ category_id ผูก Foreign Key เข้ากับตาราง products เดิม
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id INT REFERENCES categories(id) ON DELETE SET NULL;
