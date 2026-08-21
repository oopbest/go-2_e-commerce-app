package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config โครงสร้างสำหรับเก็บการตั้งค่าทั้งหมดของแอปพลิเคชัน
type Config struct {
	Port       string
	AppEnv     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

// Load โหลดการตั้งค่าจาก .env หรือ Environment Variables ของระบบ
func Load() *Config {
	// พยายามอ่านไฟล์ .env (ถ้าไม่มีไฟล์ เช่น รันบน Production / Docker จะไม่อ่านข้ามไปอ่าน os.Getenv แทน)
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  No .env file found, reading from system environment variables")
	}

	return &Config{
		Port:       getEnv("SERVER_PORT", "8080"),
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgrespassword"),
		DBName:     getEnv("DB_NAME", "ecommerce_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "default-fallback-secret-key-change-me"),
	}
}

// getEnv Helper ฟังก์ชันอ่านค่า Environment Variable พร้อมค่าเริ่มต้น (Fallback Default)
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
