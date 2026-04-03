package contact

import (
	"context"

	"github.com/azharf99/wa-gateway/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormContactRepo struct {
	db *gorm.DB
}

func NewGormContactRepository(db *gorm.DB) domain.ContactRepository {
	return &gormContactRepo{db: db}
}

func (r *gormContactRepo) Create(ctx context.Context, c *domain.Contact) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *gormContactRepo) GetAll(ctx context.Context) ([]domain.Contact, error) {
	var contacts []domain.Contact
	err := r.db.WithContext(ctx).Order("name asc").Find(&contacts).Error
	return contacts, err
}

func (r *gormContactRepo) GetByPhone(ctx context.Context, phone string) (*domain.Contact, error) {
	var c domain.Contact
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&c).Error
	return &c, err
}

func (r *gormContactRepo) Update(ctx context.Context, c *domain.Contact) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *gormContactRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.Contact{}, id).Error
}

// UPSERT (Insert or Update) ala GORM
func (r *gormContactRepo) ImportCSV(ctx context.Context, contacts []domain.Contact) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "phone"}},                       // Konflik terjadi jika phone sama
		DoUpdates: clause.AssignmentColumns([]string{"name", "category"}), // Update kolom ini
	}).Create(&contacts).Error
}
