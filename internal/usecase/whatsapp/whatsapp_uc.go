package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
)

type waUsecase struct {
	repo domain.WhatsAppRepository
}

// Constructor Usecase meng-inject Repository
func NewWhatsAppUsecase(repo domain.WhatsAppRepository) domain.WhatsAppUsecase {
	return &waUsecase{
		repo: repo,
	}
}

func (uc *waUsecase) CheckStatus() string {
	if uc.repo.IsConnected() {
		return "connected"
	}
	return "disconnected"
}

func (uc *waUsecase) GetQRCode() string {
	return uc.repo.GetQRCode()
}

func (uc *waUsecase) Logout() error {
	return uc.repo.Logout()
}

func (uc *waUsecase) SendMessage(ctx context.Context, req domain.SendMessageReq) (string, error) {
	if req.To == "" || req.Message == "" {
		return "", errors.New("nomor tujuan dan pesan tidak boleh kosong")
	}

	// Logic format nomor: pastikan menggunakan @s.whatsapp.net jika belum ada
	jid := req.To
	if !strings.Contains(jid, "@") {
		jid = jid + "@s.whatsapp.net"
	}

	return uc.repo.SendTextMessage(ctx, jid, req.Message)
}

func (uc *waUsecase) BroadcastMessages(req domain.BroadcastReq) {
	// Menjalankan proses di background agar tidak memblokir HTTP response
	go func() {
		total := len(req.Recipients)
		fmt.Printf("Mulai memproses broadcast ke %d nomor...\n", total)

		for i, recipient := range req.Recipients {
			// Menggunakan ulang fungsi SendMessage yang sudah kita buat
			_, err := uc.SendMessage(context.Background(), domain.SendMessageReq{
				To:      recipient.To,
				Message: recipient.Message,
			})

			if err != nil {
				fmt.Printf("[Broadcast] Gagal kirim ke %s: %v\n", recipient.To, err)
			} else {
				fmt.Printf("[Broadcast] %d/%d Berhasil kirim ke %s\n", i+1, total, recipient.To)
			}

			// Berikan jeda untuk SEMUA pesan kecuali pesan terakhir
			if i < total-1 {
				// Menghasilkan jeda acak antara 8 hingga 15 detik
				randomSeconds := rand.Intn(8) + 8
				delay := time.Duration(randomSeconds) * time.Second

				fmt.Printf("Menunggu %v sebelum pesan berikutnya...\n", delay)
				time.Sleep(delay)
			}
		}

		fmt.Println("🎉 Semua pesan broadcast telah selesai diproses!")
	}()
}

func (uc *waUsecase) SendMedia(ctx context.Context, req domain.SendMediaReq) (string, error) {
	if req.To == "" || len(req.FileBytes) == 0 {
		return "", errors.New("tujuan dan file media tidak boleh kosong")
	}

	// Penentuan format JID (Grup atau Nomor Pribadi)
	jid := req.To
	if !strings.Contains(jid, "@") {
		if req.IsGroup {
			jid = jid + "@g.us"
		} else {
			jid = jid + "@s.whatsapp.net"
		}
	}

	// Fallback MimeType jika kosong
	if req.MimeType == "" {
		req.MimeType = "application/octet-stream"
	}

	return uc.repo.SendMediaMessage(ctx, jid, req)
}

func (uc *waUsecase) GetJoinedGroups(ctx context.Context) ([]domain.GroupInfo, error) {
	return uc.repo.GetJoinedGroups(ctx)
}
