package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // 寫入訊息的超時時間
	pongWait       = 15 * time.Second    // Pong 等待時間（心跳檢測）- 測試用，改為 15 秒
	pingPeriod     = (pongWait * 9) / 10 // Ping 間隔（必須小於 pongWait）約 13.5 秒
	maxMessageSize = 512                 // 最大訊息大小
)

// Conn 封裝 gorilla/websocket.Conn
type Conn struct {
	*websocket.Conn
}

// ReadPump 從 WebSocket 連線讀取訊息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		// log.Printf("[Pong] Received (userID: %s, connID: %s)", c.UserID, c.ConnectionID)
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Unexpected close (userID: %s, connID: %s): %v", c.UserID, c.ConnectionID, err)
			} else {
				log.Printf("[WebSocket] Connection closed normally (userID: %s, connID: %s)", c.UserID, c.ConnectionID)
			}
			break
		}
		log.Printf("Received message from %s: %s", c.UserID, string(message))
	}
}

// WritePump 向 WebSocket 連線寫入訊息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 關閉了 channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 將佇列中的其他訊息也一起發送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[Ping] Failed to send ping (userID: %s, connID: %s): %v", c.UserID, c.ConnectionID, err)
				return
			}
			// log.Printf("[Ping] Sent (userID: %s, connID: %s)", c.UserID, c.ConnectionID)
		}
	}
}
