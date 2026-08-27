package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
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

// FindAll ดึงรายการสินค้าทั้งหมดจาก Database พร้อม JOIN หมวดหมู่และแบรนด์
func (r *postgresRepository) FindAll() ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.stock,
			p.category_id, COALESCE(c.name, ''),
			p.brand_id, COALESCE(b.name, ''),
			COALESCE(p.image_url, ''), COALESCE(p.sku, ''),
			COALESCE(p.specs, '{}'::jsonb),
			p.rating, p.reviews_count, p.created_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		ORDER BY p.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var specsBytes []byte

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.CategoryID, &p.CategoryName,
			&p.BrandID, &p.BrandName,
			&p.ImageURL, &p.SKU,
			&specsBytes,
			&p.Rating, &p.ReviewsCount, &p.CreatedAt,
		); err != nil {
			return nil, err
		}

		if len(specsBytes) > 0 {
			_ = json.Unmarshal(specsBytes, &p.Specs)
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindWithFilter ค้นหา กรอง และเรียงลำดับสินค้าตามเงื่อนไขแบบ Dynamic Parameterized Query
func (r *postgresRepository) FindWithFilter(ctx context.Context, filter domain.ProductFilter) (*domain.ProductListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	offset := (filter.Page - 1) * filter.Limit

	// 1. สร้าง Dynamic WHERE Clauses และ Parameter Slice
	var conditions []string
	var args []any
	argIdx := 1

	// ค้นหา Keyword จากชื่อหรือคำอธิบาย (Case-Insensitive)
	if strings.TrimSpace(filter.Search) != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.TrimSpace(filter.Search)+"%")
		argIdx++
	}

	// กรองตามหมวดหมู่
	if filter.CategoryID != nil && *filter.CategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	// กรองตามแบรนด์
	if filter.BrandID != nil && *filter.BrandID > 0 {
		conditions = append(conditions, fmt.Sprintf("p.brand_id = $%d", argIdx))
		args = append(args, *filter.BrandID)
		argIdx++
	}

	// กรองราคาต่ำสุด
	if filter.MinPrice != nil && *filter.MinPrice >= 0 {
		conditions = append(conditions, fmt.Sprintf("p.price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}

	// กรองราคาสูงสุด
	if filter.MaxPrice != nil && *filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("p.price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}

	// กรองเฉพาะที่มีสต็อกพร้อมส่ง
	if filter.InStockOnly {
		conditions = append(conditions, "p.stock > 0")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 2. Query นับจำนวนรวมทั้งหมด (Total Count)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(p.id) 
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		%s
	`, whereClause)

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// 3. Dynamic ORDER BY
	var orderByClause string
	switch filter.SortBy {
	case "price_asc":
		orderByClause = "ORDER BY p.price ASC, p.id ASC"
	case "price_desc":
		orderByClause = "ORDER BY p.price DESC, p.id ASC"
	case "rating":
		orderByClause = "ORDER BY p.rating DESC, p.reviews_count DESC, p.id ASC"
	case "newest":
		orderByClause = "ORDER BY p.created_at DESC, p.id ASC"
	default:
		orderByClause = "ORDER BY p.id ASC"
	}

	// 4. Query ดึงข้อมูลสินค้าพร้อม Pagination
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.name, p.description, p.price, p.stock,
			p.category_id, COALESCE(c.name, ''),
			p.brand_id, COALESCE(b.name, ''),
			COALESCE(p.image_url, ''), COALESCE(p.sku, ''),
			COALESCE(p.specs, '{}'::jsonb),
			p.rating, p.reviews_count, p.created_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIdx, argIdx+1)

	queryArgs := append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products with filter: %w", err)
	}
	defer rows.Close()

	products := []domain.Product{}
	for rows.Next() {
		var p domain.Product
		var specsBytes []byte

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.CategoryID, &p.CategoryName,
			&p.BrandID, &p.BrandName,
			&p.ImageURL, &p.SKU,
			&specsBytes,
			&p.Rating, &p.ReviewsCount, &p.CreatedAt,
		); err != nil {
			return nil, err
		}

		if len(specsBytes) > 0 {
			_ = json.Unmarshal(specsBytes, &p.Specs)
		}

		products = append(products, p)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.ProductListResponse{
		Products:   products,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// FindByID ค้นหาสินค้าตาม ID พร้อมสเปกและแบรนด์
func (r *postgresRepository) FindByID(id int) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.stock,
			p.category_id, COALESCE(c.name, ''),
			p.brand_id, COALESCE(b.name, ''),
			COALESCE(p.image_url, ''), COALESCE(p.sku, ''),
			COALESCE(p.specs, '{}'::jsonb),
			p.rating, p.reviews_count, p.created_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1
	`
	var p domain.Product
	var specsBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
		&p.CategoryID, &p.CategoryName,
		&p.BrandID, &p.BrandName,
		&p.ImageURL, &p.SKU,
		&specsBytes,
		&p.Rating, &p.ReviewsCount, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	if len(specsBytes) > 0 {
		_ = json.Unmarshal(specsBytes, &p.Specs)
	}

	return &p, nil
}

// FindAllBrands ดึงรายชื่อแบรนด์ทั้งหมด
func (r *postgresRepository) FindAllBrands() ([]domain.Brand, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, COALESCE(logo_url, ''), COALESCE(description, ''), created_at
		FROM brands
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []domain.Brand
	for rows.Next() {
		var b domain.Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.LogoURL, &b.Description, &b.CreatedAt); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, nil
}

// Create เพิ่มสินค้าใหม่ลง Database
func (r *postgresRepository) Create(input domain.CreateProductInput) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	specsJSON, _ := json.Marshal(input.Specs)

	query := `
		INSERT INTO products (name, description, price, stock, category_id, brand_id, image_url, sku, specs)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, rating, reviews_count, created_at
	`
	var p domain.Product
	p.Name = input.Name
	p.Description = input.Description
	p.Price = input.Price
	p.Stock = input.Stock
	p.CategoryID = input.CategoryID
	p.BrandID = input.BrandID
	p.ImageURL = input.ImageURL
	p.SKU = input.SKU
	p.Specs = input.Specs

	err := r.db.QueryRowContext(
		ctx, query,
		input.Name, input.Description, input.Price, input.Stock,
		input.CategoryID, input.BrandID, input.ImageURL, input.SKU, specsJSON,
	).Scan(&p.ID, &p.Rating, &p.ReviewsCount, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// Update แก้ไขสินค้า
func (r *postgresRepository) Update(id int, input domain.UpdateProductInput) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	specsJSON, _ := json.Marshal(input.Specs)

	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4,
		    category_id = $5, brand_id = $6, image_url = $7, sku = $8, specs = $9
		WHERE id = $10
		RETURNING id, rating, reviews_count, created_at
	`
	var p domain.Product
	p.ID = id
	p.Name = input.Name
	p.Description = input.Description
	p.Price = input.Price
	p.Stock = input.Stock
	p.CategoryID = input.CategoryID
	p.BrandID = input.BrandID
	p.ImageURL = input.ImageURL
	p.SKU = input.SKU
	p.Specs = input.Specs

	err := r.db.QueryRowContext(
		ctx, query,
		input.Name, input.Description, input.Price, input.Stock,
		input.CategoryID, input.BrandID, input.ImageURL, input.SKU, specsJSON, id,
	).Scan(&p.ID, &p.Rating, &p.ReviewsCount, &p.CreatedAt)
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
