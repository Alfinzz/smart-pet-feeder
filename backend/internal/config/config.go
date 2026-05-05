package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort             string
	GinMode             string
	CORSAllowOrigin     string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiresIn        time.Duration
	DeviceAPIKey        string
	SeedDemoOwner       bool
	DemoOwnerName       string
	DemoOwnerEmail      string
	DemoOwnerPassword   string
	HistoryDefaultLimit int
	HistoryMaxLimit     int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	jwtHours, err := intEnv("JWT_EXPIRES_IN_HOURS", 24)
	if err != nil {
		return Config{}, err
	}

	defaultLimit, err := intEnv("HISTORY_DEFAULT_LIMIT", 50)
	if err != nil {
		return Config{}, err
	}

	maxLimit, err := intEnv("HISTORY_MAX_LIMIT", 200)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppPort:             env("APP_PORT", "8080"),
		GinMode:             env("GIN_MODE", "debug"),
		CORSAllowOrigin:     env("CORS_ALLOW_ORIGIN", "*"),
		DatabaseURL:         env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/smart_pet_monitoring?sslmode=disable"),
		JWTSecret:           env("JWT_SECRET", "change-me-in-development"),
		JWTExpiresIn:        time.Duration(jwtHours) * time.Hour,
		DeviceAPIKey:        env("DEVICE_API_KEY", "dev-device-key"),
		SeedDemoOwner:       boolEnv("SEED_DEMO_OWNER", true),
		DemoOwnerName:       env("DEMO_OWNER_NAME", "Demo Owner"),
		DemoOwnerEmail:      env("DEMO_OWNER_EMAIL", "owner@example.com"),
		DemoOwnerPassword:   env("DEMO_OWNER_PASSWORD", "password123"),
		HistoryDefaultLimit: defaultLimit,
		HistoryMaxLimit:     maxLimit,
	}, nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
