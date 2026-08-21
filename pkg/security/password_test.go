package security_test

import (
	"testing"

	"github.com/oopbest/ecommerce-app/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashing(t *testing.T) {
	// นิยาม Table-Driven Test Cases
	testCases := []struct {
		name          string
		password      string
		checkPassword string
		shouldMatch   bool
	}{
		{
			name:          "Valid Password Match",
			password:      "mySecureP@ssw0rd!123",
			checkPassword: "mySecureP@ssw0rd!123",
			shouldMatch:   true,
		},
		{
			name:          "Wrong Password Mismatch",
			password:      "correctPassword",
			checkPassword: "wrongPassword",
			shouldMatch:   false,
		},
		{
			name:          "Empty Password Check Against Hash",
			password:      "someSecret123",
			checkPassword: "",
			shouldMatch:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. แฮชรหัสผ่าน
			hash, err := security.HashPassword(tc.password)
			require.NoError(t, err, "HashPassword should not return an error")
			assert.NotEmpty(t, hash, "Password hash should not be empty")

			// 2. ทดสอบความถูกต้องของ Hash
			matches := security.CheckPasswordHash(tc.checkPassword, hash)
			assert.Equal(t, tc.shouldMatch, matches, "Password match result should be %v", tc.shouldMatch)
		})
	}
}
