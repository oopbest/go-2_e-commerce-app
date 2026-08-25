package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims ข้อมูลที่จะฝัง (Payload) เข้าไปใน JWT Token
type CustomClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken สร้าง JWT Token ที่มีอายุตามที่กำหนด
func GenerateToken(userID int, email, role, secretKey string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)), // วันหมดอายุ
			IssuedAt:  jwt.NewNumericDate(time.Now()),               // วันที่ออก Token
		},
	}

	// เข้ารหัส Token ด้วย Algorithm HMAC-SHA256 (HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken ตรวจสอบความถูกต้องและแกะ Claims ออกมาจาก JWT Token
func ValidateToken(tokenString, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบว่า Algorithm ใน Header ตรงกับที่คาดหวังหรือไม่ (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
