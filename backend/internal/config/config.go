package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	DBDSN          string
	MigrateOnStart bool
}

func Load() Config {
	port := getOrDefault("PORT", "5000")
	dsn := getOrDefault("DB_DSN", "postgres://postgres:postgres@localhost:5432/flowforge?sslmode=disable")
	migrate := strings.EqualFold(getOrDefault("MIGRATE_ON_START", "true"), "true")
	
	return Config{
		Port:           port,
		DBDSN:          dsn,
		MigrateOnStart: migrate,
	}
}

func getOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
