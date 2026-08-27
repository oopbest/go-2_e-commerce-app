DELETE FROM products WHERE id >= 5;
DELETE FROM brands WHERE id >= 5;
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));
SELECT setval('brands_id_seq', (SELECT MAX(id) FROM brands));
