package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Мок для тестов NotificationRepository
type TestNotificationRepository struct {
	notifications        map[uuid.UUID]*Notification
	userNotifications    map[uuid.UUID][]NotificationResponse
	unreadCounts         map[uuid.UUID]int
	createdNotifications []*Notification
	markedAsReadUsers    []uuid.UUID
	errors               map[string]error
	callLog              []string
}

func NewTestNotificationRepository() *TestNotificationRepository {
	return &TestNotificationRepository{
		notifications:        make(map[uuid.UUID]*Notification),
		userNotifications:    make(map[uuid.UUID][]NotificationResponse),
		unreadCounts:         make(map[uuid.UUID]int),
		createdNotifications: []*Notification{},
		markedAsReadUsers:    []uuid.UUID{},
		errors:               make(map[string]error),
	}
}

func (r *TestNotificationRepository) Create(ctx context.Context, n *Notification) (*NotificationResponse, error) {
	r.callLog = append(r.callLog, "Create")
	if err, ok := r.errors["Create"]; ok {
		return nil, err
	}

	r.notifications[n.ID] = n
	r.createdNotifications = append(r.createdNotifications, n)

	r.unreadCounts[n.Recipient] = r.unreadCounts[n.Recipient] + 1

	avatarURL := "/uploads/avatar.jpg"
	response := &NotificationResponse{
		ID:        n.ID,
		Recipient: n.Recipient,
		Type:      n.Type,
		TargetId:  n.TargetId,
		CreatedAt: n.CreatedAt,
		Read:      n.Read,
		Actor: content.ContentAuthor{
			ID:        n.Actor,
			Username:  "testuser",
			AvatarURL: &avatarURL,
		},
	}

	return response, nil
}

func (r *TestNotificationRepository) GetByRecipient(ctx context.Context, userId uuid.UUID) ([]NotificationResponse, error) {
	r.callLog = append(r.callLog, "GetByRecipient")
	if err, ok := r.errors["GetByRecipient"]; ok {
		return nil, err
	}

	notifications, exists := r.userNotifications[userId]
	if !exists {
		return []NotificationResponse{}, nil
	}

	return notifications, nil
}

func (r *TestNotificationRepository) MarkAsRead(ctx context.Context, userId uuid.UUID) error {
	r.callLog = append(r.callLog, "MarkAsRead")
	if err, ok := r.errors["MarkAsRead"]; ok {
		return err
	}

	r.markedAsReadUsers = append(r.markedAsReadUsers, userId)

	r.unreadCounts[userId] = 0

	if notifs, exists := r.userNotifications[userId]; exists {
		for i := range notifs {
			notifs[i].Read = true
		}
	}

	return nil
}

func (r *TestNotificationRepository) GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error) {
	r.callLog = append(r.callLog, "GetUnreadCount")
	if err, ok := r.errors["GetUnreadCount"]; ok {
		return 0, err
	}

	count, exists := r.unreadCounts[userId]
	if !exists {
		return 0, nil
	}

	return count, nil
}

func (r *TestNotificationRepository) AddUserNotification(userId uuid.UUID, notification NotificationResponse) {
	if _, exists := r.userNotifications[userId]; !exists {
		r.userNotifications[userId] = []NotificationResponse{}
	}
	r.userNotifications[userId] = append(r.userNotifications[userId], notification)
}

func (r *TestNotificationRepository) SetUnreadCount(userId uuid.UUID, count int) {
	r.unreadCounts[userId] = count
}

func (r *TestNotificationRepository) SetError(methodName string, err error) {
	r.errors[methodName] = err
}

func (r *TestNotificationRepository) Reset() {
	r.notifications = make(map[uuid.UUID]*Notification)
	r.userNotifications = make(map[uuid.UUID][]NotificationResponse)
	r.unreadCounts = make(map[uuid.UUID]int)
	r.createdNotifications = []*Notification{}
	r.markedAsReadUsers = []uuid.UUID{}
	r.errors = make(map[string]error)
	r.callLog = []string{}
}

