package security_test

import (
	"testing"
	"time" // นำเข้า time

	"github.com/oopbest/ecommerce-app/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	secretKey := "super-secret-test-key-32bytes-long!"
	wrongSecretKey := "different-secret-key-for-test-32bytes"

	t.Run("Generate and Validate Valid Token", func(t *testing.T) {
		userID := 42
		email := "tester@example.com"
		role := "admin"

		// 1. สร้าง Token (ส่ง time.Hour เป็น parameter ตัวที่ 5)
		token, err := security.GenerateToken(userID, email, role, secretKey, 24*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// 2. ตรวจสอบ Token
		claims, err := security.ValidateToken(token, secretKey)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// 3. ตรวจสอบความถูกต้องของข้อมูลใน Claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("Validate Token with Wrong Secret Key", func(t *testing.T) {
		// ส่ง time.Hour
		token, err := security.GenerateToken(1, "user@test.com", "customer", secretKey, time.Hour)
		require.NoError(t, err)

		// ตรวจสอบด้วย Secret Key ที่ไม่ตรงกัน -> ต้องได้ Error
		claims, err := security.ValidateToken(token, wrongSecretKey)
		assert.Error(t, err, "Validating token with wrong secret key should fail")
		assert.Nil(t, claims)
	})

	t.Run("Validate Malformed Token", func(t *testing.T) {
		malformedToken := "invalid.jwt.token.string"

		claims, err := security.ValidateToken(malformedToken, secretKey)
		assert.Error(t, err, "Validating malformed token should fail")
		assert.Nil(t, claims)
	})
}
