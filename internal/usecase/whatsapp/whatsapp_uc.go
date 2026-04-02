package whatsapp

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
)

type waUsecase struct {
	repo        domain.WhatsAppRepository
	contactRepo domain.ContactRepository
}

// Constructor Usecase meng-inject Repository
func NewWhatsAppUsecase(repo domain.WhatsAppRepository, contactRepo domain.ContactRepository) domain.WhatsAppUsecase {
	return &waUsecase{
		repo:        repo,
		contactRepo: contactRepo, // Inject Contact Repository untuk akses data kontak
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
		if req.IsGroup {
			jid = jid + "@g.us"
		} else {
			jid = jid + "@s.whatsapp.net"
		}
	}

	if strings.Contains(req.Message, "{{nama}}") {
		// Cari kontak berdasarkan nomor HP (bersihkan format JID jika perlu)
		cleanPhone := strings.Split(req.To, "@")[0]
		contact, err := uc.contactRepo.GetByPhone(ctx, cleanPhone)

		if err == nil && contact != nil {
			req.Message = strings.ReplaceAll(req.Message, "{{nama}}", contact.Name)
		} else {
			// Jika tidak ditemukan, ganti dengan sapaan umum atau hapus placeholder
			req.Message = strings.ReplaceAll(req.Message, "{{nama}}", "Bapak/Ibu")
		}
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

func (uc *waUsecase) SendBroadcast(req domain.BroadcastReq) error {
	// 1. Baca isi file CSV
	reader := csv.NewReader(bytes.NewReader(req.FileBytes))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("gagal membaca file CSV: %v", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("file CSV kosong atau tidak memiliki baris data")
	}

	// 2. Ekstrak Header untuk dynamic variable mapping
	headers := records[0]

	// 3. Jalankan proses Broadcast di Background (Goroutine)
	// Agar request HTTP langsung selesai dan tidak Timeout di frontend
	go func() {
		fmt.Printf("Memulai proses broadcast ke %d kontak...\n", len(records)-1)
		sukses := 0
		gagal := 0

		// Mulai dari index 1 (melewati header)
		for i := 1; i < len(records); i++ {
			row := records[i]

			// Jika baris kosong, lewati
			if len(row) == 0 || row[0] == "" {
				continue
			}

			// Ambil nomor tujuan (Asumsi kolom pertama selalu nomor telepon)
			target := row[0]

			// Rakit pesan dari template
			pesanPersonal := req.MessageTemplate
			for j, headerName := range headers {
				if j < len(row) {
					// Ganti {{nama_header}} dengan isi baris
					placeholder := fmt.Sprintf("{{%s}}", headerName)
					pesanPersonal = strings.ReplaceAll(pesanPersonal, placeholder, row[j])
				}
			}

			// Buat request pengiriman teks biasa
			sendReq := domain.SendMessageReq{
				To:      target,
				Message: pesanPersonal,
				IsGroup: false,
			}

			// Kirim pesan
			_, err := uc.SendMessage(context.Background(), sendReq)

			if err != nil {
				fmt.Printf("[Broadcast Error] Gagal mengirim ke %s: %v\n", target, err)
				gagal++
			} else {
				fmt.Printf("[Broadcast Sukses] Pesan terkirim ke %s\n", target)
				sukses++
			}

			// ========================================================
			// LAYER ANTI-BANNED (WAJIB ADA)
			// Jeda acak / tetap antar pesan agar tidak terdeteksi bot spam
			// ========================================================
			time.Sleep(3 * time.Second)
		}

		fmt.Println("========================================")
		fmt.Printf("Broadcast Selesai! Sukses: %d, Gagal: %d\n", sukses, gagal)
		fmt.Println("========================================")
	}()

	return nil
}
