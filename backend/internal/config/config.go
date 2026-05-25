package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	SMTP     SMTPConfig
	Google   GoogleConfig
	Platform PlatformConfig
}

type AppConfig struct {
	Env         string
	Port        string
	FrontendURL string
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	DSN      string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

type MinIOConfig struct {
	Endpoint  string
	User      string
	Password  string
	Bucket    string
	UseSSL    bool
	PublicURL string // public base URL for file links, e.g. https://404tears.kz/minio
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Pass     string
	From     string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

type PlatformConfig struct {
	FeePercent float64
}

var Cfg *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using environment variables")
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	feePercent, _ := strconv.ParseFloat(getEnv("PLATFORM_FEE_PERCENT", "10"), 64)

	accessExpire, _ := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRE", "15m"))
	refreshExpire, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRE", "720h"))

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "smuser")
	dbPass := getEnv("DB_PASSWORD", "smpassword")
	dbName := getEnv("DB_NAME", "studentmarketplace")

	Cfg = &Config{
		App: AppConfig{
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			User:     dbUser,
			Password: dbPass,
			DSN:      "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=UTC",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "access-secret-change-me"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh-secret-change-me"),
			AccessExpire:  accessExpire,
			RefreshExpire: refreshExpire,
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			User:      getEnv("MINIO_USER", "minioadmin"),
			Password:  getEnv("MINIO_PASSWORD", "minioadmin123"),
			Bucket:    getEnv("MINIO_BUCKET", "student-marketplace"),
			UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
			PublicURL: getEnv("MINIO_PUBLIC_URL", ""),
		},
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port: smtpPort,
			User: getEnv("SMTP_USER", ""),
			Pass: getEnv("SMTP_PASS", ""),
			From: getEnv("SMTP_FROM", "noreply@studentmarketplace.com"),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID_WEB", getEnv("GOOGLE_CLIENT_ID", "")),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET_WEB", getEnv("GOOGLE_CLIENT_SECRET", "")),
			CallbackURL:  getEnv("GOOGLE_CALLBACK_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		},
		Platform: PlatformConfig{
			FeePercent: feePercent,
		},
	}

	return Cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
