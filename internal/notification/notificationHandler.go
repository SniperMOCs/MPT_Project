package notification

import (
	"encoding/json"
	"net/http"

	"github.com/SniperMOCs/MPT_Project/internal/auth"
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
	jwt     *auth.JWTManager
}

func NewNotificationHandler(service NotificationService, hub *WebSocketHub, jwt *auth.JWTManager) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		hub:     hub,
		jwt:     jwt,
	}
}

// @Summary      WebSocket соединение для уведомлений
// @Description  Устанавливает WebSocket соединение для получения уведомлений в реальном времени.
// @Description  Токен передается в URL параметре 'token' (так как браузерные WebSocket не поддерживают кастомные заголовки).
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Param        token    query     string  true  "JWT токен доступа"
// @Success      101      {string}  string  "Switching Protocols"
// @Failure      401      {string}  string  "Не авторизован"
// @Router       /ws/notifications [get]
func (h *NotificationHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := h.jwt.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// @Summary      Получить уведомления
// @Description  Возвращает список всех уведомлений текущего пользователя
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   notification.NotificationResponse  "Список уведомлений"
// @Failure      401  {string}  string                       "Не авторизован"
// @Failure      404  {string}  string                       "Уведомления не найдены"
// @Router       /notifications [get]
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

// @Summary      Отметить уведомления как прочитанные
// @Description  Отмечает все уведомления текущего пользователя как прочитанные
// @Tags         Уведомления
// @Accept       json
// @Security     BearerAuth
// @Success      200  "Уведомления отмечены как прочитанные"
// @Failure      401  {string}  string  "Не авторизован"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /notifications/read [patch]
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

// @Summary      Получить количество непрочитанных уведомлений
// @Description  Возвращает количество непрочитанных уведомлений текущего пользователя
// @Tags         Уведомления
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {integer} int     "Количество непрочитанных уведомлений"
// @Failure      401  {string}  string  "Не авторизован"
// @Failure      404  {string}  string  "Ошибка получения количества"
// @Router       /notifications/count [get]
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
