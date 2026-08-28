package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime application configuration.
type Config struct {
	AppEnv                 string
	AppPort                int
	DatabaseURL            string
	SessionSecret          string
	SessionDurationSeconds int
	CookieSecure           bool
	AutoMigrate            bool
	LogLevel               string
}

// Load reads configuration from environment variables, falling back to sensible defaults.
// It also reads a local .env file if one exists.
func Load() (*Config, error) {
	loadDotEnv(".env")

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		AppPort:                getEnvAsInt("APP_PORT", 8080),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://localhost:5432/flow_db?sslmode=disable"),
		SessionSecret:          getEnv("SESSION_SECRET", "flow_default_dev_session_secret_change_in_prod_123456"),
		SessionDurationSeconds: getEnvAsInt("SESSION_DURATION_SECONDS", 604800), // 7 days
		CookieSecure:           getEnvAsBool("COOKIE_SECURE", false),
		AutoMigrate:            getEnvAsBool("AUTO_MIGRATE", true),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required but not set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	str := getEnv(key, "")
	if str == "" {
		return fallback
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvAsBool(key string, fallback bool) bool {
	str := getEnv(key, "")
	if str == "" {
		return fallback
	}
	val, err := strconv.ParseBool(str)
	if err != nil {
		return fallback
	}
	return val
}

func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // .env is optional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Only set if not already present in environment
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val)
			}
		}
	}
}
