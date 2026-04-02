package user

import (
	"context"
	"fmt"
	"os"

	"github.com/azharf99/wa-gateway/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdminUser(repo domain.UserRepository) {
	ctx := context.Background()
	count, _ := repo.Count(ctx)

	// Jika belum ada user, buat user admin default
	if count == 0 {
		password := os.Getenv("ADMIN_PASSWORD") // Ganti ini saat production
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		admin := &domain.User{
			Username: os.Getenv("ADMIN_USERNAME"),
			Password: string(hash),
		}

		err := repo.Create(ctx, admin)
		if err != nil {
			fmt.Println("Gagal seeding user admin:", err)
		} else {
			fmt.Println("✅ Berhasil membuat user admin default!")
			fmt.Println("   Username: admin")
			fmt.Println("   Password: AdminSekolah123")
		}
	}
}
