-- สร้างตาราง products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL CHECK (price > 0),
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ใส่ข้อมูลเริ่มต้น (Seed Data)
INSERT INTO products (name, description, price, stock) VALUES
('Mechanical Keyboard', 'RGB Hot-swappable', 2590.00, 15),
('Wireless Mouse', 'Ergonomic 2.4GHz', 1290.00, 30),
('Gaming Headset', '7.1 Surround Sound', 1990.00, 20);
