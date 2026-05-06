package http

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ridwanafazn/orchestrix/internal/delivery/http/middleware"
	ws "github.com/ridwanafazn/orchestrix/internal/delivery/websocket"
	"github.com/ridwanafazn/orchestrix/internal/usecase"
)

func InitRouter(r *gin.Engine, jobHandler *JobHandler, hub *ws.Hub) {
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// ==========================================
	// KONTROL CORS (CROSS-ORIGIN RESOURCE SHARING)
	// ==========================================
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"}, // Mengizinkan UI Nuxt kita
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// DITAMBAHKAN: X-Chaos-Trigger agar UI bisa mensimulasikan error
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Chaos-Trigger"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "ridwan-secret-2026"
	}

	authUsecase := usecase.NewAuthUsecase(jwtSecret)

	// ==========================================
	// PUBLIC ROUTES
	// ==========================================
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "service": "Orchestrix API"})
	})

	r.POST("/api/v1/auth/guest", func(c *gin.Context) {
		token, err := authUsecase.GenerateGuestToken()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate guest token"})
			return
		}
		c.JSON(200, gin.H{"token": token})
	})

	// ==========================================
	// PROTECTED ROUTES
	// ==========================================
	protected := r.Group("")
	protected.Use(middleware.JWTAuth(jwtSecret))
	{
		// Endpoint API V1 yang dilindungi
		protected.POST("/api/v1/jobs/upload", jobHandler.HandleUpload)

		// Endpoint Websocket yang dilindungi
		protected.GET("/ws", func(c *gin.Context) {
			ws.ServeWS(hub, c)
		})
	}
}
