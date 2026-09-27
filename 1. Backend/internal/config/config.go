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
	AppEnv           string
	AppName          string
	Port             string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DBMaxConns       int32
	DBMinConns       int32
	JWTSecret        string
	JWTIssuer        string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	CORSOrigins      []string
	RequestTimeout   time.Duration
	BodyLimitBytes   int64
	RateLimitRPM     int
	AuthRateLimitRPM int
	CookieSecure     bool
	AdminFirstName   string
	AdminLastName    string
	AdminPhone       string
	AdminUsername    string
	AdminPassword    string
	EskizEmail       string
	EskizPassword    string
	EskizBaseURL     string
	EskizSender      string
	EskizTimeout     time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	if strings.TrimSpace(os.Getenv("APP_ENV")) != "production" {
		_ = godotenv.Overload()
	}

	accessTTL, err := time.ParseDuration(env("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(env("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}
	timeout, err := time.ParseDuration(env("REQUEST_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
	}
	eskizTimeout, err := time.ParseDuration(env("ESKIZ_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("ESKIZ_TIMEOUT: %w", err)
	}

	cfg := &Config{
		AppEnv:           env("APP_ENV", "development"),
		AppName:          env("APP_NAME", "qoshida-backend"),
		Port:             env("APP_PORT", "8080"),
		DBHost:           must("DB_HOST"),
		DBPort:           env("DB_PORT", "5432"),
		DBUser:           must("DB_USER"),
		DBPassword:       must("DB_PASSWORD"),
		DBName:           must("DB_NAME"),
		DBSSLMode:        env("DB_SSLMODE", "require"),
		DBMaxConns:       int32(envInt("DB_MAX_CONNS", 20)),
		DBMinConns:       int32(envInt("DB_MIN_CONNS", 2)),
		JWTSecret:        must("JWT_SECRET"),
		JWTIssuer:        env("JWT_ISSUER", "qoshida"),
		JWTAccessTTL:     accessTTL,
		JWTRefreshTTL:    refreshTTL,
		CORSOrigins:      split(must("CORS_ORIGINS")),
		RequestTimeout:   timeout,
		BodyLimitBytes:   int64(envInt("BODY_LIMIT_BYTES", 1<<20)),
		RateLimitRPM:     envInt("RATE_LIMIT_RPM", 120),
		AuthRateLimitRPM: envInt("AUTH_RATE_LIMIT_RPM", 10),
		CookieSecure:     env("COOKIE_SECURE", "true") == "true",
		AdminFirstName:   env("ADMIN_FIRST_NAME", "General"),
		AdminLastName:    env("ADMIN_LAST_NAME", "Admin"),
		AdminPhone:       env("ADMIN_PHONE", "+998901234567"),
		AdminUsername:    env("ADMIN_USERNAME", "admin"),
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		EskizEmail:       env("ESKIZ_EMAIL", ""),
		EskizPassword:    env("ESKIZ_PASSWORD", ""),
		EskizBaseURL:     env("ESKIZ_BASE_URL", "https://notify.eskiz.uz/api"),
		EskizSender:      env("ESKIZ_SENDER", "4546"),
		EskizTimeout:     eskizTimeout,
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET kamida 32 belgidan iborat bo'lishi kerak")
	}
	if cfg.AppEnv == "production" && cfg.DBSSLMode == "disable" {
		return nil, fmt.Errorf("production muhitida DB_SSLMODE=disable mumkin emas")
	}
	if cfg.AppEnv == "production" && !cfg.CookieSecure {
		return nil, fmt.Errorf("production muhitida COOKIE_SECURE=true bo'lishi kerak")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) IsDev() bool {
	return c.AppEnv == "development"
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func must(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("%s muhit o'zgaruvchisi majburiy", key))
	}
	return v
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func split(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
