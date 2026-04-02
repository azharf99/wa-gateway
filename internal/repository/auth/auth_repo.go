package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type authRepo struct {
	// Simulasi data dari .env atau database
	adminUser string
	adminHash string
}

func (r *authRepo) VerifyCredential(ctx context.Context, username, password string) bool {
	if username != r.adminUser {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(r.adminHash), []byte(password))
	return err == nil
}
