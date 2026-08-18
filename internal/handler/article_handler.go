package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend/internal/repository"
	"sharing-vision-backend/internal/service"
)

type ArticleHandler struct {
	service *service.PostService
}

func NewArticleHandler(postService *service.PostService) *ArticleHandler {
	return &ArticleHandler{service: postService}
}

func (h *ArticleHandler) Register(r *gin.Engine) {
	api := r.Group("/article")
	{
		api.POST("/", h.create)
		api.GET("/*path", h.getOrList)
		api.PUT("/:id", h.update)
		api.PATCH("/:id", h.update)
		api.POST("/:id", h.upsertOrDelete)
		api.DELETE("/:id", h.delete)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-backend"})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

func (h *ArticleHandler) create(c *gin.Context) {
	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
		return
	}

	post, err := h.service.Create(c.Request.Context(), payload)
	if err != nil {
		if validationErr, ok := err.(service.ValidationError); ok {
			c.JSON(http.StatusBadRequest, validationErr.Response())
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": post.ID, "message": "article created"})
}

func (h *ArticleHandler) getOrList(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		h.get(c, parts[0])
		return
	}

	if len(parts) != 2 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	h.list(c, parts[0], parts[1])
}

func (h *ArticleHandler) list(c *gin.Context, limitParam string, offsetParam string) {
	limit, err := parsePositiveInt(limitParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}
	offset, err := parseInt(offsetParam)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be zero or a positive integer"})
		return
	}

	posts, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load articles"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *ArticleHandler) get(c *gin.Context, idParam string) {
	id, err := parseUint(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
		return
	}

	post, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *ArticleHandler) update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
		return
	}

	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
		return
	}

	_, err = h.service.Update(c.Request.Context(), id, payload)
	if err != nil {
		switch e := err.(type) {
		case service.ValidationError:
			c.JSON(http.StatusBadRequest, e.Response())
			return
		default:
			if err == repository.ErrPostNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "article updated"})
}

func (h *ArticleHandler) upsertOrDelete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
		return
	}

	if shouldDeleteByQuery(c) {
		err = h.service.Delete(c.Request.Context(), id)
		if err != nil {
			if err == repository.ErrPostNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete article"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "message": "article deleted"})
		return
	}

	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json payload"})
		return
	}

	_, err = h.service.Update(c.Request.Context(), id, payload)
	if err != nil {
		switch e := err.(type) {
		case service.ValidationError:
			c.JSON(http.StatusBadRequest, e.Response())
			return
		default:
			if err == repository.ErrPostNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "article updated"})
}

func (h *ArticleHandler) delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be an integer"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "article deleted"})
}

func parsePositiveInt(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return n, nil
}

func parseInt(value string) (int, error) {
	return strconv.Atoi(value)
}

func parseUint(value string) (uint, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, strconv.ErrSyntax
	}
	return uint(n), nil
}

func shouldDeleteByQuery(c *gin.Context) bool {
	action := strings.ToLower(strings.TrimSpace(c.Query("action")))
	force := strings.ToLower(strings.TrimSpace(c.Query("force")))
	return action == "delete" || force == "delete"
}
