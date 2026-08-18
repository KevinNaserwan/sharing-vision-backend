package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCorsAllowsWhitelistedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Cors([]string{"https://allowed.example", "https://*.vercel.app"}))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Origin", "https://allowed.example")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().Header.Get("Access-Control-Allow-Origin") != "https://allowed.example" {
		t.Fatalf("expected CORS origin allowed, got: %q", resp.Result().Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsRejectsUnknownOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Cors([]string{"https://allowed.example", "https://*.vercel.app"}))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Origin", "https://unknown.example")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Result().Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected CORS origin rejected, got: %q", resp.Result().Header.Get("Access-Control-Allow-Origin"))
	}
}
