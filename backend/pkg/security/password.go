package security

import "golang.org/x/crypto/bcrypt"

// HashPassword แปลง Plain Text Password เป็น Hash ด้วย bcrypt (Cost = 10)
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash เปรียบเทียบ Plain Password กับ Hash ใน Database
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // ถ้าตรงกัน err จะเป็น nil (return true)
}
