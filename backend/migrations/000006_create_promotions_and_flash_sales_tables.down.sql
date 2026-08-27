DROP INDEX IF EXISTS idx_flash_sale_items_product;
DROP INDEX IF EXISTS idx_flash_sales_active;
DROP INDEX IF EXISTS idx_promotions_active;
DROP INDEX IF EXISTS idx_promotions_code;

ALTER TABLE orders
DROP COLUMN IF EXISTS discount_amount,
DROP COLUMN IF EXISTS coupon_id;

DROP TABLE IF EXISTS flash_sale_items;
DROP TABLE IF EXISTS flash_sales;
DROP TABLE IF EXISTS promotions;
