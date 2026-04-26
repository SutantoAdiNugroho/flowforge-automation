package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DBDSN          string
	MigrateOnStart bool
	JWTSecret      string
	TokenTTL       time.Duration
}

func Load() Config {
	port := getOrDefault("PORT", "5000")
	dsn := getOrDefault("DB_DSN", "postgres://postgres:12345@postgres:5432/flowforge?sslmode=disable")
	migrate := strings.EqualFold(getOrDefault("MIGRATE_ON_START", "true"), "true")
	jwtSecret := getOrDefault("JWT_SECRET", "your-super-secret-key-change-in-production")
	tokenTTLStr := getOrDefault("TOKEN_TTL", "24")
	tokenTTL := time.Duration(getIntOrDefault(tokenTTLStr, 24)) * time.Hour

	return Config{
		Port:           port,
		DBDSN:          dsn,
		MigrateOnStart: migrate,
		JWTSecret:      jwtSecret,
		TokenTTL:       tokenTTL,
	}
}

func getOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getIntOrDefault(val string, fallback int) int {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}
