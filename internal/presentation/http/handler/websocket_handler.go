package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tangyuweng/ecom/internal/application/dto"
	ws "github.com/tangyuweng/ecom/internal/infrastructure/websocket"
	"github.com/tangyuweng/ecom/internal/presentation/http/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebsocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// HandleOrderNotifications 處理訂單通知的 WebSocket 連線
func (h *WebSocketHandler) HandlerNotfications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("unauthorized"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 生成唯一的 connectionID
	connectionID := uuid.New().String()

	client := &ws.Client{
		Hub:          h.hub,
		Conn:         &ws.Conn{Conn: conn},
		Send:         make(chan []byte, 256),
		UserID:       userID,
		ConnectionID: connectionID,
	}

	client.Hub.RegisterClient(client)

	// 在新的 goroutine 中處理讀寫
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket connection established for user: %s (connectionID: %s)", userID, connectionID)
}
