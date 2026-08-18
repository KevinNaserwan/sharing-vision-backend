package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend/internal/model"
	"sharing-vision-backend/internal/repository"
	"sharing-vision-backend/internal/service"
)

type fakeRepoForHandler struct {
	postID uint
}

func (f *fakeRepoForHandler) Create(_ context.Context, post *model.Post) error {
	f.postID++
	post.ID = f.postID
	return nil
}

func (f *fakeRepoForHandler) List(_ context.Context, limit, offset int) ([]model.Post, error) {
	return []model.Post{
		{ID: 1, Title: "A published post for tests", Content: "c", Category: "Tech", Status: "publish"},
		{ID: 2, Title: "A draft post for tests", Content: "c", Category: "Tech", Status: "draft"},
	}, nil
}

func (f *fakeRepoForHandler) GetByID(_ context.Context, id uint) (*model.Post, error) {
	if id == 1 {
		return &model.Post{
			ID:       1,
			Title:    "A published post for tests",
			Content:  "c",
			Category: "Tech",
			Status:   "publish",
		}, nil
	}
	return nil, repository.ErrPostNotFound
}

func (f *fakeRepoForHandler) Update(_ context.Context, _ uint, payload *model.Post) error {
	return nil
}

func (f *fakeRepoForHandler) Delete(_ context.Context, _ uint) error {
	return nil
}

func TestArticleRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepoForHandler{}
	svc := service.NewPostService(repo)
	h := NewArticleHandler(svc)
	router := gin.New()
	h.Register(router)

	reqBody := `{
		"title": "Valid title for creating article in test",
		"content": "Artikel ini dipakai untuk pengujian handler endpoint create. Konten sengaja dibuat panjang melebihi dua ratus karakter supaya service layer menerima payload dan tidak memicu validasi gagal. Setiap field berisi data yang realistis untuk alur publish draft thrash sehingga jalur pembuatan artikel dapat diuji secara menyeluruh dari layer HTTP.",
		"category": "Tech",
		"status": "draft"
	}`

	t.Run("create article", func(t *testing.T) {
		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/article/", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(r, req)
		if r.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d, body=%s", r.Code, r.Body.String())
		}
		var body map[string]interface{}
		if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if _, ok := body["id"]; !ok {
			t.Fatalf("response does not contain id: %s", r.Body.String())
		}
	})

	t.Run("list article", func(t *testing.T) {
		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/article/2/0", nil)
		router.ServeHTTP(r, req)
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", r.Code, r.Body.String())
		}
	})

	t.Run("get article by id", func(t *testing.T) {
		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/article/1", nil)
		router.ServeHTTP(r, req)
		if r.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", r.Code, r.Body.String())
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/article/abc", nil)
		router.ServeHTTP(r, req)
		if r.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", r.Code)
		}
	})
}
