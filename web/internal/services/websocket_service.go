package services

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pageza/landscaping-app/web/internal/config"
)

type WebSocketService struct {
	config   *config.Config
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]ClientInfo
	mutex    sync.RWMutex
}

type ClientInfo struct {
	UserID   string
	TenantID string
	Role     string
}

type Message struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	UserID  string      `json:"user_id,omitempty"`
	Channel string      `json:"channel,omitempty"`
}

func NewWebSocketService(cfg *config.Config) *WebSocketService {
	return &WebSocketService{
		config: cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
		clients: make(map[*websocket.Conn]ClientInfo),
	}
}

// UpgradeConnection upgrades HTTP connection to WebSocket
func (ws *WebSocketService) UpgradeConnection(w http.ResponseWriter, r *http.Request, userID, tenantID, role string) (*websocket.Conn, error) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	// Store client info
	ws.mutex.Lock()
	ws.clients[conn] = ClientInfo{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
	}
	ws.mutex.Unlock()

	// Handle connection cleanup
	conn.SetCloseHandler(func(code int, text string) error {
		ws.mutex.Lock()
		delete(ws.clients, conn)
		ws.mutex.Unlock()
		return nil
	})

	return conn, nil
}

// BroadcastToTenant sends a message to all clients in a tenant
func (ws *WebSocketService) BroadcastToTenant(tenantID string, message Message) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for conn, client := range ws.clients {
		if client.TenantID == tenantID {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				// Connection failed, remove client
				go func(c *websocket.Conn) {
					ws.mutex.Lock()
					delete(ws.clients, c)
					ws.mutex.Unlock()
					c.Close()
				}(conn)
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (ws *WebSocketService) SendToUser(userID string, message Message) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for conn, client := range ws.clients {
		if client.UserID == userID {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				// Connection failed, remove client
				go func(c *websocket.Conn) {
					ws.mutex.Lock()
					delete(ws.clients, c)
					ws.mutex.Unlock()
					c.Close()
				}(conn)
			}
		}
	}
}

// HandleMessage processes incoming WebSocket messages
func (ws *WebSocketService) HandleMessage(conn *websocket.Conn, messageType int, data []byte) error {
	// Parse the message
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	// Get client info
	ws.mutex.RLock()
	client, exists := ws.clients[conn]
	ws.mutex.RUnlock()

	if !exists {
		return nil
	}

	// Handle different message types
	switch msg.Type {
	case "ping":
		// Respond with pong
		response := Message{Type: "pong"}
		return ws.sendToConnection(conn, response)
	case "chat":
		// Handle AI chat messages
		return ws.handleChatMessage(conn, client, msg)
	case "notification_read":
		// Handle notification read status
		return ws.handleNotificationRead(conn, client, msg)
	default:
		// Unknown message type
		return nil
	}
}

func (ws *WebSocketService) sendToConnection(conn *websocket.Conn, message Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (ws *WebSocketService) handleChatMessage(conn *websocket.Conn, client ClientInfo, msg Message) error {
	// In a real implementation, this would:
	// 1. Forward to AI service
	// 2. Get response
	// 3. Send back to client
	
	// For now, just echo back
	response := Message{
		Type: "chat_response",
		Data: map[string]interface{}{
			"message": "AI response to: " + msg.Data.(string),
			"id":      "msg_123",
		},
	}
	
	return ws.sendToConnection(conn, response)
}

func (ws *WebSocketService) handleNotificationRead(conn *websocket.Conn, client ClientInfo, msg Message) error {
	// Handle notification read status update
	// This would typically update the database
	return nil
}

// GetConnectedUsers returns the number of connected users for a tenant
func (ws *WebSocketService) GetConnectedUsers(tenantID string) int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	count := 0
	for _, client := range ws.clients {
		if client.TenantID == tenantID {
			count++
		}
	}
	return count
}