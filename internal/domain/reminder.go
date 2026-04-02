package domain

import "context"

type Reminder struct {
	ID           int64  `json:"id"`
	To           string `json:"to"`
	Message      string `json:"message"`
	IsGroup      bool   `json:"is_group"`
	IntervalDays int    `json:"interval_days"` // Nilai "n" hari pengulangan
	NextRun      string `json:"next_run"`      // Format: YYYY-MM-DD HH:MM:SS
	IsActive     bool   `json:"is_active"`
}

type ReminderRepository interface {
	Create(ctx context.Context, r *Reminder) error
	GetAll(ctx context.Context) ([]Reminder, error)
	GetByID(ctx context.Context, id int64) (*Reminder, error)
	Update(ctx context.Context, r *Reminder) error
	Delete(ctx context.Context, id int64) error
}

type ReminderUsecase interface {
	Start() // WAJIB ADA: Untuk memuat ulang jadwal dari SQLite saat VPS di-restart
	AddReminder(ctx context.Context, r Reminder) error
	ListReminders(ctx context.Context) ([]Reminder, error)
	UpdateReminder(ctx context.Context, r Reminder) error
	DeleteReminder(ctx context.Context, id int64) error
	ProcessReminder(ctx context.Context, id int64) error
}
