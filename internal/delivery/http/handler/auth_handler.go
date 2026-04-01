package handler

import (
	"net/http"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	uc domain.AuthUsecase
}

func NewAuthHandler(r *gin.Engine, uc domain.AuthUsecase) {
	handler := &AuthHandler{uc: uc}

	authRoutes := r.Group("/api/v1/auth")
	{
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh", handler.Refresh)
		authRoutes.POST("/logout", handler.Logout)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Payload tidak valid"})
		return
	}

	accessToken, refreshToken, err := h.uc.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.Response{Status: "error", Message: err.Error()})
		return
	}

	// SET HTTP-ONLY COOKIE UNTUK REFRESH TOKEN
	// Parameter: name, value, maxAge (detik), path, domain, secure, httpOnly
	maxAge := 7 * 24 * 60 * 60 // 7 Hari dalam detik

	// Catatan: Jika testing di localhost HTTP, Secure bisa diset false sementara.
	// Jika VPS sudah pakai HTTPS/SSL, wajib true.
	isSecure := false

	c.SetCookie("refresh_token", refreshToken, maxAge, "/", "", isSecure, true)

	// Kembalikan HANYA access_token di body JSON
	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Login berhasil",
		Data:    domain.TokenResponse{AccessToken: accessToken},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	// Ambil refresh token dari Cookie, bukan dari Body/Header
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.Response{
			Status:  "error",
			Message: "Refresh token tidak ditemukan di cookie",
		})
		return
	}

	// Minta Access Token baru ke Usecase
	newAccessToken, err := h.uc.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Token berhasil diperbarui",
		Data:    domain.TokenResponse{AccessToken: newAccessToken},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Hapus cookie dengan mengatur maxAge menjadi -1
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Logout berhasil",
	})
}
