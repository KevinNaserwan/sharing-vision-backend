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
	ArticleHTTPAddress   string
	ArticleGRPCAddress   string
	GatewayHTTPAddress   string
	HealthHTTPAddress    string
	ArticleGRPCTarget    string
	EnableEventConsumer  bool
	DatabaseDSN          string
	AllowedOrigins       []string
	RequestTimeout       time.Duration
	ReadHeaderTimeout    time.Duration
	MaxRequestBodyBytes  int64
	EnableRequestLogging bool
}

func Load() AppConfig {
	serverAddress := getenv("SERVER_ADDRESS", ":8000")
	allowed := splitList(getenv("ALLOWED_ORIGINS", "https://be-sharing-vision.meetsin.id,http://localhost:5173,http://localhost:3000,https://*.vercel.app"))

	requestTimeout := getDuration("REQUEST_TIMEOUT_SECONDS", 15)
	readHeaderTimeout := getDuration("READ_HEADER_TIMEOUT_SECONDS", 5)
	maxBody := getenvInt64("MAX_REQUEST_BODY_BYTES", 1024*1024)

	return AppConfig{
		Environment:          getenv("APP_ENV", "production"),
		ServerAddress:        serverAddress,
		ArticleHTTPAddress:   getenv("ARTICLE_HTTP_ADDRESS", ":8001"),
		ArticleGRPCAddress:   getenv("ARTICLE_GRPC_ADDRESS", ":9001"),
		GatewayHTTPAddress:   getenv("GATEWAY_HTTP_ADDRESS", serverAddress),
		HealthHTTPAddress:    getenv("HEALTH_HTTP_ADDRESS", ":8002"),
		ArticleGRPCTarget:    getenv("ARTICLE_SERVICE_GRPC_TARGET", "127.0.0.1:9001"),
		EnableEventConsumer:  getenvBool("ENABLE_EVENT_CONSUMER", true),
		DatabaseDSN:          getenv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local"),
		AllowedOrigins:       allowed,
		RequestTimeout:       time.Duration(requestTimeout) * time.Second,
		ReadHeaderTimeout:    time.Duration(readHeaderTimeout) * time.Second,
		MaxRequestBodyBytes:  maxBody,
		EnableRequestLogging: getenv("ENABLE_REQUEST_LOGGING", "true") == "true",
	}
}

func getenvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return result
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
