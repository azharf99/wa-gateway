package domain

import "context"

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse sekarang hanya mengembalikan Access Token ke body JSON
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type AuthRepository interface {
	VerifyCredential(ctx context.Context, username, password string) bool
}

type AuthUsecase interface {
	// Login sekarang mengembalikan (AccessToken, RefreshToken, Error)
	Login(ctx context.Context, req LoginReq) (string, string, error)
	GetUserById(ctx context.Context, userID uint) (*User, error)
	// Refresh menerima RefreshToken dari cookie dan mengembalikan AccessToken baru
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordReq) error // WAJIB ADA
}
