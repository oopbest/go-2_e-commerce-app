package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/oopbest/ecommerce-app/internal/domain"
)

type repository struct {
	db *sql.DB
}

// NewRepository Constructor สำหรับสร้าง User Repository
func NewRepository(db *sql.DB) domain.UserRepository {
	return &repository{
		db: db,
	}
}

// Create บันทึกผู้ใช้งานใหม่ลง Database
func (r *repository) Create(u *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.Role).Scan(
		&u.ID, &u.CreatedAt,
	)
	if err != nil {
		// ตรวจสอบ Unique Constraint Violation (Email ซ้ำ) ใน PostgreSQL คือ Error Code 23505
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, err
	}

	return u, nil
}

// FindByEmail ค้นหาผู้ใช้จาก Email
func (r *repository) FindByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials // ไม่พบ Email
		}
		return nil, err
	}

	return &u, nil
}

// FindByID ค้นหาผู้ใช้จาก ID
func (r *repository) FindByID(id int) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &u, nil
}
