package domain

import (
	"context"
)

type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string `json:"password" gorm:"type:text;not null"`
	CreatedAt string `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt string `json:"updated_at" gorm:"type:timestamp"`
	DeletedAt string `json:"deleted_at" gorm:"type:timestamp"`
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	Count(ctx context.Context) (int64, error)
}
