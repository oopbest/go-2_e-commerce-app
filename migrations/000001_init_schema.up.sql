-- สร้างตาราง products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL CHECK (price > 0),
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ใส่ข้อมูลเริ่มต้น (Seed Data โดยระบุ ID ชัดเจน)
INSERT INTO products (id, name, description, price, stock) VALUES
(1, 'Mechanical Keyboard', 'RGB Hot-swappable', 2590.00, 15),
(2, 'Wireless Mouse', 'Ergonomic 2.4GHz', 1290.00, 30),
(3, 'Gaming Headset', '7.1 Surround Sound', 1990.00, 20)
ON CONFLICT (id) DO NOTHING;

-- รีเซ็ตค่า Sequence ของ ID ให้ตรงกับค่าล่าสุด
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));

-- สร้างตาราง users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- สร้างตาราง orders (หัวบิลคำสั่งซื้อ)
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, paid, cancelled
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- สร้างตาราง order_items (รายการสินค้าในคำสั่งซื้อ)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    price NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
