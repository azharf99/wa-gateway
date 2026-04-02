package domain

import "context"

type Contact struct {
	ID       int64  `json:"id"`
	Name     string `json:"name" binding:"required"`                                                    // Nama lengkap kontak
	Phone    string `json:"phone" binding:"required"`                                                   // Format: 62812...
	Category string `json:"category" default:"Siswa" binding:"oneof=Siswa Orangtua Guru Karyawan Umum"` // Misal: "Siswa", "Wali Murid", "Guru"
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
