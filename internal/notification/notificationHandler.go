package notification

import (
	"encoding/json"
	"net/http"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type NotificationHandler struct {
	service NotificationService
	hub     *WebSocketHub
}

func NewNotificationHandler(service NotificationService, hub *WebSocketHub) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		hub:     hub,
	}
}

func (h *NotificationHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)
	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *NotificationHandler) GetMy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)
	list, err := h.service.GetByRecipient(r.Context(), userID)
	if err != nil {
		http.Error(w, "cannot get notifications", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	if err := h.service.MarkAsRead(r.Context(), userID); err != nil {
		http.Error(w, "cannot read notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)
	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(count)
}
