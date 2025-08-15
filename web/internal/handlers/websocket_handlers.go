package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pageza/landscaping-app/web/internal/services"
)

// handleWebSocket upgrades HTTP connection to WebSocket for general real-time features
func (h *Handlers) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := h.services.WebSocket.UpgradeConnection(w, r, user.ID, user.TenantID, user.Role)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send welcome message
	welcomeMsg := services.Message{
		Type: "welcome",
		Data: map[string]interface{}{
			"message": "Connected to real-time updates",
			"user_id": user.ID,
		},
	}
	
	if err := h.sendMessage(conn, welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			if err := h.services.WebSocket.HandleMessage(conn, messageType, message); err != nil {
				log.Printf("Error handling WebSocket message: %v", err)
			}
		}
	}
}

// handleChatWebSocket handles AI assistant chat WebSocket connections
func (h *Handlers) handleChatWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := h.services.WebSocket.UpgradeConnection(w, r, user.ID, user.TenantID, user.Role)
	if err != nil {
		log.Printf("Chat WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send chat welcome message
	welcomeMsg := services.Message{
		Type: "chat_welcome",
		Data: map[string]interface{}{
			"message": "AI Assistant is ready to help",
			"user_id": user.ID,
		},
	}
	
	if err := h.sendMessage(conn, welcomeMsg); err != nil {
		log.Printf("Failed to send chat welcome message: %v", err)
		return
	}

	// Handle chat messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Chat WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			if err := h.handleChatMessage(conn, user, message); err != nil {
				log.Printf("Error handling chat message: %v", err)
			}
		}
	}
}

// sendMessage sends a message through WebSocket connection
func (h *Handlers) sendMessage(conn *websocket.Conn, msg services.Message) error {
	return conn.WriteJSON(msg)
}

// handleChatMessage processes AI chat messages
func (h *Handlers) handleChatMessage(conn *websocket.Conn, user *services.User, message []byte) error {
	// Parse the incoming message
	var chatMsg struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Context string `json:"context,omitempty"`
	}

	if err := conn.ReadJSON(&chatMsg); err != nil {
		return err
	}

	// Handle different chat message types
	switch chatMsg.Type {
	case "chat_message":
		return h.processAIChatMessage(conn, user, chatMsg.Message, chatMsg.Context)
	case "typing":
		return h.handleTypingIndicator(conn, user)
	default:
		return nil
	}
}

// processAIChatMessage sends message to AI service and returns response
func (h *Handlers) processAIChatMessage(conn *websocket.Conn, user *services.User, message, context string) error {
	// Send typing indicator
	typingMsg := services.Message{
		Type: "ai_typing",
		Data: map[string]interface{}{
			"typing": true,
		},
	}
	h.sendMessage(conn, typingMsg)

	// In a real implementation, this would:
	// 1. Send to AI service
	// 2. Get response
	// 3. Store conversation history
	// 4. Return formatted response

	// For now, simulate AI response
	response := h.generateAIResponse(message, context, user)

	// Send response
	responseMsg := services.Message{
		Type: "chat_response",
		Data: map[string]interface{}{
			"message":   response,
			"timestamp": "2024-01-01T12:00:00Z",
			"id":        "msg_" + generateMessageID(),
		},
	}

	return h.sendMessage(conn, responseMsg)
}

// generateAIResponse simulates AI response generation
func (h *Handlers) generateAIResponse(message, context string, user *services.User) string {
	// Simulate different types of responses based on context
	switch context {
	case "scheduling":
		return "I can help you schedule a job. What service do you need and when would you like it done?"
	case "billing":
		return "I can assist with billing questions. Are you looking for invoice information or payment options?"
	case "customer_info":
		return "I can help you find customer information. What would you like to know?"
	default:
		return "I'm here to help with your landscaping business. How can I assist you today?"
	}
}

// handleTypingIndicator handles typing indicator messages
func (h *Handlers) handleTypingIndicator(conn *websocket.Conn, user *services.User) error {
	// In a real implementation, you might broadcast typing status to other users
	return nil
}

// getNotifications returns notifications as HTML fragment
func (h *Handlers) getNotifications(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get notifications from API
	notifications, err := h.getUserNotifications(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := services.TemplateData{
		User:   user,
		IsHTMX: true,
		Data:   notifications,
	}

	content, err := h.services.Template.Render("notifications.html", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// Helper functions

type Notification struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

func (h *Handlers) getUserNotifications(user *services.User) ([]Notification, error) {
	// In real implementation, fetch from API
	notifications := []Notification{
		{
			ID:        "1",
			Type:      "job_update",
			Title:     "Job Completed",
			Message:   "Lawn mowing job at 123 Main St has been completed",
			Read:      false,
			CreatedAt: "2024-01-01T10:00:00Z",
		},
		{
			ID:        "2",
			Type:      "payment",
			Title:     "Payment Received",
			Message:   "Payment of $150.00 received from John Doe",
			Read:      true,
			CreatedAt: "2024-01-01T09:00:00Z",
		},
	}

	return notifications, nil
}

func generateMessageID() string {
	// Simple ID generation - in production, use proper UUID or timestamp
	return "123456"
}

// Broadcast notifications to all connected clients
func (h *Handlers) broadcastNotification(tenantID string, notification Notification) {
	msg := services.Message{
		Type: "notification",
		Data: notification,
	}
	
	h.services.WebSocket.BroadcastToTenant(tenantID, msg)
}

// Send real-time job updates
func (h *Handlers) broadcastJobUpdate(tenantID string, jobUpdate interface{}) {
	msg := services.Message{
		Type: "job_update",
		Data: jobUpdate,
	}
	
	h.services.WebSocket.BroadcastToTenant(tenantID, msg)
}

// Send real-time dashboard updates
func (h *Handlers) broadcastDashboardUpdate(tenantID string, stats interface{}) {
	msg := services.Message{
		Type: "dashboard_update",
		Data: stats,
	}
	
	h.services.WebSocket.BroadcastToTenant(tenantID, msg)
}