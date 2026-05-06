package websocket

import (
	"log"
)

type ClientRegister struct {
	Client *Client
	UserID string
}

// OPTIMIZATION: Struct baru agar Hub tidak perlu parsing JSON
// saat mem-broadcast pesan. Jauh lebih hemat CPU (O(1) routing).
type WsMessage struct {
	UserID  string
	Payload []byte
}

type Hub struct {
	Clients      map[string]map[*Client]bool
	Broadcast    chan WsMessage // Menggunakan struct WsMessage
	Register     chan ClientRegister
	Unregister   chan *Client
	ClientToUser map[*Client]string
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:    make(chan WsMessage), // Inisiasi channel baru
		Register:     make(chan ClientRegister),
		Unregister:   make(chan *Client),
		Clients:      make(map[string]map[*Client]bool),
		ClientToUser: make(map[*Client]string),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case req := <-h.Register:
			if h.Clients[req.UserID] == nil {
				h.Clients[req.UserID] = make(map[*Client]bool)
			}
			h.Clients[req.UserID][req.Client] = true
			h.ClientToUser[req.Client] = req.UserID
			log.Printf("🔌 WS: Client connected for User: %s\n", req.UserID)

		case client := <-h.Unregister:
			if userID, ok := h.ClientToUser[client]; ok {
				if _, ok := h.Clients[userID][client]; ok {
					delete(h.Clients[userID], client)
					delete(h.ClientToUser, client)
					close(client.Send)
					log.Printf("🔌 WS: Client disconnected from User: %s\n", userID)

					if len(h.Clients[userID]) == 0 {
						delete(h.Clients, userID)
					}
				}
			}

		case message := <-h.Broadcast:
			// SUPER FAST ROUTING: Langsung tembak berdasarkan UserID dari struct,
			// tanpa membedah isi byte payload.
			if userClients, ok := h.Clients[message.UserID]; ok {
				for client := range userClients {
					select {
					case client.Send <- message.Payload: // Hanya kirim byte-nya
					default:
						close(client.Send)
						delete(userClients, client)
						delete(h.ClientToUser, client)
					}
				}
			}
		}
	}
}
