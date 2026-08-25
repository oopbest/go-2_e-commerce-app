package main

import (
	"errors"
	"flag"
	"fmt"
	"net/mail"
	"os"

	"github.com/oopbest/ecommerce-app/internal/config"
	"github.com/oopbest/ecommerce-app/internal/database"
	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/internal/user"
	"github.com/oopbest/ecommerce-app/pkg/security"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {

	email := flag.String("email", "", "admin email address")
	flag.Parse()
	normalizedEmail := domain.NormalizeEmail(*email)
	if err := validateEmail(normalizedEmail); err != nil {
		return err
	}

	password, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	confirmation, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}

	if err := validatePasswords(password, confirmation); err != nil {
		return err
	}

	cfg := config.Load()

	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: close database: %v\n", err)
		}
	}()

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	repo := user.NewRepository(db)

	createdUser, err := repo.Create(&domain.User{
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		Role:         domain.RoleAdmin,
	})
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		return fmt.Errorf("user %q already exists", normalizedEmail)
	}
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	fmt.Printf("Admin created: id=%d email=%s\n", createdUser.ID, createdUser.Email)
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	address, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("a valid email is required")
	}

	// ไม่อนุญาตรูปแบบ "John Doe <john@example.com>"
	if address.Address != email {
		return errors.New("email must contain only the email address")
	}
	return nil
}

func validatePasswords(password, confirmation string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if password != confirmation {
		return errors.New("passwords do not match")
	}
	return nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	return string(passwordBytes), nil
}
