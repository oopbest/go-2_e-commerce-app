package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/oopbest/ecommerce-app/migrations"
)

// RunMigrations รัน Migration อัตโนมัติจากไฟล์ SQL ที่ฝังอยู่ใน Binary
func RunMigrations(db *sql.DB) error {
	// 1. อ่านไฟล์ SQL จาก embed.FS
	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	// 2. ผูกกับ PostgreSQL Database Connection
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	// 3. สร้าง Migrate Instance
	m, err := migrate.NewWithInstance("iofs", driver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 4. สั่งรัน Migration UP ทั้งหมด
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply database migrations: %w", err)
	}

	slog.Info("🗄️ Database schema migrated successfully (Up to date)")
	return nil
}
