package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Membuka origin untuk mempermudah tes dari Postman/Nuxt.js lokal
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Ambil userID dari JWT middleware
	userID := c.MustGet("user_id").(string)

	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}

	// Kirim menggunakan struct wrapper baru
	client.Hub.Register <- ClientRegister{Client: client, UserID: userID}

	go client.WritePump()
	go client.ReadPump()
}
