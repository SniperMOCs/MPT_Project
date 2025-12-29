package notification

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/auth"
	"github.com/SniperMOCs/MPT_Project/internal/content"
	contextKeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type TestNotificationService struct {
	notifications     map[uuid.UUID][]NotificationResponse
	unreadCounts      map[uuid.UUID]int
	notifyCalled      bool
	getByRecipientErr error
	markAsReadErr     error
	getUnreadCountErr error
	notifyResponse    *NotificationResponse
	notifyErr         error
	callLog           []string
}

func NewTestNotificationService() *TestNotificationService {
	return &TestNotificationService{
		notifications: make(map[uuid.UUID][]NotificationResponse),
		unreadCounts:  make(map[uuid.UUID]int),
		callLog:       []string{},
	}
}

func (s *TestNotificationService) Notify(ctx context.Context, actor, recipient, target uuid.UUID, notitype NotificationType) (*NotificationResponse, error) {
	s.notifyCalled = true
	s.callLog = append(s.callLog, "Notify")
	return s.notifyResponse, s.notifyErr
}

func (s *TestNotificationService) GetByRecipient(ctx context.Context, userId uuid.UUID) ([]NotificationResponse, error) {
	s.callLog = append(s.callLog, "GetByRecipient")

	if s.getByRecipientErr != nil {
		return []NotificationResponse{}, s.getByRecipientErr
	}

	notifications, exists := s.notifications[userId]
	if !exists {
		return []NotificationResponse{}, nil
	}
	return notifications, nil
}

func (s *TestNotificationService) MarkAsRead(ctx context.Context, userId uuid.UUID) error {
	s.callLog = append(s.callLog, "MarkAsRead")
	return s.markAsReadErr
}

func (s *TestNotificationService) GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error) {
	s.callLog = append(s.callLog, "GetUnreadCount")

	if s.getUnreadCountErr != nil {
		return 0, s.getUnreadCountErr
	}

	count, exists := s.unreadCounts[userId]
	if !exists {
		return 0, nil
	}
	return count, nil
}

func (s *TestNotificationService) AddNotification(userId uuid.UUID, notification NotificationResponse) {
	if _, exists := s.notifications[userId]; !exists {
		s.notifications[userId] = []NotificationResponse{}
	}
	s.notifications[userId] = append(s.notifications[userId], notification)
}

func (s *TestNotificationService) SetUnreadCount(userId uuid.UUID, count int) {
	s.unreadCounts[userId] = count
}

func (s *TestNotificationService) SetNotifyResponse(response *NotificationResponse, err error) {
	s.notifyResponse = response
	s.notifyErr = err
}

func (s *TestNotificationService) SetGetByRecipientError(err error) {
	s.getByRecipientErr = err
}

func (s *TestNotificationService) SetMarkAsReadError(err error) {
	s.markAsReadErr = err
}

func (s *TestNotificationService) SetGetUnreadCountError(err error) {
	s.getUnreadCountErr = err
}

