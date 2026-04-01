package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/go-co-op/gocron"
)

type schedulerUsecase struct {
	waUC domain.WhatsAppUsecase
	cron *gocron.Scheduler
}

// Constructor meng-inject WhatsAppUsecase
func NewSchedulerUsecase(waUC domain.WhatsAppUsecase) domain.SchedulerUsecase {
	// Inisialisasi scheduler dengan zona waktu lokal (WIB)
	s := gocron.NewScheduler(time.Local)

	return &schedulerUsecase{
		waUC: waUC,
		cron: s,
	}
}

func (s *schedulerUsecase) Start() {
	s.cron.StartAsync()
	fmt.Println("Scheduler berjalan di background...")
}

func (s *schedulerUsecase) Stop() {
	s.cron.Stop()
}

func (s *schedulerUsecase) ScheduleMessage(req domain.ScheduleReq) error {
	// Parsing string waktu dari request JSON ke tipe time.Time
	layout := "2006-01-02 15:04:05"
	runTime, err := time.ParseInLocation(layout, req.RunAt, time.Local)
	if err != nil {
		return fmt.Errorf("format waktu salah, gunakan YYYY-MM-DD HH:MM:SS: %v", err)
	}

	// Mendaftarkan job satu kali jalan (One-off task)
	_, err = s.cron.Every(1).LimitRunsTo(1).StartAt(runTime).Do(func() {
		ctx := context.Background()

		fmt.Printf("[%v] Mengeksekusi pesan terjadwal ke: %s\n", time.Now().Format(layout), req.To)

		_, err := s.waUC.SendMessage(ctx, domain.SendMessageReq{
			To:      req.To,
			Message: req.Message,
		})

		if err != nil {
			fmt.Printf("Gagal mengirim pesan ke %s: %v\n", req.To, err)
		}
	})

	return err
}
