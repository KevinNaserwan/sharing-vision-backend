package handler

import (
	"fmt"
	"net/http"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sharing-vision-backend/internal/repository"
	"sharing-vision-backend/internal/response"
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
		api.POST("", h.create)
		api.POST("/", h.create)
		api.GET("/*path", h.getOrList)
		api.HEAD("/*path", h.getOrList)
		api.PUT("/:id", h.update)
		api.PATCH("/:id", h.update)
		api.POST("/:id", h.upsertOrDelete)
		api.DELETE("/:id", h.delete)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-backend"})
	})
	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "sharing-vision-backend"})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.HEAD("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

func (h *ArticleHandler) create(c *gin.Context) {
	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidJSON, "invalid json payload", nil))
		return
	}

	post, err := h.service.Create(c.Request.Context(), payload)
	if err != nil {
		if validationErr, ok := asValidationError(err); ok {
			c.JSON(http.StatusBadRequest, validationErr.Response())
			return
		}
		log.Printf("article create failed: type=%T message=%v", err, err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to create article", nil))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": post.ID, "message": "article created"})
}

func (h *ArticleHandler) getOrList(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	if path == "" {
		c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "not found", nil))
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		h.get(c, parts[0])
		return
	}

	if len(parts) != 2 {
		c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "not found", nil))
		return
	}

	h.list(c, parts[0], parts[1])
}

func (h *ArticleHandler) list(c *gin.Context, limitParam string, offsetParam string) {
	limit, err := parsePositiveInt(limitParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "limit must be a positive integer", nil))
		return
	}
	offset, err := parseInt(offsetParam)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "offset must be zero or a positive integer", nil))
		return
	}

	posts, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to load articles", nil))
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *ArticleHandler) get(c *gin.Context, idParam string) {
	id, err := parseUint(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "id must be an integer", nil))
		return
	}

	post, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "article not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "internal server error", nil))
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *ArticleHandler) update(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "id must be an integer", nil))
		return
	}

	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidJSON, "invalid json payload", nil))
		return
	}

	_, err = h.service.Update(c.Request.Context(), id, payload)
	if err != nil {
		if validationErr, ok := asValidationError(err); ok {
			c.JSON(http.StatusBadRequest, validationErr.Response())
			return
		}
		log.Printf("article update failed: type=%T message=%v", err, err)

		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "article not found", nil))
			return
		}

		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to update article", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "article updated"})
}

func (h *ArticleHandler) upsertOrDelete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "id must be an integer", nil))
		return
	}

	if shouldDeleteByQuery(c) {
		err = h.service.Delete(c.Request.Context(), id)
		if err != nil {
			if err == repository.ErrPostNotFound {
				c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "article not found", nil))
				return
			}

			c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to delete article", nil))
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "message": "article deleted"})
		return
	}

	var payload service.PostPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidJSON, "invalid json payload", nil))
		return
	}

	_, err = h.service.Update(c.Request.Context(), id, payload)
	if err != nil {
		if validationErr, ok := asValidationError(err); ok {
			c.JSON(http.StatusBadRequest, validationErr.Response())
			return
		}
		log.Printf("article upsert/update failed: type=%T message=%v", err, err)

		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "article not found", nil))
			return
		}

		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to update article", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "message": "article updated"})
}

func (h *ArticleHandler) delete(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse(response.ErrorCodeInvalidArgument, "id must be an integer", nil))
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if err == repository.ErrPostNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse(response.ErrorCodeNotFound, "article not found", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ErrorCodeInternal, "failed to delete article", nil))
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