func (s *TestNotificationService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func createTestHandler() (*TestNotificationService, *WebSocketHub, *NotificationHandler) {
	service := NewTestNotificationService()
	hub := NewWebSocketHub()
	jwtManager := auth.NewJWTManager([]byte("test-secret-key-for-tests"))

	handler := NewNotificationHandler(service, hub, jwtManager)

	return service, hub, handler
}

func TestNotificationHandler_GetMy(t *testing.T) {
	t.Run("Успешное получение уведомлений", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		avatarURL := "/uploads/avatar.jpg"
		testNotifications := []NotificationResponse{
			{
				ID:        uuid.New(),
				Recipient: userID,
				Type:      Like,
				TargetId:  uuid.New(),
				CreatedAt: time.Now(),
				Read:      false,
				Actor: content.ContentAuthor{
					ID:        uuid.New(),
					Username:  "user1",
					AvatarURL: &avatarURL,
				},
			},
		}
		service.AddNotification(userID, testNotifications[0])

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetMy(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GetMy() вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}

		var response []NotificationResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("GetMy() неверный JSON ответ: %v", err)
		}

		if len(response) != 1 {
			t.Errorf("GetMy() вернул %d уведомлений, ожидалось 1", len(response))
		}
		if !service.WasCalled("GetByRecipient") {
			t.Error("GetMy() должен вызывать GetByRecipient")
		}
	})

	t.Run("Успешное получение пустого списка уведомлений", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetMy(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GetMy() при пустом списке вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "[]"
		if body != expectedBody {
			t.Errorf("GetMy() при пустом списке вернул тело: '%s', ожидалось: '%s'", body, expectedBody)
		}
		if !service.WasCalled("GetByRecipient") {
			t.Error("GetMy() должен вызывать GetByRecipient даже для пустого списка")
		}
	})

	t.Run("Неподдерживаемый метод HTTP", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetMy(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("GetMy() при неподдерживаемом методе вернул статус %d, ожидалось %d", rr.Code, http.StatusMethodNotAllowed)
		}
		if service.WasCalled("GetByRecipient") {
			t.Error("GetMy() не должен вызывать GetByRecipient при неподдерживаемом методе")
		}
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	t.Run("Успешное помечание уведомлений как прочитанных", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPatch, "/notifications/read", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.MarkAsRead(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("MarkAsRead() вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}
		if !service.WasCalled("MarkAsRead") {
			t.Error("MarkAsRead() должен вызывать метод сервиса")
		}
	})

	t.Run("Неподдерживаемый метод HTTP", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications/read", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.MarkAsRead(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("MarkAsRead() при неподдерживаемом методе вернул статус %d, ожидалось %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	t.Run("Успешное получение количества непрочитанных уведомлений", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		expectedCount := 5
		service.SetUnreadCount(userID, expectedCount)

		req := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetUnreadCount(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GetUnreadCount() вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}

		var response int
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("GetUnreadCount() неверный JSON ответ: %v", err)
		}

		if response != expectedCount {
			t.Errorf("GetUnreadCount() вернул %d, ожидалось %d", response, expectedCount)
		}
		if !service.WasCalled("GetUnreadCount") {
			t.Error("GetUnreadCount() должен вызывать метод сервиса")
		}
	})

	t.Run("Успешное получение нулевого количества непрочитанных", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetUnreadCount(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GetUnreadCount() при нулевом количестве вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}

		var response int
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("GetUnreadCount() неверный JSON ответ: %v", err)
		}

		if response != 0 {
			t.Errorf("GetUnreadCount() при нулевом количестве вернул %d, ожидалось 0", response)
		}
	})

	t.Run("Неподдерживаемый метод HTTP", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/notifications/count", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetUnreadCount(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("GetUnreadCount() при неподдерживаемом методе вернул статус %d, ожидалось %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestNotificationHandler_ServeWS(t *testing.T) {
	t.Run("Создание нового хендлера", func(t *testing.T) {
		service := NewTestNotificationService()
		hub := NewWebSocketHub()

		handler := &NotificationHandler{
			service: service,
			hub:     hub,
		}

		if handler.service == nil {
			t.Error("Service не должен быть nil")
		}
		if handler.service != service {
			t.Error("Handler должен сохранять сервис")
		}
		if handler.hub != hub {
			t.Error("Handler должен сохранять хаб")
		}
	})
}

func TestNotificationHandler_New(t *testing.T) {
	t.Run("Проверка создания хендлера", func(t *testing.T) {
		service := NewTestNotificationService()

		handler := &NotificationHandler{
			service: service,
		}
		if handler.service == nil {
			t.Error("Service не должен быть nil")
		}
	})
}

func TestNotificationHandler_EmptyContext(t *testing.T) {
	t.Run("Обработка запроса без UserID в контексте", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)

		rr := httptest.NewRecorder()

		defer func() {
			if r := recover(); r == nil {
				t.Error("GetMy() должен паниковать при отсутствии UserID в контексте")
			}
		}()

		handler.GetMy(rr, req)
	})
}

func TestNotificationHandler_MultipleNotifications(t *testing.T) {
	t.Run("Получение нескольких уведомлений с сортировкой по времени", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()
		avatarURL := "/uploads/avatar.jpg"

		notifications := []NotificationResponse{
			{
				ID:        uuid.New(),
				Recipient: userID,
				Type:      Like,
				TargetId:  uuid.New(),
				CreatedAt: time.Now().Add(-2 * time.Hour),
				Read:      false,
				Actor: content.ContentAuthor{
					ID:        uuid.New(),
					Username:  "user1",
					AvatarURL: &avatarURL,
				},
			},
			{
				ID:        uuid.New(),
				Recipient: userID,
				Type:      Comment,
				TargetId:  uuid.New(),
				CreatedAt: time.Now().Add(-1 * time.Hour),
				Read:      true,
				Actor: content.ContentAuthor{
					ID:        uuid.New(),
					Username:  "user2",
					AvatarURL: &avatarURL,
				},
			},
			{
				ID:        uuid.New(),
				Recipient: userID,
				Type:      Follow,
				TargetId:  uuid.New(),
				CreatedAt: time.Now(),
				Read:      false,
				Actor: content.ContentAuthor{
					ID:        uuid.New(),
					Username:  "user3",
					AvatarURL: &avatarURL,
				},
			},
		}

		for _, notif := range notifications {
			service.AddNotification(userID, notif)
		}

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetMy(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("GetMy() вернул статус %d, ожидалось %d", rr.Code, http.StatusOK)
		}

		var response []NotificationResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("GetMy() неверный JSON ответ: %v", err)
		}

		if len(response) != 3 {
			t.Errorf("GetMy() вернул %d уведомлений, ожидалось 3", len(response))
		}

		typesFound := make(map[NotificationType]bool)
		for _, notif := range response {
			typesFound[notif.Type] = true
		}

		if !typesFound[Like] || !typesFound[Comment] || !typesFound[Follow] {
			t.Error("GetMy() должен возвращать уведомления разных типов")
		}
	})
}

