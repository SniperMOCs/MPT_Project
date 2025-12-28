package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestNotificationRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок БД: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	t.Run("Успешное создание уведомления", func(t *testing.T) {
		notificationID := uuid.New()
		recipientID := uuid.New()
		actorID := uuid.New()
		targetID := uuid.New()
		createdAt := time.Now()

		n := &Notification{
			ID:        notificationID,
			Recipient: recipientID,
			Actor:     actorID,
			Type:      Like,
			TargetId:  targetID,
			CreatedAt: createdAt,
			Read:      false,
		}

		rows := sqlmock.NewRows([]string{
			"id", "recipient_id", "target_id", "type", "created_at", "read",
			"id", "username", "avatar_url",
		}).AddRow(
			notificationID, recipientID, targetID, Like, createdAt, false,
			actorID, "testuser", "/uploads/avatar.jpg",
		)

		mock.ExpectQuery(`WITH inserted AS`).
			WithArgs(notificationID, recipientID, targetID, Like, createdAt, actorID, false).
			WillReturnRows(rows)

		response, err := repo.Create(context.Background(), n)

		if err != nil {
			t.Errorf("Create() ошибка: %v", err)
		}
		if response == nil {
			t.Error("Create() должен вернуть response")
		}
		if response.ID != notificationID {
			t.Errorf("Create() вернул неверный ID: got %v, want %v", response.ID, notificationID)
		}
		if response.Actor.Username != "testuser" {
			t.Errorf("Create() вернул неверное имя пользователя: got %v, want %v", response.Actor.Username, "testuser")
		}
	})

	t.Run("Ошибка при создании уведомления", func(t *testing.T) {
		n := &Notification{
			ID:        uuid.New(),
			Recipient: uuid.New(),
			Actor:     uuid.New(),
			Type:      Comment,
			TargetId:  uuid.New(),
			CreatedAt: time.Now(),
			Read:      false,
		}

		mock.ExpectQuery(`WITH inserted AS`).
			WillReturnError(errors.New("database error"))

		response, err := repo.Create(context.Background(), n)

		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке БД")
		}
		if response != nil {
			t.Error("Create() должен вернуть nil response при ошибке")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не все ожидания были выполнены: %v", err)
	}
}

func TestNotificationRepository_GetByRecipient(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок БД: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	t.Run("Успешное получение уведомлений пользователя", func(t *testing.T) {
		userID := uuid.New()
		notificationID1 := uuid.New()
		notificationID2 := uuid.New()
		actorID := uuid.New()
		targetID := uuid.New()
		createdAt := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "recipient_id", "target_id", "type", "created_at", "read", "actor_id", "username", "avatar_url",
		}).
			AddRow(notificationID1, userID, targetID, Like, createdAt, false, actorID, "user1", "/uploads/avatar1.jpg").
			AddRow(notificationID2, userID, targetID, Comment, createdAt.Add(-time.Hour), true, actorID, "user2", "/uploads/avatar2.jpg")

		mock.ExpectQuery(`SELECT n.id, n.recipient_id, n.target_id, n.type, n.created_at, n.read, n.actor_id, u.username, u.avatar_url`).
			WithArgs(userID).
			WillReturnRows(rows)

		notifications, err := repo.GetByRecipient(context.Background(), userID)

		if err != nil {
			t.Errorf("GetByRecipient() ошибка: %v", err)
		}
		if len(notifications) != 2 {
			t.Errorf("GetByRecipient() вернул %d уведомлений, ожидалось 2", len(notifications))
		}
		if notifications[0].Type != Like {
			t.Errorf("GetByRecipient() вернул неверный тип первого уведомления: got %v, want %v", notifications[0].Type, Like)
		}
		if !notifications[1].Read {
			t.Error("GetByRecipient() второе уведомление должно быть прочитано")
		}
	})

	t.Run("Получение уведомлений для пользователя без уведомлений", func(t *testing.T) {
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "recipient_id", "target_id", "type", "created_at", "read", "actor_id", "username", "avatar_url",
		})

		mock.ExpectQuery(`SELECT n.id, n.recipient_id, n.target_id, n.type, n.created_at, n.read, n.actor_id, u.username, u.avatar_url`).
			WithArgs(userID).
			WillReturnRows(rows)

		notifications, err := repo.GetByRecipient(context.Background(), userID)

		if err != nil {
			t.Errorf("GetByRecipient() для пустого списка ошибка: %v", err)
		}
		if len(notifications) != 0 {
			t.Errorf("GetByRecipient() для пустого списка вернул %d уведомлений, ожидалось 0", len(notifications))
		}
	})

	t.Run("Ошибка при получении уведомлений", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT n.id, n.recipient_id, n.target_id, n.type, n.created_at, n.read, n.actor_id, u.username, u.avatar_url`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetByRecipient(context.Background(), userID)

		if err == nil {
			t.Error("GetByRecipient() должен вернуть ошибку при ошибке БД")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не все ожидания были выполнены: %v", err)
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок БД: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	t.Run("Успешное помечание уведомлений как прочитанных", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec(`UPDATE notifications`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 3))

		err := repo.MarkAsRead(context.Background(), userID)

		if err != nil {
			t.Errorf("MarkAsRead() ошибка: %v", err)
		}
	})

	t.Run("Ошибка при помечании уведомлений как прочитанных", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectExec(`UPDATE notifications`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		err := repo.MarkAsRead(context.Background(), userID)

		if err == nil {
			t.Error("MarkAsRead() должен вернуть ошибку при ошибке БД")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не все ожидания были выполнены: %v", err)
	}
}

func TestNotificationRepository_GetUnreadCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать мок БД: %v", err)
	}
	defer db.Close()

	repo := NewNotificationRepository(db)

	t.Run("Успешное получение количества непрочитанных уведомлений", func(t *testing.T) {
		userID := uuid.New()
		expectedCount := 5

		rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications WHERE recipient_id = \$1 AND read = false`).
			WithArgs(userID).
			WillReturnRows(rows)

		count, err := repo.GetUnreadCount(context.Background(), userID)

		if err != nil {
			t.Errorf("GetUnreadCount() ошибка: %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetUnreadCount() вернул %d, ожидалось %d", count, expectedCount)
		}
	})

	t.Run("Получение нулевого количества непрочитанных", func(t *testing.T) {
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications WHERE recipient_id = \$1 AND read = false`).
			WithArgs(userID).
			WillReturnRows(rows)

		count, err := repo.GetUnreadCount(context.Background(), userID)

		if err != nil {
			t.Errorf("GetUnreadCount() для нулевого количества ошибка: %v", err)
		}
		if count != 0 {
			t.Errorf("GetUnreadCount() для нулевого количества вернул %d, ожидалось 0", count)
		}
	})

	t.Run("Ошибка при получении количества", func(t *testing.T) {
		userID := uuid.New()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM notifications WHERE recipient_id = \$1 AND read = false`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetUnreadCount(context.Background(), userID)

		if err == nil {
			t.Error("GetUnreadCount() должен вернуть ошибку при ошибке БД")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не все ожидания были выполнены: %v", err)
	}
}

func TestWebSocketHub_Unregister(t *testing.T) {
	t.Run("Unregister несуществующего пользователя", func(t *testing.T) {
		hub := NewWebSocketHub()
		userID := uuid.New()

		hub.Unregister(userID)

		hub.Unregister(userID)
		hub.Unregister(userID)
	})

	t.Run("Unregister после Register", func(t *testing.T) {
		hub := NewWebSocketHub()
		userID := uuid.New()

		hub.Register(userID, nil)

		hub.Unregister(userID)

		hub.Unregister(userID)
	})
}
