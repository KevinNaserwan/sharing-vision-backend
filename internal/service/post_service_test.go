package service

import (
	"context"
	"errors"
	"testing"

	"sharing-vision-backend/internal/model"
)

type fakePostRepository struct {
	posts  map[uint]*model.Post
	nextID uint
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{
		posts:  make(map[uint]*model.Post),
		nextID: 1,
	}
}

func (f *fakePostRepository) Create(_ context.Context, post *model.Post) error {
	post.ID = f.nextID
	f.nextID++
	f.posts[post.ID] = post
	return nil
}

func (f *fakePostRepository) List(_ context.Context, _ int, _ int) ([]model.Post, error) {
	posts := make([]model.Post, 0, len(f.posts))
	for _, post := range f.posts {
		posts = append(posts, *post)
	}
	return posts, nil
}

func (f *fakePostRepository) GetByID(_ context.Context, id uint) (*model.Post, error) {
	post, ok := f.posts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return post, nil
}

func (f *fakePostRepository) Update(_ context.Context, id uint, payload *model.Post) error {
	post, ok := f.posts[id]
	if !ok {
		return errors.New("not found")
	}
	*post = *payload
	post.ID = id
	return nil
}

func (f *fakePostRepository) Delete(_ context.Context, id uint) error {
	if _, ok := f.posts[id]; !ok {
		return errors.New("not found")
	}
	delete(f.posts, id)
	return nil
}

func TestCreateAndUpdateValidation(t *testing.T) {
	svc := NewPostService(newFakePostRepository())

	payload := PostPayload{
		Title:    "Valid Title For Sharing Vision",
		Content:  "Artikel ini memuat penjelasan yang lebih dari dua ratus karakter agar validasi panjang konten terpenuhi dengan aman. Isi konten menjelaskan kebutuhan sistem posting, alur simpan, status publish/draft/trash, lalu menunjukkan bahwa data memang lolos ketika diproses melalui service layer dengan rule bisnis yang ketat di backend.",
		Category: "Tech",
		Status:   "publish",
	}

	created, err := svc.Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("create should succeed, got: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created post id")
	}

	invalidPayload := PostPayload{
		Title:    "short",
		Content:  "invalid",
		Category: "x",
		Status:   "invalid",
	}

	if _, err := svc.Create(context.Background(), invalidPayload); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdateNotFound(t *testing.T) {
	svc := NewPostService(newFakePostRepository())
	_, err := svc.Update(context.Background(), 999, PostPayload{
		Title:    "Another valid title for update flow",
		Content:  "Konten lain yang digunakan untuk test pembaruan juga ditulis panjang agar tidak tersentuh validasi minimal dua ratus karakter. Pada alur update, service tetap melakukan trim string dan normalisasi status sehingga hasilnya konsisten di setiap permintaan API.",
		Category: "Education",
		Status:   "draft",
	})
	if err == nil {
		t.Fatal("expected error when data not found")
	}
}
