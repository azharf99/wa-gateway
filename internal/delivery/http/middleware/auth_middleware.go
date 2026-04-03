package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SmartAuthMiddleware menerima JWT dari Frontend ATAU X-API-Key dari layanan Backend lain
func SmartAuthMiddleware(userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. CEK API KEY DARI HEADER (Prioritas untuk B2B)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := userRepo.GetByApiKey(c.Request.Context(), apiKey)
			if err == nil && user != nil {
				// Sesi API Key Valid
				c.Set("user_id", user.ID)
				c.Set("username", user.Username)
				c.Set("role", "api_client") // Penanda bahwa ini diakses via API Key
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "API Key tidak valid"})
			return
		}

		// 2. JIKA TIDAK ADA API KEY, CEK JWT (Untuk sesi Frontend)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Akses ditolak. Sertakan JWT atau X-API-Key"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parsing dan Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// PERBAIKAN: Gunakan jwtAccessSecret
			return jwtAccessSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.Response{
				Status:  "error",
				Message: "Invalid or expired token",
				Data:    err.Error(),
			})
			return
		}

		// Jika sukses, ekstrak payload (claims) dan simpan di context Gin
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("username", claims["username"])
			c.Set("user_id", claims["user_id"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}
