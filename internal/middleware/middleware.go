package middleware

import (
	"bytes"
	"log"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend/internal/response"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cross-Origin-Resource-Policy", "same-site")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func Cors(allowedOrigins []string) gin.HandlerFunc {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowedVercel := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			continue
		}
		allowed[origin] = struct{}{}
		if strings.HasSuffix(origin, ".vercel.app") {
			allowedVercel = true
		}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if allowedVercel && strings.HasSuffix(origin, ".vercel.app") {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "false")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 && c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
			return
		}

		if maxBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func RecoverJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered panic: %v", rec)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "internal server error", nil))
			}
		}()
		c.Next()
	}
}

func PreserveBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			buf := new(bytes.Buffer)
			_, err := io.Copy(buf, c.Request.Body)
			if err == nil {
				c.Set("raw_body", buf.String())
				c.Request.Body = io.NopCloser(strings.NewReader(buf.String()))
			}
		}
		c.Next()
	}
}
