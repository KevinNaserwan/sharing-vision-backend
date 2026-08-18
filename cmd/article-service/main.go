package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sharing-vision-backend/internal/articlepb"
	"sharing-vision-backend/internal/config"
	"sharing-vision-backend/internal/handler"
	"sharing-vision-backend/internal/middleware"
	"sharing-vision-backend/internal/model"
	"sharing-vision-backend/internal/pubsub"
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
	bus := pubsub.NewEventBus()

	h := handler.NewArticleHandler(svc)
	httpServer := buildHTTPServer(cfg, h)

	grpcListener, err := net.Listen("tcp", cfg.ArticleGRPCAddress)
	if err != nil {
		log.Fatalf("grpc listen failed on %s: %v", cfg.ArticleGRPCAddress, err)
	}

	grpcServer := grpc.NewServer(grpc.ForceServerCodec(articlepb.JSONCodec{}))
	articlepb.RegisterArticleServiceServer(grpcServer, &grpcArticleService{svc: svc, bus: bus})

	httpDone := make(chan struct{}, 1)
	grpcDone := make(chan struct{}, 1)

	go func() {
		log.Printf("article service HTTP started on %s", cfg.ArticleHTTPAddress)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http article service stopped unexpectedly: %v", err)
		}
		httpDone <- struct{}{}
	}()

	go func() {
		log.Printf("article service gRPC started on %s", cfg.ArticleGRPCAddress)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Printf("grpc article service stopped unexpectedly: %v", err)
		}
		grpcDone <- struct{}{}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http article service shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	if err := grpcListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		log.Printf("grpc listener close error: %v", err)
	}

	<-httpDone
	<-grpcDone
}

type grpcArticleService struct {
	svc *service.PostService
	bus *pubsub.EventBus
}

func (g *grpcArticleService) CreateArticle(ctx context.Context, req *articlepb.CreateArticleRequest) (*articlepb.CreateArticleResponse, error) {
	post, err := g.svc.Create(ctx, toPostPayload(req))
	if err != nil {
		return nil, mapSvcErr(err)
	}

	g.bus.Publish(ctx, pubsub.Event{Type: pubsub.EventCreated, Post: post, Occurred: time.Now()})
	return &articlepb.CreateArticleResponse{
		ID:      post.ID,
		Message: "article created",
	}, nil
}

func (g *grpcArticleService) ListArticles(ctx context.Context, req *articlepb.ListArticlesRequest) (*articlepb.ListArticlesResponse, error) {
	posts, err := g.svc.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	responsePosts := make([]*articlepb.Post, 0, len(posts))
	for _, item := range posts {
		if pb := toPBPost(&item); pb != nil {
			responsePosts = append(responsePosts, pb)
		}
	}

	return &articlepb.ListArticlesResponse{Posts: responsePosts}, nil
}

func (g *grpcArticleService) GetArticle(ctx context.Context, req *articlepb.GetArticleRequest) (*articlepb.GetArticleResponse, error) {
	post, err := g.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	return &articlepb.GetArticleResponse{Post: toPBPost(post)}, nil
}

func (g *grpcArticleService) UpdateArticle(ctx context.Context, req *articlepb.UpdateArticleRequest) (*articlepb.UpdateArticleResponse, error) {
	post, err := g.svc.Update(ctx, req.Id, toPostPayloadFromUpdate(req))
	if err != nil {
		return nil, mapSvcErr(err)
	}

	g.bus.Publish(ctx, pubsub.Event{Type: pubsub.EventUpdated, Post: post, Occurred: time.Now()})
	return &articlepb.UpdateArticleResponse{
		ID:      post.ID,
		Message: "article updated",
	}, nil
}

func (g *grpcArticleService) DeleteArticle(ctx context.Context, req *articlepb.DeleteArticleRequest) (*articlepb.DeleteArticleResponse, error) {
	post, err := g.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapSvcErr(err)
	}

	if err := g.svc.Delete(ctx, req.Id); err != nil {
		return nil, mapSvcErr(err)
	}

	g.bus.Publish(ctx, pubsub.Event{Type: pubsub.EventDeleted, Post: post, Occurred: time.Now()})
	return &articlepb.DeleteArticleResponse{
		ID:      req.Id,
		Message: "article deleted",
	}, nil
}

func (g *grpcArticleService) SubscribeEvents(_ *articlepb.SubscribeEventsRequest, stream articlepb.ArticleService_SubscribeEventsServer) error {
	events := g.bus.Subscribe(stream.Context())

	for event := range events {
		if err := stream.Send(toPBEvent(event)); err != nil {
			return err
		}
	}

	return nil
}

func mapSvcErr(err error) error {
	if err == nil {
		return nil
	}

	if validationErr, ok := asValidationError(err); ok {
		badRequest := &errdetails.BadRequest{}
		for _, issue := range validationErr.Issues {
			field := strings.TrimSpace(issue.Field)
			message := strings.TrimSpace(issue.Message)
			if field == "" || message == "" {
				continue
			}
			badRequest.FieldViolations = append(badRequest.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: message,
			})
		}

		base := status.New(codes.InvalidArgument, "validation failed")
		if len(badRequest.FieldViolations) > 0 {
			withDetails, detailErr := base.WithDetails(badRequest)
			if detailErr == nil {
				return withDetails.Err()
			}
		}
		return base.Err()
	}

	if errors.Is(err, repository.ErrPostNotFound) {
		return status.Error(codes.NotFound, "article not found")
	}

	return status.Error(codes.Internal, "internal server error")
}

func asValidationError(err error) (service.ValidationError, bool) {
	switch typed := err.(type) {
	case service.ValidationError:
		return typed, true
	case *service.ValidationError:
		if typed == nil {
			return service.ValidationError{}, false
		}
		return *typed, true
	default:
		return service.ValidationError{}, false
	}
}

func buildHTTPServer(cfg config.AppConfig, h *handler.ArticleHandler) *http.Server {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RecoverJSON(), middleware.SecurityHeaders(), middleware.Cors(cfg.AllowedOrigins), middleware.MaxBodySize(cfg.MaxRequestBodyBytes))
	if cfg.EnableRequestLogging {
		r.Use(gin.Logger())
	} else {
		r.Use(gin.LoggerWithWriter(io.Discard))
	}

	h.Register(r)

	return &http.Server{
		Addr:              cfg.ArticleHTTPAddress,
		Handler:           r,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       cfg.RequestTimeout,
		MaxHeaderBytes:    1 << 20,
	}
}

func toPBPost(post *model.Post) *articlepb.Post {
	if post == nil {
		return nil
	}

	return &articlepb.Post{
		ID:          post.ID,
		Title:       post.Title,
		Content:     post.Content,
		Category:    post.Category,
		Status:      post.Status,
		CreatedDate: post.CreatedDate.Format(time.RFC3339),
		UpdatedDate: post.UpdatedDate.Format(time.RFC3339),
	}
}

func toPBEvent(event pubsub.Event) *articlepb.ArticleEvent {
	var payload *articlepb.Post
	if event.Post != nil {
		payload = toPBPost(event.Post)
	}

	return &articlepb.ArticleEvent{
		Type:    string(event.Type),
		Post:    payload,
		EventAt: event.Occurred.Format(time.RFC3339),
	}
}

func toPostPayload(req *articlepb.CreateArticleRequest) service.PostPayload {
	return service.PostPayload{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Status:   req.Status,
	}
}

func toPostPayloadFromUpdate(req *articlepb.UpdateArticleRequest) service.PostPayload {
	return service.PostPayload{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Status:   req.Status,
	}
}
