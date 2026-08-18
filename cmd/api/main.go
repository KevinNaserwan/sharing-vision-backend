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
	"sharing-vision-backend/internal/handler"
	"sharing-vision-backend/internal/middleware"
	"sharing-vision-backend/internal/repository"
	"sharing-vision-backend/internal/service"
	"sharing-vision-backend/internal/storage"
)

func main() {
	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := storage.NewMySQL(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("mysql connect failed: %v", err)
	}

	repo := repository.NewPostRepository(db)
	svc := service.NewPostService(repo)
	h := handler.NewArticleHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RecoverJSON(), middleware.SecurityHeaders(), middleware.Cors(cfg.AllowedOrigins), middleware.MaxBodySize(cfg.MaxRequestBodyBytes))
	if cfg.EnableRequestLogging {
		r.Use(gin.Logger())
	} else {
		r.Use(gin.LoggerWithWriter(io.Discard))
	}

	h.Register(r)

	srv := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           r,
		ReadHeaderTimeout:  cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       cfg.RequestTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Printf("sharing-vision backend started on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