func (r *TestNotificationRepository) WasCalled(methodName string) bool {
	for _, call := range r.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func (r *TestNotificationRepository) GetCallCount(methodName string) int {
	count := 0
	for _, call := range r.callLog {
		if call == methodName {
			count++
		}
	}
	return count
}

func createTestNotificationService(repo *TestNotificationRepository) *notificationService {
	hub := NewWebSocketHub()
	return &notificationService{
		repo: repo,
		hub:  hub,
	}
}

func createTestNotificationResponse() NotificationResponse {
	avatarURL := "/uploads/avatar.jpg"
	return NotificationResponse{
		ID:        uuid.New(),
		Recipient: uuid.New(),
		Type:      Comment,
		TargetId:  uuid.New(),
		CreatedAt: time.Now(),
		Read:      false,
		Actor: content.ContentAuthor{
			ID:        uuid.New(),
			Username:  "testactor",
			AvatarURL: &avatarURL,
		},
	}
}

func TestNotificationService_Notify_EdgeCases(t *testing.T) {
	t.Run("Уведомление самому себе должно быть проигнорировано", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		actorID := uuid.New()
		response, err := service.Notify(context.Background(), actorID, actorID, uuid.New(), Like)

		if err != nil {
			t.Errorf("Notify() при уведомлении себя не должен возвращать ошибку: %v", err)
		}
		if response != nil {
			t.Error("Notify() при уведомлении себя должен вернуть nil response")
		}
		if repo.WasCalled("Create") {
			t.Error("Notify() при уведомлении себя не должен вызывать Create")
		}
	})

	t.Run("Успешное создание уведомления", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		actorID := uuid.New()
		recipientID := uuid.New()
		targetID := uuid.New()

		response, err := service.Notify(context.Background(), actorID, recipientID, targetID, Follow)

		if err != nil {
			t.Errorf("Notify() ошибка: %v", err)
		}
		if response == nil {
			t.Error("Notify() должен вернуть response")
		}
		if !repo.WasCalled("Create") {
			t.Error("Notify() должен вызывать Create")
		}
		if response.Type != Follow {
			t.Errorf("Notify() вернул неверный тип: got %v, want %v", response.Type, Follow)
		}
	})

	t.Run("Ошибка при создании в репозитории", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("Create", errors.New("database error"))
		service := createTestNotificationService(repo)

		actorID := uuid.New()
		recipientID := uuid.New()

		response, err := service.Notify(context.Background(), actorID, recipientID, uuid.New(), Comment)

		if err == nil {
			t.Error("Notify() должен вернуть ошибку при ошибке создания")
		}
		if response != nil {
			t.Error("Notify() должен вернуть nil response при ошибке")
		}
	})

	t.Run("Все типы уведомлений работают корректно", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		actorID := uuid.New()
		recipientID := uuid.New()
		targetID := uuid.New()

		notificationTypes := []NotificationType{Like, Comment, Follow, Response}

		for _, notifType := range notificationTypes {
			repo.Reset()

			response, err := service.Notify(context.Background(), actorID, recipientID, targetID, notifType)

			if err != nil {
				t.Errorf("Notify() с типом %v ошибка: %v", notifType, err)
			}
			if response.Type != notifType {
				t.Errorf("Notify() с типом %v вернул неверный тип: got %v", notifType, response.Type)
			}
		}
	})
}

