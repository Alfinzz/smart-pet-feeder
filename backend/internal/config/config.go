package config

import (
	"fmt"
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
	DeviceID            string
	UploadDir           string
	MaxUploadSize       int64
	PublicBaseURL       string
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
	maxUploadSize, err := int64Env("MAX_UPLOAD_SIZE", 5*1024*1024)
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	jwtSecret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	deviceAPIKey, err := requiredEnv("DEVICE_API_KEY")
	if err != nil {
		return Config{}, err
	}
	seedDemoOwner := boolEnv("SEED_DEMO_OWNER", true)
	demoOwnerEmail := ""
	demoOwnerPassword := ""
	if seedDemoOwner {
		demoOwnerEmail, err = requiredEnv("DEMO_OWNER_EMAIL")
		if err != nil {
			return Config{}, err
		}
		demoOwnerPassword, err = requiredEnv("DEMO_OWNER_PASSWORD")
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		AppPort:             env("APP_PORT", "8080"),
		GinMode:             env("GIN_MODE", "debug"),
		CORSAllowOrigin:     env("CORS_ALLOW_ORIGIN", "*"),
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		JWTExpiresIn:        time.Duration(jwtHours) * time.Hour,
		DeviceAPIKey:        deviceAPIKey,
		DeviceID:            env("DEVICE_ID", "ESP32-001"),
		UploadDir:           env("UPLOAD_DIR", "/app/uploads"),
		MaxUploadSize:       maxUploadSize,
		PublicBaseURL:       env("PUBLIC_BASE_URL", ""),
		SeedDemoOwner:       seedDemoOwner,
		DemoOwnerName:       env("DEMO_OWNER_NAME", "Demo Owner"),
		DemoOwnerEmail:      demoOwnerEmail,
		DemoOwnerPassword:   demoOwnerPassword,
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

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func int64Env(key string, fallback int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
