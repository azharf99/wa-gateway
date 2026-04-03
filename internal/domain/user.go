package domain

import (
	"context"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // GORM otomatis membuatkan field ID, CreatedAt, UpdatedAt, DeletedAt dengan tipe data yang sempurna
	Username   string `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password   string `json:"password" gorm:"type:text;not null"`
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	Count(ctx context.Context) (int64, error)
}