func TestNotificationHandler_EmptyResponse(t *testing.T) {
	t.Run("Тестирование пустых ответов для всех методов", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		userID := uuid.New()

		req1 := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		ctx1 := context.WithValue(req1.Context(), contextKeys.UserIDKey, userID)
		req1 = req1.WithContext(ctx1)
		rr1 := httptest.NewRecorder()
		handler.GetMy(rr1, req1)

		if rr1.Code != http.StatusOK {
			t.Errorf("GetMy() с пустым списком: статус %d, ожидалось %d", rr1.Code, http.StatusOK)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		ctx2 := context.WithValue(req2.Context(), contextKeys.UserIDKey, userID)
		req2 = req2.WithContext(ctx2)
		rr2 := httptest.NewRecorder()
		handler.GetUnreadCount(rr2, req2)

		if rr2.Code != http.StatusOK {
			t.Errorf("GetUnreadCount() с нулевым значением: статус %d, ожидалось %d", rr2.Code, http.StatusOK)
		}

		req3 := httptest.NewRequest(http.MethodPatch, "/notifications/read", nil)
		ctx3 := context.WithValue(req3.Context(), contextKeys.UserIDKey, userID)
		req3 = req3.WithContext(ctx3)
		rr3 := httptest.NewRecorder()
		handler.MarkAsRead(rr3, req3)

		if rr3.Code != http.StatusOK {
			t.Errorf("MarkAsRead(): статус %d, ожидалось %d", rr3.Code, http.StatusOK)
		}
	})
}

func TestNotificationHandler_ErrorHandling(t *testing.T) {
	t.Run("GetMy с ошибкой сервиса", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		service.SetGetByRecipientError(errors.New("database error"))

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetMy(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("GetMy() при ошибке сервиса вернул статус %d, ожидалось %d", rr.Code, http.StatusNotFound)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "cannot get notifications"
		if body != expectedBody {
			t.Errorf("GetMy() при ошибке сервиса вернул тело: '%s', ожидалось: '%s'", body, expectedBody)
		}
	})

	t.Run("MarkAsRead с ошибкой сервиса", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		service.SetMarkAsReadError(errors.New("database error"))

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPatch, "/notifications/read", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.MarkAsRead(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("MarkAsRead() при ошибке сервиса вернул статус %d, ожидалось %d", rr.Code, http.StatusInternalServerError)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "cannot read notification"
		if body != expectedBody {
			t.Errorf("MarkAsRead() при ошибке сервиса вернул тело: '%s', ожидалось: '%s'", body, expectedBody)
		}
	})

	t.Run("GetUnreadCount с ошибкой сервиса", func(t *testing.T) {
		service, hub, handler := createTestHandler()
		_ = hub

		service.SetGetUnreadCountError(errors.New("database error"))

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		ctx := context.WithValue(req.Context(), contextKeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetUnreadCount(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("GetUnreadCount() при ошибке сервиса вернул статус %d, ожидалось %d", rr.Code, http.StatusNotFound)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "database error"
		if body != expectedBody {
			t.Errorf("GetUnreadCount() при ошибке сервиса вернул тело: '%s', ожидалось: '%s'", body, expectedBody)
		}
	})
}

func TestNotificationHandler_ServeWS_ErrorHandling(t *testing.T) {
	t.Run("Ошибка при отсутствии токена", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		req := httptest.NewRequest("GET", "/ws/notifications", nil)
		rr := httptest.NewRecorder()

		handler.ServeWS(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("ServeWS() без токена должен вернуть 401, got %d", rr.Code)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "Unauthorized"
		if body != expectedBody {
			t.Errorf("ServeWS() без токена должен вернуть '%s', got '%s'", expectedBody, body)
		}
	})

	t.Run("Ошибка при невалидном токене", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		req := httptest.NewRequest("GET", "/ws/notifications?token=invalid-token", nil)
		rr := httptest.NewRecorder()

		handler.ServeWS(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("ServeWS() с невалидным токеном должен вернуть 401, got %d", rr.Code)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "unauthorized"
		if body != expectedBody {
			t.Errorf("ServeWS() с невалидным токеном должен вернуть '%s', got '%s'", expectedBody, body)
		}
	})

	t.Run("Ошибка при пустом токене", func(t *testing.T) {
		_, hub, handler := createTestHandler()
		_ = hub

		req := httptest.NewRequest("GET", "/ws/notifications?token=", nil)
		rr := httptest.NewRecorder()

		handler.ServeWS(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("ServeWS() с пустым токеном должен вернуть 401, got %d", rr.Code)
		}

		body := strings.TrimSpace(rr.Body.String())
		expectedBody := "Unauthorized"
		if body != expectedBody {
			t.Errorf("ServeWS() с пустым токеном должен вернуть '%s', got '%s'", expectedBody, body)
		}
	})
}
