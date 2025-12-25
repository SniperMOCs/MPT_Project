package notification

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		connections: make(map[string]*websocket.Conn),
	}
}

func (h *WebSocketHub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID.String()] = conn
}

func (h *WebSocketHub) Unregister(userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, userID.String())
}

func (h *WebSocketHub) SendToUser(userID uuid.UUID, n *Notification) {
	h.mu.RLock()
	conn, ok := h.connections[userID.String()]
	h.mu.RUnlock()
	if !ok {
		return
	}
	conn.WriteJSON(n)
}
