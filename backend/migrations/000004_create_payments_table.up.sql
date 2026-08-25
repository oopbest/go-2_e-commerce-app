-- สร้างตาราง payments สำหรับบันทึกประวัติการชำระเงินและใบเสร็จ
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(50) NOT NULL, -- credit_card, promptpay, wallet, bank_transfer
    transaction_ref VARCHAR(100) UNIQUE NOT NULL, -- รหัสอ้างอิงจาก Payment Gateway (เช่น TXN_MOCK_xxxx)
    status VARCHAR(50) NOT NULL DEFAULT 'completed', -- completed, failed, refunded
    paid_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- สร้าง Index เพื่อให้ค้นหาประวัติการชำระเงินตาม order_id และ transaction_ref ได้เร็วขึ้น
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_ref ON payments(transaction_ref);
