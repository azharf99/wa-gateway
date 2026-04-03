package reminder

import (
	"context"
	"fmt"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/go-co-op/gocron"
)

type reminderUsecase struct {
	repo domain.ReminderRepository
	waUC domain.WhatsAppUsecase
	cron *gocron.Scheduler
}

func NewReminderUsecase(repo domain.ReminderRepository, waUC domain.WhatsAppUsecase, cron *gocron.Scheduler) domain.ReminderUsecase {
	return &reminderUsecase{
		repo: repo,
		waUC: waUC,
		cron: cron,
	}
}

// Start memuat semua reminder aktif dari DB ke scheduler saat app boot
func (uc *reminderUsecase) Start() {
	ctx := context.Background()
	reminders, _ := uc.repo.GetAll(ctx)
	for _, rem := range reminders {
		if rem.IsActive {
			uc.scheduleNext(&rem)
		}
	}
}

func (uc *reminderUsecase) AddReminder(ctx context.Context, r domain.Reminder) error {
	r.IsActive = true
	err := uc.repo.Create(ctx, &r)
	if err != nil {
		return err
	}
	uc.scheduleNext(&r)
	return nil
}

func (uc *reminderUsecase) UpdateReminder(ctx context.Context, r domain.Reminder) error {
	err := uc.repo.Update(ctx, &r)
	if err != nil {
		return err
	}
	uc.scheduleNext(&r)
	return nil
}

func (uc *reminderUsecase) ProcessReminder(ctx context.Context, id int64) error {
	rem, err := uc.repo.GetByID(ctx, id)
	if err != nil || !rem.IsActive {
		return nil
	}

	// 1. Kirim pesan
	_, err = uc.waUC.SendMessage(ctx, domain.SendMessageReq{
		To:      rem.To,
		Message: rem.Message,
		IsGroup: rem.IsGroup,
	})
	if err != nil {
		fmt.Printf("[Reminder Error] Gagal kirim ke %s: %v\n", rem.To, err)
	}

	// 2. Hitung jadwal berikutnya
	layout := "2006-01-02 15:04:05"
	lastRun, _ := time.ParseInLocation(layout, rem.NextRun, time.Local)
	nextRunTime := lastRun.AddDate(0, 0, rem.IntervalDays)
	rem.NextRun = nextRunTime.Format(layout)

	// 3. Update DB & Jadwalkan ulang
	uc.repo.Update(ctx, rem)
	uc.scheduleNext(rem)

	return nil
}

func (uc *reminderUsecase) scheduleNext(rem *domain.Reminder) {
	layout := "2006-01-02 15:04:05"
	t, _ := time.ParseInLocation(layout, rem.NextRun, time.Local)

	// Gunakan Tag agar job bisa diidentifikasi/dihapus jika reminder di-update
	tag := fmt.Sprintf("reminder-%d", rem.ID)
	uc.cron.RemoveByTag(tag)

	uc.cron.Every(1).LimitRunsTo(1).StartAt(t).Tag(tag).Do(func() {
		uc.ProcessReminder(context.Background(), int64(rem.ID))
	})
}

// Implementasi List, Update, Delete mengikuti pola CRUD standar...
func (uc *reminderUsecase) ListReminders(ctx context.Context) ([]domain.Reminder, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *reminderUsecase) DeleteReminder(ctx context.Context, id int64) error {
	uc.cron.RemoveByTag(fmt.Sprintf("reminder-%d", id))
	return uc.repo.Delete(ctx, id)
}
