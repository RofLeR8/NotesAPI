package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	Port        string
	JWTSecret   string
	JWTExpHours int
}

func Load() *Config {
	godotenv.Load() // загружаем .env

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "notes_user"),
		DBPassword:  getEnv("DB_PASSWORD", "notes_password"),
		DBName:      getEnv("DB_NAME", "notes_db"),
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpHours: getEnvInt("JWT_EXP_HOURS", 12),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var v int
		fmt.Sscanf(val, "%d", &v)
		return v
	}
	return defaultVal
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
