package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

const (
	statusPublish = "publish"
	statusDraft   = "draft"
	statusThrash  = "thrash"
)

var allowedStatuses = map[string]struct{}{
	statusPublish: {},
	statusDraft:   {},
	statusThrash:  {},
}

type post struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	Category    string    `json:"category" gorm:"type:varchar(100);not null"`
	Status      string    `json:"status" gorm:"type:varchar(100);not null;index"`
	CreatedDate time.Time `json:"created_date" gorm:"autoCreateTime:milli"`
	UpdatedDate time.Time `json:"updated_date" gorm:"autoUpdateTime:milli"`
}

type postPayload struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

func main() {
	r := gin.Default()
	r.Use(corsMiddleware())

	dsn := getEnv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local")
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect mysql: %v", err))
	}
	db = conn

	// Keep migration optional. For production use: go run ./migrations -path .

	api := r.Group("/article")
	{
		api.POST("/", createArticle)
		api.GET("/:limit/:offset", listArticles)
		api.GET("/:id", getArticle)
		api.POST("/:id", postArticle)
		api.PUT("/:id", postArticle)
		api.PATCH("/:id", postArticle)
		api.DELETE("/:id", deleteArticle)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	err = r.Run(":8000")
	if err != nil {
		panic(err)
	}
}

func createArticle(c *gin.Context) {
	var payload postPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	if err := validatePostPayload(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := post{
		Title:    payload.Title,
		Content:  payload.Content,
		Category: payload.Category,
		Status:   strings.ToLower(payload.Status),
	}

	if err := db.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": record.ID, "message": "article created"})
}

func listArticles(c *gin.Context) {
	limit, err := toInt(c.Param("limit"))
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be positive integer"})
		return
	}
	offset, err := toInt(c.Param("offset"))
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be zero-or-positive integer"})
		return
	}

	var posts []post
	result := db.Order("updated_date desc").Limit(limit).Offset(offset).Find(&posts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch articles"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func getArticle(c *gin.Context) {
	id, err := toInt(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be integer"})
		return
	}

	record, err := findArticleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

func postArticle(c *gin.Context) {
	id, err := toInt(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be integer"})
		return
	}

	if shouldDelete(c) {
		if err := removeArticle(uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "article deleted"})
		return
	}

	var payload postPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}

	if err := validatePostPayload(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := findArticleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	record.Title = payload.Title
	record.Content = payload.Content
	record.Category = payload.Category
	record.Status = strings.ToLower(payload.Status)

	if err := db.Save(record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article updated", "id": record.ID})
}

func deleteArticle(c *gin.Context) {
	id, err := toInt(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be integer"})
		return
	}

	if err := removeArticle(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article deleted"})
}

func validatePostPayload(payload postPayload) error {
	title := strings.TrimSpace(payload.Title)
	content := strings.TrimSpace(payload.Content)
	category := strings.TrimSpace(payload.Category)
	status := strings.ToLower(strings.TrimSpace(payload.Status))

	if len(title) < 20 {
		return errors.New("title is required and minimum 20 characters")
	}
	if len(content) < 200 {
		return errors.New("content is required and minimum 200 characters")
	}
	if len(category) < 3 {
		return errors.New("category is required and minimum 3 characters")
	}
	if _, ok := allowedStatuses[status]; !ok {
		return errors.New("status must be publish, draft, or thrash")
	}
	return nil
}

func shouldDelete(c *gin.Context) bool {
	if c.Query("force") == "delete" {
		return true
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	return false
}

func findArticleByID(id uint) (*post, error) {
	var p post
	if err := db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func removeArticle(id uint) error {
	result := db.Delete(&post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func toInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusOK)
			return
		}
		c.Next()
	}
}
