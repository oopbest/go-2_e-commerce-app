package user

import (
	"errors"
	"strings"
	"time"

	"github.com/oopbest/ecommerce-app/internal/domain"
	"github.com/oopbest/ecommerce-app/pkg/security"
)

type service struct {
	repo      domain.UserRepository
	jwtSecret string
}

// NewService Constructor สำหรับสร้าง User Service
func NewService(repo domain.UserRepository, jwtSecret string) domain.UserService {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register สมัครสมาชิกใหม่
func (s *service) Register(input domain.RegisterInput) (*domain.AuthResponse, error) {
	// 1. Validation
	input.Email = domain.NormalizeEmail(input.Email)
	if input.Email == "" || !strings.Contains(input.Email, "@") {
		return nil, errors.New("a valid email is required")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// 2. Hash Password ด้วย bcrypt
	hashedPassword, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. บันทึกลง Database
	newUser := &domain.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         domain.RoleCustomer,
	}

	createdUser, err := s.repo.Create(newUser)
	if err != nil {
		return nil, err
	}

	// 4. สร้าง JWT Token อายุ 24 ชั่วโมง
	token, err := security.GenerateToken(createdUser.ID, createdUser.Email, createdUser.Role, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *createdUser,
	}, nil
}

// Login เข้าสู่ระบบ
func (s *service) Login(input domain.LoginInput) (*domain.AuthResponse, error) {
	// 1. Validation
	input.Email = domain.NormalizeEmail(input.Email)
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// 2. ค้นหา User จาก Email
	u, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 3. เปรียบเทียบรหัสผ่าน
	if !security.CheckPasswordHash(input.Password, u.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// 4. ออก JWT Token
	token, err := security.GenerateToken(u.ID, u.Email, u.Role, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *u,
	}, nil
}

// GetProfile ดึงข้อมูลผู้ใช้ตาม ID
func (s *service) GetProfile(userID int) (*domain.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.FindByID(userID)
}
