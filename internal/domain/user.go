package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Password tidak akan muncul di JSON
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	Count(ctx context.Context) (int64, error)
}