func TestNotificationService_GetByRecipient_EdgeCases(t *testing.T) {
	t.Run("Успешное получение уведомлений пользователя", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		testNotifications := []NotificationResponse{
			createTestNotificationResponse(),
			createTestNotificationResponse(),
		}

		for _, notif := range testNotifications {
			notifCopy := notif
			notifCopy.Recipient = userID
			repo.AddUserNotification(userID, notifCopy)
		}

		notifications, err := service.GetByRecipient(context.Background(), userID)

		if err != nil {
			t.Errorf("GetByRecipient() ошибка: %v", err)
		}
		if len(notifications) != 2 {
			t.Errorf("GetByRecipient() вернул %d уведомлений, ожидалось 2", len(notifications))
		}
		if !repo.WasCalled("GetByRecipient") {
			t.Error("GetByRecipient() должен вызывать репозиторий")
		}
	})

	t.Run("Получение уведомлений для пользователя без уведомлений", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		notifications, err := service.GetByRecipient(context.Background(), userID)

		if err != nil {
			t.Errorf("GetByRecipient() для пустого списка ошибка: %v", err)
		}
		if len(notifications) != 0 {
			t.Errorf("GetByRecipient() для пустого списка вернул %d уведомлений, ожидалось 0", len(notifications))
		}
	})

	t.Run("Ошибка при получении из репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("GetByRecipient", errors.New("database error"))
		service := createTestNotificationService(repo)

		userID := uuid.New()
		_, err := service.GetByRecipient(context.Background(), userID)

		if err == nil {
			t.Error("GetByRecipient() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestNotificationService_MarkAsRead_EdgeCases(t *testing.T) {
	t.Run("Успешное помечание всех уведомлений как прочитанных", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		repo.SetUnreadCount(userID, 5)

		err := service.MarkAsRead(context.Background(), userID)

		if err != nil {
			t.Errorf("MarkAsRead() ошибка: %v", err)
		}
		if !repo.WasCalled("MarkAsRead") {
			t.Error("MarkAsRead() должен вызывать репозиторий")
		}

		found := false
		for _, markedUser := range repo.markedAsReadUsers {
			if markedUser == userID {
				found = true
				break
			}
		}
		if !found {
			t.Error("MarkAsRead() должен добавить пользователя в список помеченных как прочитанные")
		}
	})

	t.Run("Ошибка при помечании как прочитанных в репозитории", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("MarkAsRead", errors.New("database error"))
		service := createTestNotificationService(repo)

		userID := uuid.New()
		err := service.MarkAsRead(context.Background(), userID)

		if err == nil {
			t.Error("MarkAsRead() должен вернуть ошибку при ошибке репозитория")
		}
	})

	t.Run("Помечание как прочитанных для пользователя без уведомлений", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		err := service.MarkAsRead(context.Background(), userID)

		if err != nil {
			t.Errorf("MarkAsRead() для пользователя без уведомлений ошибка: %v", err)
		}
		if !repo.WasCalled("MarkAsRead") {
			t.Error("MarkAsRead() должен вызывать репозиторий даже для пользователя без уведомлений")
		}
	})
}

