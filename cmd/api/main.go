package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/azharf99/wa-gateway/internal/delivery/http/handler"
	"github.com/azharf99/wa-gateway/internal/delivery/http/middleware"
	"github.com/azharf99/wa-gateway/internal/repository/whatsapp"

	userRepo "github.com/azharf99/wa-gateway/internal/repository/user"
	authUC "github.com/azharf99/wa-gateway/internal/usecase/auth"
	schedUsecase "github.com/azharf99/wa-gateway/internal/usecase/scheduler"
	waUsecase "github.com/azharf99/wa-gateway/internal/usecase/whatsapp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca environment variable dari sistem (Docker/GCP)")
	}

	// Atur Gin Mode sesuai environment (release untuk GCP, debug untuk lokal)
	ginMode := os.Getenv("GIN_MODE")

	var db *sql.DB

	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
		// 1. Inisialisasi Database SQLite (Satu koneksi untuk semua)
		db, err = sql.Open("sqlite", "file:data/sessions.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
		if err != nil {
			panic(err)
		}
	} else {
		// 1. Inisialisasi Database SQLite (Satu koneksi untuk semua)
		db, err = sql.Open("sqlite", "file:sessions.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
		if err != nil {
			panic(err)
		}
	}

	// 2. Setup Repository & Seeder
	uRepo := userRepo.NewSqliteUserRepository(db)
	userRepo.SeedAdminUser(uRepo) // Jalankan seeder

	// 1. Setup Repository
	waRepo := whatsapp.NewWhatsmeowRepository()
	go func() {
		if err := waRepo.Connect(); err != nil {
			fmt.Println("Gagal koneksi WA:", err)
		}
	}()

	// 2. Setup Usecases
	aUC := authUC.NewAuthUsecase(uRepo)
	waUC := waUsecase.NewWhatsAppUsecase(waRepo)
	schedulerUC := schedUsecase.NewSchedulerUsecase(waUC)
	// Mulai jalankan engine gocron
	schedulerUC.Start()
	// Pastikan scheduler dihentikan saat aplikasi mati
	defer schedulerUC.Stop()

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
	}

	// Jalankan Server
	fmt.Println("Server berjalan di port 8003...")
	if err := r.Run(":8003"); err != nil {
		panic(err)
	}
}
