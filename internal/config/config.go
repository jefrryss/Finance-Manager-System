package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort     string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	JWTSecret   string
	JWTTTLHours int
}

func Load() Config {
	cfg := Config{
		AppPort:     getenv("APP_PORT", "8080"),
		DBHost:      getenv("DB_HOST", "db"),
		DBPort:      getenv("DB_PORT", "5432"),
		DBUser:      getenv("DB_USER", "postgres"),
		DBPassword:  getenv("DB_PASSWORD", "postgres"),
		DBName:      getenv("DB_NAME", "expenses"),
		DBSSLMode:   getenv("DB_SSLMODE", "disable"),
		JWTSecret:   getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTLHours: getenvInt("JWT_TTL_HOURS", 168),
	}
	return cfg
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
