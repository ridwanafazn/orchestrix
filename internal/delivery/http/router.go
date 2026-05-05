package http

import (
	"github.com/gin-gonic/gin"
	ws "github.com/ridwanafazn/orchestrix/internal/delivery/websocket"
)

// UPDATE: Tambahkan parameter Hub
func InitRouter(r *gin.Engine, jobHandler *JobHandler, hub *ws.Hub) {
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Endpoint Websocket
	r.GET("/ws", func(c *gin.Context) {
		ws.ServeWS(hub, c)
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "OK", "service": "Orchestrix API"})
		})
		v1.POST("/jobs/upload", jobHandler.HandleUpload)
	}
}
