package user

import (
	"context"

	"github.com/azharf99/wa-gateway/internal/domain"
	"gorm.io/gorm"
)

type gormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) domain.UserRepository {
	return &gormUserRepo{db: db}
}

func (r *gormUserRepo) Create(ctx context.Context, rem *domain.User) error {
	return r.db.WithContext(ctx).Create(rem).Error
}

func (r *gormUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var rem domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&rem).Error
	return &rem, err
}

func (r *gormUserRepo) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var rem domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rem).Error
	return &rem, err
}

func (r *gormUserRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *gormUserRepo) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
