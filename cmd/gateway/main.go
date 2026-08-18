package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"sharing-vision-backend/internal/articlepb"
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

	client, closeClient := mustArticleClient(cfg.ArticleGRPCTarget)
	defer closeClient()
	api := r.Group("/article")
	{
		api.POST("", createArticleHandler(client))
		api.POST("/", createArticleHandler(client))
		api.GET(":limit/:offset", listArticlesHandler(client))
		api.GET(":id", getArticleHandler(client))
		api.PUT(":id", updateArticleHandler(client))
		api.PATCH(":id", updateArticleHandler(client))
		api.POST(":id", upsertOrDeleteArticleHandler(client))
		api.DELETE(":id", deleteArticleHandler(client))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "sharing-vision-gateway",
		})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "sharing-vision-gateway",
		})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.HEAD("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "sharing-vision-gateway",
			"message": "sharing-vision api gateway running",
		})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.EnableEventConsumer {
		go runEventConsumer(ctx, client)
	}

	srv := &http.Server{
		Addr:              cfg.GatewayHTTPAddress,
		Handler:           r,
		ReadHeaderTimeout:  cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       cfg.RequestTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	serverDone := make(chan struct{}, 1)
	go func() {
		log.Printf("gateway started on %s", cfg.GatewayHTTPAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("gateway failed: %v", err)
		}
		serverDone <- struct{}{}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("gateway shutdown error: %v", err)
	}
	<-serverDone
}

func mustArticleClient(target string) (articlepb.ArticleServiceClient, func()) {
	var (
		conn *grpc.ClientConn
		err  error
	)
	retryAt := time.Now().Add(30 * time.Second)
	for {
		conn, err = grpc.NewClient(
			target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(articlepb.JSONCodec{})),
		)
		if err == nil {
			break
		}

		if time.Now().After(retryAt) {
			log.Fatalf("failed to connect article service grpc %s: %v", target, err)
		}
		log.Printf("grpc unavailable, retry 2s: %v", err)
		time.Sleep(2 * time.Second)
	}

	closeConn := func() {
		if err := conn.Close(); err != nil {
			log.Printf("grpc connection close error: %v", err)
		}
	}

	return articlepb.NewArticleServiceClient(conn), closeConn
}

func createArticleHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload articlepb.CreateArticleRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
			return
		}

		resp, err := client.CreateArticle(c.Request.Context(), &payload)
		if err != nil {
			toHTTPError(c, err)
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func listArticlesHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, err := parsePositiveInt(c.Param("limit"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
			return
		}

		offset, err := parseInt(c.Param("offset"))
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be zero or a positive integer"})
			return
		}

		resp, err := client.ListArticles(c.Request.Context(), &articlepb.ListArticlesRequest{Limit: limit, Offset: offset})
		if err != nil {
			toHTTPError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp.Posts)
	}
}

func getArticleHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
			return
		}

		resp, err := client.GetArticle(c.Request.Context(), &articlepb.GetArticleRequest{Id: id})
		if err != nil {
			toHTTPError(c, err)
			return
		}
		if resp == nil || resp.Post == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}

		c.JSON(http.StatusOK, resp.Post)
	}
}

func updateArticleHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
			return
		}

		var payload articlepb.CreateArticleRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
			return
		}

		resp, err := client.UpdateArticle(c.Request.Context(), &articlepb.UpdateArticleRequest{
			Id:       id,
			Title:    payload.Title,
			Content:  payload.Content,
			Category: payload.Category,
			Status:   payload.Status,
		})
		if err != nil {
			toHTTPError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func upsertOrDeleteArticleHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
			return
		}

		if shouldDeleteByQuery(c) {
			resp, err := client.DeleteArticle(c.Request.Context(), &articlepb.DeleteArticleRequest{Id: id})
			if err != nil {
				toHTTPError(c, err)
				return
			}
			c.JSON(http.StatusOK, resp)
			return
		}

		var payload articlepb.CreateArticleRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
			return
		}

		resp, err := client.UpdateArticle(c.Request.Context(), &articlepb.UpdateArticleRequest{
			Id:       id,
			Title:    payload.Title,
			Content:  payload.Content,
			Category: payload.Category,
			Status:   payload.Status,
		})
		if err != nil {
			toHTTPError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func deleteArticleHandler(client articlepb.ArticleServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseUint(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
			return
		}

		resp, err := client.DeleteArticle(c.Request.Context(), &articlepb.DeleteArticleRequest{Id: id})
		if err != nil {
			toHTTPError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func parsePositiveInt(value string) (int, error) {
	result, err := strconv.Atoi(value)
	if err != nil || result <= 0 {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return result, nil
}

func parseInt(value string) (int, error) {
	return strconv.Atoi(value)
}

func parseUint(value string) (uint, error) {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil || result < 1 {
		return 0, fmt.Errorf("invalid uint")
	}
	return uint(result), nil
}

func shouldDeleteByQuery(c *gin.Context) bool {
	action := strings.ToLower(strings.TrimSpace(c.Query("action")))
	force := strings.ToLower(strings.TrimSpace(c.Query("force")))
	return action == "delete" || force == "delete"
}

func toHTTPError(c *gin.Context, err error) {
	statusCode := status.Code(err)
	code := http.StatusInternalServerError
	message := "internal server error"
	if statusErr, ok := status.FromError(err); ok {
		message = statusErr.Message()

		switch statusErr.Code() {
		case codes.InvalidArgument:
			code = http.StatusBadRequest
		case codes.NotFound:
			code = http.StatusNotFound
		case codes.Canceled:
			code = http.StatusRequestTimeout
		case codes.DeadlineExceeded:
			code = http.StatusGatewayTimeout
		case codes.PermissionDenied:
			code = http.StatusForbidden
		case codes.Unauthenticated:
			code = http.StatusUnauthorized
		case codes.Unavailable:
			code = http.StatusServiceUnavailable
		}
	}

	if statusCode != codes.Unknown && statusCode != codes.Internal {
		c.JSON(code, gin.H{"error": message})
		return
	}
	c.JSON(code, gin.H{"error": message})
}

func runEventConsumer(ctx context.Context, client articlepb.ArticleServiceClient) {
	for {
		if ctx.Err() != nil {
			return
		}

		stream, err := client.SubscribeEvents(ctx, &articlepb.SubscribeEventsRequest{})
		if err != nil {
			log.Printf("grpc event stream unavailable: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for {
			event, err := stream.Recv()
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("grpc event stream ended: %v", err)
					time.Sleep(3 * time.Second)
				}
				break
			}
			logArticleEvent(event)
		}
	}
}

func logArticleEvent(event *articlepb.ArticleEvent) {
	if event == nil {
		return
	}

	var id uint
	var statusValue string
	var title string
	if event.Post != nil {
		id = event.Post.ID
		statusValue = event.Post.Status
		title = event.Post.Title
	}

	log.Printf("article event => type=%s post_id=%d status=%s title=%q", event.Type, id, statusValue, title)
}
