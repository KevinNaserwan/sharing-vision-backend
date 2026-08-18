package model

import "time"

type PostStatus string

const (
	StatusPublish PostStatus = "publish"
	StatusDraft   PostStatus = "draft"
	StatusThrash  PostStatus = "thrash"
)

type Post struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string    `json:"title" gorm:"type:varchar(200);not null"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	Category    string    `json:"category" gorm:"type:varchar(100);not null"`
	Status      string    `json:"status" gorm:"type:varchar(100);not null;index"`
	CreatedDate time.Time `json:"created_date" gorm:"column:created_date;autoCreateTime:milli"`
	UpdatedDate time.Time `json:"updated_date" gorm:"column:updated_date;autoUpdateTime:milli"`
}

func (Post) TableName() string {
	return "posts"
}

func (s PostStatus) IsValid() bool {
	return s == StatusPublish || s == StatusDraft || s == StatusThrash
}
