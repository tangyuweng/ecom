package websocket

import (
	"log"
	"sync"
)

// Client 代表一個 WebSocket 連線的客戶端
type Client struct {
	Hub          *Hub
	Conn         *Conn
	Send         chan []byte
	UserID       string
	ConnectionID string // 唯一連線 ID
}

// Hub 管理所有的 WebSocket 連線
type Hub struct {
	clients    map[string]*Client  // 已註冊的客戶端（key 是 connectionID）
	userIndex  map[string][]string // userID -> []connectionID 的映射
	register   chan *Client        // 註冊請求
	unregister chan *Client        // 取消註冊請求
	broadcast  chan []byte         // 廣播訊息給所有客戶端
	sendToUser chan *UserMessage   // 發送訊息給特定用戶
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
		userIndex:  make(map[string][]string),
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
			h.clients[client.ConnectionID] = client

			h.userIndex[client.UserID] = append(h.userIndex[client.UserID], client.ConnectionID)
			log.Printf("User %s connected (connectionID: %s), total connections: %d",
				client.UserID, client.ConnectionID, len(h.userIndex[client.UserID]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ConnectionID]; ok {
				delete(h.clients, client.ConnectionID)
				close(client.Send)

				connections := h.userIndex[client.UserID]
				for i, connID := range connections {
					if connID == client.ConnectionID {
						h.userIndex[client.UserID] = append(connections[:i], connections[i+1:]...)
						break
					}
				}

				if len(h.userIndex[client.UserID]) == 0 {
					delete(h.userIndex, client.UserID)
				}

				log.Printf("User %s disconnected (connectionID: %s), remaining connections: %d",
					client.UserID, client.ConnectionID, len(h.userIndex[client.UserID]))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for connID, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, connID)
				}
			}
			h.mu.Unlock()

		case userMsg := <-h.sendToUser:
			h.mu.Lock()
			connectionIDs, ok := h.userIndex[userMsg.UserID]
			if !ok || len(connectionIDs) == 0 {
				log.Printf("User not connected: %s", userMsg.UserID)
				h.mu.Unlock()
				continue
			}

			// 向該用戶的所有連線發送消息
			for _, connID := range connectionIDs {
				if client, ok := h.clients[connID]; ok {
					select {
					case client.Send <- userMsg.Message:
						log.Printf("Message sent to user %s (connectionID: %s)", userMsg.UserID, connID)
					default:
						// 發送失敗，關閉該連線
						close(client.Send)
						delete(h.clients, connID)
						log.Printf("Failed to send to user %s (connectionID: %s), connection closed", userMsg.UserID, connID)
					}
				}
			}
			h.mu.Unlock()
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

// GetConnectedUserCount 獲取當前連線的用戶數量（唯一用戶數）
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userIndex)
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}
