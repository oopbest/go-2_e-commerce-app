package domain

import (
	"errors"
	"time"
)

// ==========================================
// 1. Auth & User Domain Errors
// ==========================================

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized: missing or invalid token")
	ErrForbidden          = errors.New("forbidden: insufficient permissions")
)

// ==========================================
// 2. Entities & DTOs
// ==========================================

// User โครงสร้างข้อมูลผู้ใช้งาน
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // json:"-" สำคัญมาก! ป้องกันไม่ให้ส่ง Password Hash ออกไปใน JSON Response เด็ดขาด
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterInput ข้อมูลที่รับเข้ามาตอนสมัครสมาชิก
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // customer หรือ admin
}

// LoginInput ข้อมูลที่รับเข้ามาตอนเข้าสู่ระบบ
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse ข้อมูลตอบกลับเมื่อ Login หรือ Register สำเร็จ
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ==========================================
// 3. Interfaces
// ==========================================

// UserRepository สัญญาการทำงานของ Data Access Layer สำหรับ User
type UserRepository interface {
	Create(user *User) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByID(id int) (*User, error)
}

// UserService สัญญาการทำงานของ Business Logic สำหรับ User & Auth
type UserService interface {
	Register(input RegisterInput) (*AuthResponse, error)
	Login(input LoginInput) (*AuthResponse, error)
	GetProfile(userID int) (*User, error)
}
