package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment
// variables (and optionally a .env file for local development).
type Config struct {
	AppEnv          string
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiryHours  int
	UploadDir       string
	MaxUploadSizeMB int64
	AllowedOrigins  string
	AdminEmail      string
	AdminPassword   string
}

// LoadDotEnv reads a simple KEY=VALUE .env file (if present) and sets
// any variables that are not already present in the environment. It never
// overrides variables that are already set, so real environment
// variables (e.g. injected by Docker/systemd) always win.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

// Load builds a Config from environment variables, loading .env first.
func Load() *Config {
	LoadDotEnv(".env")

	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnv("APP_PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiryHours:  getEnvInt("JWT_EXPIRY_HOURS", 72),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB: getEnvInt64("MAX_UPLOAD_SIZE_MB", 5),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "*"),
		AdminEmail:      getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", "AdminPass123!"),
	}
}
