package domain

import (
	"context"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // GORM otomatis membuatkan field ID, CreatedAt, UpdatedAt, DeletedAt dengan tipe data yang sempurna
	Username   string `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password   string `json:"-" gorm:"type:text;not null"`
	ApiKey     string `json:"api_key" gorm:"type:varchar(100);uniqueIndex"`
}

// Request Body dari Frontend
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	Create(ctx context.Context, user *User) error
	GetByApiKey(ctx context.Context, apiKey string) (*User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, user *User) error
}
