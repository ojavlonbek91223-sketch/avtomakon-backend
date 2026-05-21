package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv  string
	AppPort string
	AppName string

	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
	MinIO MinIOConfig

	CORSAllowedOrigins      string
	RateLimitRequests       int
	RateLimitWindowSeconds  int
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode, c.MaxConns,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	// PublicURL — backend orqali fayllarni ko'rsatish uchun tashqi manzil
	// (masalan, Cloudflare Tunnel URL). Bo'sh bo'lsa to'g'ridan-to'g'ri
	// MinIO endpoint ishlatiladi (faqat lokal tarmoqda ishlaydi).
	PublicURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		// Render/Heroku kabi platformalar PORT env'ni beradi; bo'lmasa APP_PORT.
		AppPort: getEnv("PORT", getEnv("APP_PORT", "8000")),
		AppName: getEnv("APP_NAME", "AvtoMakon"),

		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "avtomakon"),
			Password: getEnvRequired("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "avtomakon"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvInt("DB_MAX_CONNS", 25),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},

		JWT: JWTConfig{
			Secret:     getEnvRequired("JWT_SECRET"),
			AccessTTL:  time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
			RefreshTTL: time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 30)) * 24 * time.Hour,
		},

		MinIO: MinIOConfig{
			Endpoint: getEnv("MINIO_ENDPOINT", "localhost:9000"),
			// Ixtiyoriy: sozlanmagan bo'lsa fayl yuklash o'chiriladi, backend baribir ishlaydi.
			AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey: getEnv("MINIO_SECRET_KEY", ""),
			Bucket:    getEnv("MINIO_BUCKET", "avtomakon"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			PublicURL: getEnv("MINIO_PUBLIC_URL", ""),
		},

		CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", "*"),
		RateLimitRequests:      getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}

	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET kamida 32 belgi bo'lishi kerak")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("majburiy environment variable o'rnatilmagan: %s", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
