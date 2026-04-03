package user

import (
	"context"
	"fmt"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdminUser(repo domain.UserRepository) {
	ctx := context.Background()
	count, _ := repo.Count(ctx)

	// Jika belum ada user, buat user admin default
	if count == 0 {
		password := "admin123" // Ganti ini saat production
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		admin := &domain.User{
			Username: "admin",
			Password: string(hash),
		}

		err := repo.Create(ctx, admin)
		if err != nil {
			fmt.Println("Gagal seeding user admin:", err)
		} else {
			fmt.Println("✅ Berhasil membuat user admin default!")
		}
	}
}

// SEEDER MENGGUNAKAN GORM
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
