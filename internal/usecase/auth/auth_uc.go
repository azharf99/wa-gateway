package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Gunakan environment variables (os.Getenv) di production!
var jwtAccessSecret = []byte(os.Getenv("JWT_SECRET"))
var jwtRefreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

type authUsecase struct {
	userRepo domain.UserRepository // Ganti dari AuthRepository ke UserRepository
}

func NewAuthUsecase(repo domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{userRepo: repo}
}

func (uc *authUsecase) Login(ctx context.Context, req domain.LoginReq) (string, string, error) {
	// Ambil user dari DB
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil || user == nil {
		return "", "", errors.New("username atau password salah")
	}

	// Cek password dengan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", errors.New("username atau password salah")
	}

	return uc.generateTokens(user.Username, user.ID)
}

func (uc *authUsecase) RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
	// 1. Parsing dan Validasi Refresh Token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtRefreshSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("refresh token tidak valid atau sudah kedaluwarsa")
	}

	// 2. Ekstrak klaim untuk membuat Access Token baru
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("gagal membaca klaim token")
	}

	username := claims["username"].(string)
	userID := claims["user_id"].(float64)

	// 3. Buat Access Token baru (15 Menit)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"user_id":  userID,
		"role":     "admin",
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})

	newAccessTokenString, err := accessToken.SignedString(jwtAccessSecret)
	if err != nil {
		return "", err
	}

	return newAccessTokenString, nil
}

// Helper untuk men-generate kedua token
func (uc *authUsecase) generateTokens(username string, userID uint) (string, string, error) {
	// Access Token (15 Menit)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"user_id":  userID,
		"role":     "admin",
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(jwtAccessSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (7 Hari)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(jwtRefreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (uc *authUsecase) GetUserById(ctx context.Context, userID uint) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *authUsecase) ChangePassword(ctx context.Context, userID uint, req domain.ChangePasswordReq) error {
	// 1. Ambil data user dari database
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 2. Verifikasi password lama
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return errors.New("password lama tidak sesuai")
	}

	// 3. Hash password baru
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal memproses password baru")
	}

	// 4. Update data user
	user.Password = string(newHash)
	return uc.userRepo.Update(ctx, user)
}
