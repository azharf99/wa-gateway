package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/azharf99/wa-gateway/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Harus sama dengan secret di Usecase
var jwtAccessSecret = []byte(os.Getenv("JWT_SECRET"))

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.Response{
				Status:  "error",
				Message: "Authorization header is required",
			})
			return
		}

		// Format header harus "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.Response{
				Status:  "error",
				Message: "Authorization header format must be Bearer {token}",
			})
			return
		}

		tokenString := parts[1]

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

		c.Next() // Lanjutkan ke handler utama
	}
}
