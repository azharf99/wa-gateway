package domain

import "context"

type Contact struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"type:varchar(255);not null"`
	Phone    string `json:"phone" gorm:"type:varchar(50);uniqueIndex;not null"`
	Category string `json:"category" gorm:"type:varchar(100)"`
}

type ContactRepository interface {
	Create(ctx context.Context, c *Contact) error
	GetAll(ctx context.Context) ([]Contact, error)
	GetByPhone(ctx context.Context, phone string) (*Contact, error)
	Update(ctx context.Context, c *Contact) error
	Delete(ctx context.Context, id int64) error
	ImportCSV(ctx context.Context, contacts []Contact) error
}

type ContactUsecase interface {
	ListContacts(ctx context.Context) ([]Contact, error)
	AddContact(ctx context.Context, c Contact) error
	UpdateContact(ctx context.Context, c Contact) error
	RemoveContact(ctx context.Context, id int64) error
	ImportFromCSV(ctx context.Context, fileBytes []byte) error
}
