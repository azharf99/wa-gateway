package handler

import (
	"net/http"
	"strconv"

	"github.com/azharf99/wa-gateway/internal/delivery/http/middleware"
	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	uc domain.AuthUsecase
}

func NewAuthHandler(r *gin.Engine, uc domain.AuthUsecase) {
	handler := &AuthHandler{uc: uc}

	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh", handler.Refresh)
		authRoutes.GET("/user/:id", handler.GetUser)
		authRoutes.POST("/logout", handler.Logout)
	}
	r.Use(middleware.JWTAuthMiddleware())
	r.PUT("/api/auth/change-password", handler.ChangePassword)

	api := r.Group("/api/auth/api-key")
	api.Use(middleware.JWTAuthMiddleware()) // Gunakan middleware JWT yang sudah ada
	{
		api.GET("/", func(c *gin.Context) {
			userID := c.MustGet("user_id").(uint)
			key, err := uc.GetApiKey(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.Response{Status: "error", Message: err.Error()})
				return
			}
			c.JSON(http.StatusOK, domain.Response{Status: "success", Data: gin.H{"api_key": key}})
		})

		api.POST("/regenerate", func(c *gin.Context) {
			userID := c.MustGet("user_id").(uint)
			newKey, err := uc.GenerateApiKey(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.Response{Status: "error", Message: err.Error()})
				return
			}
			c.JSON(http.StatusOK, domain.Response{Status: "success", Message: "API Key berhasil diregenerate", Data: gin.H{"api_key": newKey}})
		})
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
	isSecure := true

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

func (h *AuthHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "User ID tidak valid"})
		return
	}
	user, err := h.uc.GetUserById(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, domain.Response{Status: "error", Message: "User tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "User ditemukan",
		Data:    user,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req domain.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.Response{Status: "error", Message: "Payload tidak valid"})
		return
	}
	// AMBIL USER ID DARI JWT CONTEXT
	// (Pastikan string "user_id" sesuai dengan key yang kamu set di jwt_middleware.go)
	userIDAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.Response{Status: "error", Message: "Sesi tidak valid atau belum login"})
		return
	}
	// JWT biasanya menyimpan angka sebagai float64 setelah di-parse
	var userID uint
	switch v := userIDAny.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, domain.Response{Status: "error", Message: "Gagal memproses User ID"})
		return
	}

	if err := h.uc.ChangePassword(c.Request.Context(), uint(userID), req); err != nil {
		c.JSON(http.StatusInternalServerError, domain.Response{Status: "error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, domain.Response{
		Status:  "success",
		Message: "Password berhasil diubah",
	})
}
