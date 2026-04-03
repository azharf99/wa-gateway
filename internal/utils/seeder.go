package utils

import (
	"fmt"

	"github.com/azharf99/wa-gateway/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&domain.User{}).Where("username = ?", "admin").Count(&count)

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if count == 0 {
		adminUser := domain.User{
			Username: "admin",
			Password: string(hash),
			ApiKey:   "wa_admin_default_secret_key_123",
		}
		if err := db.Create(&adminUser).Error; err != nil {
			fmt.Println("❌ Gagal membuat akun admin:", err)
		} else {
			fmt.Println("✅ SEEDER: Akun Admin berhasil dibuat (admin / admin123)!")
		}
	} else {
		fmt.Println("✅ SEEDER: Akun Admin sudah eksis, melewati proses seeding.")
	}
}
