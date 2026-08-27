package promotion

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/oopbest/ecommerce-app/internal/domain"
)

type repository struct {
	db *sql.DB
}

// NewRepository Constructor สำหรับ Promotion Repository
func NewRepository(db *sql.DB) domain.PromotionRepository {
	return &repository{db: db}
}

// FindAllActive ดึงคูปองส่วนลดทั้งหมดที่ยังเปิดใช้งานและยังไม่หมดอายุ
func (r *repository) FindAllActive(ctx context.Context) ([]domain.Promotion, error) {
	query := `
		SELECT id, code, title, description, discount_type, discount_value,
		       min_spend, max_discount, total_quota, used_count,
		       starts_at, expires_at, is_active, created_at
		FROM promotions
		WHERE is_active = TRUE 
		  AND expires_at > CURRENT_TIMESTAMP 
		  AND starts_at <= CURRENT_TIMESTAMP
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Promotion
	for rows.Next() {
		var p domain.Promotion
		var desc sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Title, &desc, &p.DiscountType, &p.DiscountValue,
			&p.MinSpend, &p.MaxDiscount, &p.TotalQuota, &p.UsedCount,
			&p.StartsAt, &p.ExpiresAt, &p.IsActive, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		if desc.Valid {
			p.Description = desc.String
		}
		list = append(list, p)
	}

	return list, nil
}

// FindByCode ค้นหาคูปองจากรหัสโค้ด (Case-Insensitive)
func (r *repository) FindByCode(ctx context.Context, code string) (*domain.Promotion, error) {
	query := `
		SELECT id, code, title, description, discount_type, discount_value,
		       min_spend, max_discount, total_quota, used_count,
		       starts_at, expires_at, is_active, created_at
		FROM promotions
		WHERE UPPER(code) = UPPER($1)
	`
	var p domain.Promotion
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(code)).Scan(
		&p.ID, &p.Code, &p.Title, &desc, &p.DiscountType, &p.DiscountValue,
		&p.MinSpend, &p.MaxDiscount, &p.TotalQuota, &p.UsedCount,
		&p.StartsAt, &p.ExpiresAt, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPromotionNotFound
		}
		return nil, err
	}
	if desc.Valid {
		p.Description = desc.String
	}

	return &p, nil
}

// IncrementUsageTx อัปเดตเพิ่มจำนวนการใช้คูปองภายใน Database Transaction อย่างปลอดภัย
func (r *repository) IncrementUsageTx(ctx context.Context, tx *sql.Tx, promotionID int) error {
	query := `
		UPDATE promotions
		SET used_count = used_count + 1
		WHERE id = $1 AND used_count < total_quota
	`
	res, err := tx.ExecContext(ctx, query, promotionID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrPromotionQuotaExceeded
	}

	return nil
}