func TestNotificationService_GetUnreadCount_EdgeCases(t *testing.T) {
	t.Run("Успешное получение количества непрочитанных уведомлений", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		expectedCount := 3
		repo.SetUnreadCount(userID, expectedCount)

		count, err := service.GetUnreadCount(context.Background(), userID)

		if err != nil {
			t.Errorf("GetUnreadCount() ошибка: %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetUnreadCount() вернул %d, ожидалось %d", count, expectedCount)
		}
		if !repo.WasCalled("GetUnreadCount") {
			t.Error("GetUnreadCount() должен вызывать репозиторий")
		}
	})

	t.Run("Получение количества для пользователя без непрочитанных", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()
		count, err := service.GetUnreadCount(context.Background(), userID)

		if err != nil {
			t.Errorf("GetUnreadCount() для нулевого количества ошибка: %v", err)
		}
		if count != 0 {
			t.Errorf("GetUnreadCount() для нулевого количества вернул %d, ожидалось 0", count)
		}
	})

	t.Run("Ошибка при получении количества из репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("GetUnreadCount", errors.New("database error"))
		service := createTestNotificationService(repo)

		userID := uuid.New()
		_, err := service.GetUnreadCount(context.Background(), userID)

		if err == nil {
			t.Error("GetUnreadCount() должен вернуть ошибку при ошибке репозитория")
		}
	})

	t.Run("Обновление счетчика после создания уведомления", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		service := createTestNotificationService(repo)

		userID := uuid.New()

		initialCount, _ := service.GetUnreadCount(context.Background(), userID)
		if initialCount != 0 {
			t.Errorf("Начальный счетчик должен быть 0, got %d", initialCount)
		}

		_, err := service.Notify(context.Background(), uuid.New(), userID, uuid.New(), Like)
		if err != nil {
			t.Errorf("Notify() ошибка: %v", err)
		}

		updatedCount, err := service.GetUnreadCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetUnreadCount() после создания ошибка: %v", err)
		}
		if updatedCount != 1 {
			t.Errorf("Счетчик после создания должен быть 1, got %d", updatedCount)
		}

		err = service.MarkAsRead(context.Background(), userID)
		if err != nil {
			t.Errorf("MarkAsRead() ошибка: %v", err)
		}

		finalCount, err := service.GetUnreadCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetUnreadCount() после MarkAsRead ошибка: %v", err)
		}
		if finalCount != 0 {
			t.Errorf("Счетчик после MarkAsRead должен быть 0, got %d", finalCount)
		}
	})
}

// Тесты для WebSocketHub
func TestWebSocketHub_RegisterUnregister(t *testing.T) {
	hub := NewWebSocketHub()
	userID := uuid.New()
	conn := &websocket.Conn{}

	hub.Register(userID, conn)
}

func TestWebSocketHub_SendToUser(t *testing.T) {
	hub := NewWebSocketHub()
	userID := uuid.New()

	avatarURL := "/uploads/avatar.jpg"
	nr := &NotificationResponse{
		ID:        uuid.New(),
		Recipient: userID,
		Type:      Like,
		TargetId:  uuid.New(),
		CreatedAt: time.Now(),
		Read:      false,
		Actor: content.ContentAuthor{
			ID:        uuid.New(),
			Username:  "test",
			AvatarURL: &avatarURL,
		},
	}

	hub.SendToUser(userID, nr)
}

func TestNotificationService_ServiceErrors(t *testing.T) {
	t.Run("Notify с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("Create", errors.New("database error"))
		hub := NewWebSocketHub()
		service := NewNotificationService(repo, hub)

		response, err := service.Notify(context.Background(), uuid.New(), uuid.New(), uuid.New(), Like)

		if err == nil {
			t.Error("Notify() должен вернуть ошибку при ошибке репозитория")
		}
		if response != nil {
			t.Error("Notify() должен вернуть nil response при ошибке")
		}
	})

	t.Run("GetByRecipient с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("GetByRecipient", errors.New("database error"))
		hub := NewWebSocketHub()
		service := NewNotificationService(repo, hub)

		_, err := service.GetByRecipient(context.Background(), uuid.New())

		if err == nil {
			t.Error("GetByRecipient() должен вернуть ошибку при ошибке репозитория")
		}
	})

	t.Run("GetUnreadCount с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("GetUnreadCount", errors.New("database error"))
		hub := NewWebSocketHub()
		service := NewNotificationService(repo, hub)

		_, err := service.GetUnreadCount(context.Background(), uuid.New())

		if err == nil {
			t.Error("GetUnreadCount() должен вернуть ошибку при ошибке репозитория")
		}
	})

	t.Run("MarkAsRead с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestNotificationRepository()
		repo.SetError("MarkAsRead", errors.New("database error"))
		hub := NewWebSocketHub()
		service := NewNotificationService(repo, hub)

		err := service.MarkAsRead(context.Background(), uuid.New())

		if err == nil {
			t.Error("MarkAsRead() должен вернуть ошибку при ошибке репозитория")
		}
	})
}
