package websocket

import (
	"log"
	"sync"
)

// Client 代表一個 WebSocket 連線的客戶端
type Client struct {
	Hub    *Hub
	Conn   *Conn
	Send   chan []byte
	UserID string
}

// Hub 管理所有的 WebSocket 連線
type Hub struct {
	clients    map[string]*Client // 已註冊的客戶端（key 是 userID）
	register   chan *Client       // 註冊請求
	unregister chan *Client       // 取消註冊請求
	broadcast  chan []byte        // 廣播訊息給所有客戶端
	sendToUser chan *UserMessage  // 發送訊息給特定用戶
	mu         sync.RWMutex
}

// UserMessage 代表發送給特定用戶的訊息
type UserMessage struct {
	UserID  string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		sendToUser: make(chan *UserMessage),
	}
}

// Run 啟動 Hub 的主循環
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s (total: %d)", client.UserID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
				log.Printf("Client disconnected: %s (total: %d)", client.UserID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()

		case userMsg := <-h.sendToUser:
			h.mu.RLock()
			if client, ok := h.clients[userMsg.UserID]; ok {
				select {
				case client.Send <- userMsg.Message:
					log.Printf("Message sent to user: %s", userMsg.UserID)
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
					log.Printf("Failed to send message to user: %s", userMsg.UserID)
				}
			} else {
				log.Printf("User not connected: %s", userMsg.UserID)
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser 發送訊息給特定用戶
func (h *Hub) SendToUser(userID string, message []byte) {
	h.sendToUser <- &UserMessage{
		UserID:  userID,
		Message: message,
	}
}

// Broadcast 廣播訊息給所有連線的用戶
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// GetConnectedUserCount 獲取當前連線的用戶數量
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}
