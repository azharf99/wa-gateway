package reminder

import (
	"context"

	"github.com/azharf99/wa-gateway/internal/domain"
	"gorm.io/gorm"
)

type gormReminderRepo struct {
	db *gorm.DB
}

func NewGormReminderRepository(db *gorm.DB) domain.ReminderRepository {
	return &gormReminderRepo{db: db}
}

func (r *gormReminderRepo) Create(ctx context.Context, rem *domain.Reminder) error {
	return r.db.WithContext(ctx).Create(rem).Error
}

func (r *gormReminderRepo) GetAll(ctx context.Context) ([]domain.Reminder, error) {
	var reminders []domain.Reminder
	err := r.db.WithContext(ctx).Find(&reminders).Error

	// Format timestamp
	for i := range reminders {
		if len(reminders[i].NextRun) >= 19 {
			reminders[i].NextRun = reminders[i].NextRun[:19]
		}
	}
	return reminders, err
}

func (r *gormReminderRepo) GetByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	var rem domain.Reminder
	err := r.db.WithContext(ctx).First(&rem, id).Error
	if err == nil && len(rem.NextRun) >= 19 {
		rem.NextRun = rem.NextRun[:19]
	}
	return &rem, err
}

func (r *gormReminderRepo) Update(ctx context.Context, rem *domain.Reminder) error {
	return r.db.WithContext(ctx).Save(rem).Error
}

func (r *gormReminderRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.Reminder{}, id).Error
}
