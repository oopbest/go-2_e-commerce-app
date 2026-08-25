package product

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository Constructor สำหรับสร้าง PostgreSQL Repository
func NewPostgresRepository(db *sql.DB) domain.ProductRepository {
	return &postgresRepository{
		db: db,
	}
}

// FindAll ดึงรายการสินค้าทั้งหมดจาก Database
func (r *postgresRepository) FindAll() ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, price, stock, created_at
		FROM products
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // ต้องปิด rows เสมอเมื่ออ่านเสร็จเพื่อคืน Connection เข้า Pool

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		// Scan ตัวแปรตามลำดับคอลัมน์ที่ SELECT ออกมา
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindByID ค้นหาสินค้าตาม ID
func (r *postgresRepository) FindByID(id int) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, price, stock, created_at
		FROM products
		WHERE id = $1
	`
	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return &p, nil
}

// Create เพิ่มสินค้าใหม่ลง Database
func (r *postgresRepository) Create(input domain.CreateProductInput) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ใช้ RETURNING เพื่อเอา ID และ CreatedAt ที่ Database เจนให้ออกมาทันที
	query := `
		INSERT INTO products (name, description, price, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	var p domain.Product
	p.Name = input.Name
	p.Description = input.Description
	p.Price = input.Price
	p.Stock = input.Stock

	err := r.db.QueryRowContext(ctx, query, input.Name, input.Description, input.Price, input.Stock).Scan(
		&p.ID, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// Update แก้ไขสินค้า
func (r *postgresRepository) Update(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4
		WHERE id = $5
		RETURNING created_at
	`
	var p domain.Product
	p.ID = id
	p.Name = input.Name
	p.Description = input.Description
	p.Price = input.Price
	p.Stock = input.Stock

	err := r.db.QueryRowContext(ctx, query, input.Name, input.Description, input.Price, input.Stock, id).Scan(
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return &p, nil
}

// Delete ลบสินค้า
func (r *postgresRepository) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}
