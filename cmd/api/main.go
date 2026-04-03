package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/azharf99/wa-gateway/internal/delivery/http/handler"
	"github.com/azharf99/wa-gateway/internal/delivery/http/middleware"
	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/azharf99/wa-gateway/internal/repository/contact"
	"github.com/azharf99/wa-gateway/internal/repository/reminder"
	"github.com/azharf99/wa-gateway/internal/repository/whatsapp"

	userRepo "github.com/azharf99/wa-gateway/internal/repository/user"
	authUC "github.com/azharf99/wa-gateway/internal/usecase/auth"
	contactUsecase "github.com/azharf99/wa-gateway/internal/usecase/contact"
	reminderUC "github.com/azharf99/wa-gateway/internal/usecase/reminder"
	schedUsecase "github.com/azharf99/wa-gateway/internal/usecase/scheduler"
	waUsecase "github.com/azharf99/wa-gateway/internal/usecase/whatsapp"
)

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&domain.User{}).Where("username = ?", "admin").Count(&count)

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if count == 0 {
		adminUser := domain.User{
			Username:  "admin",
			Password:  string(hash),
			CreatedAt: time.Now().Local().String(),
			UpdatedAt: time.Now().Local().String(),
		}
		if err := db.Create(&adminUser).Error; err != nil {
			fmt.Println("Gagal membuat akun admin:", err)
		} else {
			fmt.Println("✅ SEEDER: Akun Admin berhasil dibuat (admin / admin123)!")
		}
	} else {
		fmt.Println("✅ SEEDER: Akun Admin sudah eksis, melewati proses seeding.")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca environment variable dari sistem (Docker/GCP)")
	}

	// 1. KONEKSI POSTGRESQL + GORM
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=nama_container_postgres user=postgres password=rahasia dbname=wa_gateway port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Gagal inisialisasi GORM PostgreSQL: %v", err))
	}
	fmt.Println("✅ Database PostgreSQL Terhubung via GORM!")

	// 2. AUTO MIGRATE (Otomatis membuat tabel Contacts, Reminders, Users)
	err = db.AutoMigrate(&domain.User{}, &domain.Contact{}, &domain.Reminder{})
	if err != nil {
		panic(fmt.Sprintf("Gagal AutoMigrate: %v", err))
	}

	// 3. JALANKAN SEEDER
	seedAdmin(db)

	scheduler := gocron.NewScheduler(time.Local)
	scheduler.StartAsync()

	// 4. INIT REPOSITORY GORM
	uRepo := userRepo.NewGormUserRepository(db)
	contactRepo := contact.NewGormContactRepository(db)
	remRepo := reminder.NewGormReminderRepository(db)

	// ==============================================================
	// 5. SETUP WHATSMEOW (Whatsmeow butuh driver asli, bukan GORM)
	// ==============================================================
	// Kita bisa mengekstrak koneksi SQL asli dari GORM:
	sqlDB, _ := db.DB()

	// Kita pasang "postgres" dialect, URI-nya pakai DSN, nil log
	container := sqlstore.NewWithDB(sqlDB, "postgres", nil)
	waRepo := whatsapp.NewWhatsmeowRepository(container) // Sesuaikan constructor repo Bapak

	go func() {
		if err := waRepo.Connect(); err != nil {
			fmt.Println("Gagal koneksi WA:", err)
		}
	}()

	// 2. Setup Usecases
	aUC := authUC.NewAuthUsecase(uRepo)
	waUC := waUsecase.NewWhatsAppUsecase(waRepo, contactRepo)
	schedulerUC := schedUsecase.NewSchedulerUsecase(waUC)
	contactUC := contactUsecase.NewContactUsecase(contactRepo)
	remUC := reminderUC.NewReminderUsecase(remRepo, waUC, scheduler)
	// Mulai jalankan engine gocron
	schedulerUC.Start()
	// Pastikan scheduler dihentikan saat aplikasi mati
	defer schedulerUC.Stop()

	// 4. PENTING: Load kembali reminder dari DB saat startup
	remUC.Start()

	// 3. Setup Router Gin & Handlers
	r := gin.Default()

	// KEAMANAN: Mematikan kepercayaan proxy secara default untuk mencegah spoofing IP.
	// Jika nanti kamu butuh IP asli user di belakang GCP Load Balancer, ubah nil menjadi IP Load Balancer GCP.
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Println("Peringatan gagal mengatur Trusted Proxies:", err)
	}

	// KEAMANAN: Konfigurasi CORS Dinamis
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv == "" {
		allowedOrigins = []string{"http://localhost:5173"} // Fallback aman
	} else {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins, // Alamat Frontend (Vite)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // SANGAT PENTING: Mengizinkan pengiriman HttpOnly Cookie
		MaxAge:           12 * time.Hour,
	}))

	handler.NewAuthHandler(r, aUC)
	protected := r.Group("/")
	protected.Use(middleware.JWTAuthMiddleware()) // Pasang gemboknya di sini
	{
		// Oper route group yang sudah dilindungi ke handler
		handler.NewWhatsAppHandler(protected, waUC)
		handler.NewSchedulerHandler(protected, schedulerUC)
		handler.NewContactHandler(protected, contactUC)
		handler.NewReminderHandler(protected, remUC)
	}

	// Jalankan Server
	fmt.Println("Server berjalan di port 8080...")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
