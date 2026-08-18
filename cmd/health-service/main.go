package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"sharing-vision-backend/internal/config"
	"sharing-vision-backend/internal/middleware"
)

func main() {
	cfg := config.Load()
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RecoverJSON(), middleware.SecurityHeaders(), middleware.Cors(cfg.AllowedOrigins), middleware.MaxBodySize(cfg.MaxRequestBodyBytes))
	if cfg.EnableRequestLogging {
		r.Use(gin.Logger())
	} else {
		r.Use(gin.LoggerWithWriter(io.Discard))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-health"})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-health"})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.HEAD("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-health"})
	})

	server := &http.Server{
		Addr:              cfg.HealthHTTPAddress,
		Handler:           r,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       cfg.RequestTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Printf("health service started on %s", cfg.HealthHTTPAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("health service failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("health service shutdown error: %v", err)
	}
}
