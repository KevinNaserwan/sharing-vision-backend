package service

import (
	"context"
	"fmt"
	"strings"

	"sharing-vision-backend/internal/model"
	"sharing-vision-backend/internal/repository"
	"sharing-vision-backend/internal/response"
)

type ValidationIssue struct {
	Field   string
	Message string
}

type ValidationError struct {
	Issues []ValidationIssue
}

func (e ValidationError) Error() string {
	return "validation error"
}

func (e ValidationError) Fields() map[string]string {
	fields := make(map[string]string, len(e.Issues))
	for _, issue := range e.Issues {
		field := strings.TrimSpace(issue.Field)
		message := strings.TrimSpace(issue.Message)
		if field == "" || message == "" {
			continue
		}
		fields[field] = message
	}
	return fields
}

type PostPayload struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

type PostService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) Create(ctx context.Context, payload PostPayload) (*model.Post, error) {
	if err := validatePostPayload(payload); err != nil {
		return nil, err
	}

	post := &model.Post{
		Title:    strings.TrimSpace(payload.Title),
		Content:  strings.TrimSpace(payload.Content),
		Category: strings.TrimSpace(payload.Category),
		Status:   strings.ToLower(strings.TrimSpace(payload.Status)),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) List(ctx context.Context, limit, offset int) ([]model.Post, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *PostService) GetByID(ctx context.Context, id uint) (*model.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) Update(ctx context.Context, id uint, payload PostPayload) (*model.Post, error) {
	if err := validatePostPayload(payload); err != nil {
		return nil, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	current.Title = strings.TrimSpace(payload.Title)
	current.Content = strings.TrimSpace(payload.Content)
	current.Category = strings.TrimSpace(payload.Category)
	current.Status = strings.ToLower(strings.TrimSpace(payload.Status))

	if err := s.repo.Update(ctx, id, current); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *PostService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func validatePostPayload(payload PostPayload) error {
	issues := make([]ValidationIssue, 0)

	title := strings.TrimSpace(payload.Title)
	content := strings.TrimSpace(payload.Content)
	category := strings.TrimSpace(payload.Category)
	status := strings.ToLower(strings.TrimSpace(payload.Status))

	if title == "" {
		issues = append(issues, ValidationIssue{Field: "title", Message: "title is required"})
	} else if len([]rune(title)) < 20 {
		issues = append(issues, ValidationIssue{Field: "title", Message: "minimum 20 characters"})
	}

	if content == "" {
		issues = append(issues, ValidationIssue{Field: "content", Message: "content is required"})
	} else if len([]rune(content)) < 200 {
		issues = append(issues, ValidationIssue{Field: "content", Message: "minimum 200 characters"})
	}

	if category == "" {
		issues = append(issues, ValidationIssue{Field: "category", Message: "category is required"})
	} else if len([]rune(category)) < 3 {
		issues = append(issues, ValidationIssue{Field: "category", Message: "minimum 3 characters"})
	}

	if status == "" {
		issues = append(issues, ValidationIssue{Field: "status", Message: "status is required"})
	} else if !model.PostStatus(status).IsValid() {
		issues = append(issues, ValidationIssue{Field: "status", Message: "status must be publish, draft, or thrash"})
	}

	if len(issues) > 0 {
		return ValidationError{Issues: issues}
	}

	return nil
}

func (e ValidationError) Response() map[string]any {
	return response.ErrorResponse(
		response.ErrorCodeValidation,
		fmt.Sprintf("validation failed (%d issue)", len(e.Issues)),
		e.Fields(),
	)
}
