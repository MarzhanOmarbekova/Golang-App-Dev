package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PgURL         string
	JWTSecret     string
	JWTAccessTTL  int // minutes
	JWTRefreshTTL int // minutes
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
	Port          string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	accessTTL, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	refreshTTL, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))

	if accessTTL == 0 {
		accessTTL = 15
	}
	if refreshTTL == 0 {
		refreshTTL = 10080
	}
	if smtpPort == 0 {
		smtpPort = 587
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	return &Config{
		PgURL:         os.Getenv("PG_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      smtpPort,
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:      os.Getenv("SMTP_FROM"),
		Port:          port,
	}, nil
}
