package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Untuk WebSocket, browser tidak bisa mengirim Header.
		// Jadi kita PRIORITASKAN Query Parameter "token" terlebih dahulu.
		tokenString := c.Query("token")

		// 2. Jika di Query kosong, baru kita cari di Header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Jika masih kosong juga, tolak.
		if tokenString == "" {
			log.Println("❌ JWT Middleware: Akses ditolak karena Token kosong!")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing Token"})
			c.Abort()
			return
		}

		// Parse dan Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		// INI KUNCI DEBUGGING KITA: Mencetak error spesifik dari JWT
		if err != nil {
			log.Printf("❌ JWT Parse Error: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid Token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("❌ JWT Validation Error: Token sudah expired atau tidak valid!")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Token Invalid"})
			c.Abort()
			return
		}

		// Ekstraksi Claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
		} else {
			log.Println("❌ JWT Claims Error: Format payload token tidak sesuai!")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid Claims"})
			c.Abort()
			return
		}

		// Lolos! Lanjutkan ke Controller/Handler berikutnya
		c.Next()
	}
}
