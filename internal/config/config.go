package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	Environment          string
	ServerAddress        string
	DatabaseDSN          string
	AllowedOrigins       []string
	RequestTimeout       time.Duration
	ReadHeaderTimeout    time.Duration
	MaxRequestBodyBytes  int64
	EnableRequestLogging bool
}

func Load() AppConfig {
	allowed := splitList(getenv("ALLOWED_ORIGINS", "https://be-sharing-vision.meetsin.id,http://localhost:5173,http://localhost:3000,https://*.vercel.app"))

	requestTimeout := getDuration("REQUEST_TIMEOUT_SECONDS", 15)
	readHeaderTimeout := getDuration("READ_HEADER_TIMEOUT_SECONDS", 5)
	maxBody := getenvInt64("MAX_REQUEST_BODY_BYTES", 1024*1024)

	return AppConfig{
		Environment:          getenv("APP_ENV", "production"),
		ServerAddress:        getenv("SERVER_ADDRESS", ":8000"),
		DatabaseDSN:          getenv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local"),
		AllowedOrigins:       allowed,
		RequestTimeout:       time.Duration(requestTimeout) * time.Second,
		ReadHeaderTimeout:    time.Duration(readHeaderTimeout) * time.Second,
		MaxRequestBodyBytes:  maxBody,
		EnableRequestLogging: getenv("ENABLE_REQUEST_LOGGING", "true") == "true",
	}
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		t := strings.TrimSpace(item)
		if t == "" {
			continue
		}
		result = append(result, t)
	}
	return result
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getDuration(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
