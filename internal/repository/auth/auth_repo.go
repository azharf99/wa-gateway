package auth

import (
	"context"

	"github.com/azharf99/wa-gateway/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authRepo struct {
	// Simulasi data dari .env atau database
	adminUser string
	adminHash string
}

func NewAuthRepository() domain.AuthRepository {
	// Contoh: Password aslinya adalah "SuperRahasia123"
	// Di environment production, simpan hash ini di database atau .env
	hash, _ := bcrypt.GenerateFromPassword([]byte("SuperRahasia123"), bcrypt.DefaultCost)

	return &authRepo{
		adminUser: "admin_sekolah",
		adminHash: string(hash),
	}
}

func (r *authRepo) VerifyCredential(ctx context.Context, username, password string) bool {
	if username != r.adminUser {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(r.adminHash), []byte(password))
	return err == nil
}
