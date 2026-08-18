package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"sharing-vision-backend/internal/model"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	List(ctx context.Context, limit, offset int) ([]model.Post, error)
	GetByID(ctx context.Context, id uint) (*model.Post, error)
	Update(ctx context.Context, id uint, post *model.Post) error
	Delete(ctx context.Context, id uint) error
}

type mysqlPostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &mysqlPostRepository{db: db}
}

func (r *mysqlPostRepository) Create(_ context.Context, post *model.Post) error {
	return r.db.Create(post).Error
}

func (r *mysqlPostRepository) List(_ context.Context, limit, offset int) ([]model.Post, error) {
	var posts []model.Post
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	err := r.db.Order("updated_date DESC").Limit(limit).Offset(offset).Find(&posts).Error
	return posts, err
}

func (r *mysqlPostRepository) GetByID(_ context.Context, id uint) (*model.Post, error) {
	var post model.Post
	err := r.db.First(&post, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrPostNotFound
	}
	return &post, err
}

func (r *mysqlPostRepository) Update(_ context.Context, id uint, payload *model.Post) error {
	result := r.db.Model(&model.Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"title":        payload.Title,
		"content":      payload.Content,
		"category":     payload.Category,
		"status":       payload.Status,
		"updated_date": gorm.Expr("NOW()"),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *mysqlPostRepository) Delete(_ context.Context, id uint) error {
	result := r.db.Delete(&model.Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}
