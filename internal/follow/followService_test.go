package follow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
)

// Мок для FollowRepository
type TestFollowRepository struct {
	createErr               error
	deleteErr               error
	existsResult            bool
	existsErr               error
	getFollowersResult      []*user.User
	getFollowersErr         error
	getFollowingResult      []*user.User
	getFollowingErr         error
	followersCount          int
	followersCountErr       error
	followingCount          int
	followingCountErr       error
	createCalled            bool
	deleteCalled            bool
	existsCalled            bool
	getFollowersCalled      bool
	getFollowingCalled      bool
	getFollowersCountCalled bool
	getFollowingCountCalled bool
}

func (m *TestFollowRepository) Create(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	m.createCalled = true
	return m.createErr
}

func (m *TestFollowRepository) Delete(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	m.deleteCalled = true
	return m.deleteErr
}

func (m *TestFollowRepository) Exists(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) (bool, error) {
	m.existsCalled = true
	return m.existsResult, m.existsErr
}

func (m *TestFollowRepository) GetFollowers(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	m.getFollowersCalled = true
	return m.getFollowersResult, m.getFollowersErr
}

func (m *TestFollowRepository) GetFollowing(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	m.getFollowingCalled = true
	return m.getFollowingResult, m.getFollowingErr
}

func (m *TestFollowRepository) GetFollowersCount(ctx context.Context, userId uuid.UUID) (int, error) {
	m.getFollowersCountCalled = true
	return m.followersCount, m.followersCountErr
}

func (m *TestFollowRepository) GetFollowingCount(ctx context.Context, userId uuid.UUID) (int, error) {
	m.getFollowingCountCalled = true
	return m.followingCount, m.followingCountErr
}

// Мок для NotificationService
type TestNotificationService struct {
	notifyCalled    bool
	notifyErr       error
	notifyResult    *notification.NotificationResponse
	lastFollowerID  uuid.UUID
	lastFollowingID uuid.UUID
	lastContentID   uuid.UUID
	lastType        notification.NotificationType
}

func (m *TestNotificationService) Notify(ctx context.Context, fromUserID, toUserID, contentID uuid.UUID, notificationType notification.NotificationType) (*notification.NotificationResponse, error) {
	m.notifyCalled = true
	m.lastFollowerID = fromUserID
	m.lastFollowingID = toUserID
	m.lastContentID = contentID
	m.lastType = notificationType

	if m.notifyErr != nil {
		return nil, m.notifyErr
	}

	return m.notifyResult, nil
}

func (m *TestNotificationService) GetByRecipient(ctx context.Context, recipientID uuid.UUID) ([]notification.NotificationResponse, error) {
	return []notification.NotificationResponse{}, nil
}

func (m *TestNotificationService) GetUnreadCount(ctx context.Context, recipientID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *TestNotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return nil
}

func (m *TestNotificationService) Delete(ctx context.Context, notificationID uuid.UUID) error {
	return nil
}

func createTestFollowService(repo *TestFollowRepository, notif *TestNotificationService) *followService {
	return &followService{
		repo:                repo,
		notificationService: notif,
	}
}

func createTestUser() *user.User {
	fullName := "Test User"
	description := "Test Description"
	avatarURL := "/uploads/avatar.jpg"

	return &user.User{
		ID:          uuid.New(),
		Username:    "testuser",
		FullName:    &fullName,
		Password:    "hashedpassword",
		AvatarURL:   &avatarURL,
		Description: &description,
		CreatedAt:   time.Now(),
	}
}

func TestFollowService_Create(t *testing.T) {
	t.Run("Успешное создание подписки", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsResult: false,
		}
		notif := &TestNotificationService{
			notifyResult: &notification.NotificationResponse{},
		}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Create(context.Background(), followerID, followingID)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if !repo.createCalled {
			t.Error("Create() должен вызывать репозиторий")
		}
		if !repo.existsCalled {
			t.Error("Create() должен проверять существование подписки")
		}
		if !notif.notifyCalled {
			t.Error("Create() должен отправлять уведомление")
		}
		if notif.lastType != notification.Follow {
			t.Errorf("Тип уведомления должен быть Follow, получен %v", notif.lastType)
		}
	})

	t.Run("Ошибка при попытке подписаться на себя", func(t *testing.T) {
		repo := &TestFollowRepository{}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		err := service.Create(context.Background(), userID, userID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при подписке на себя")
		}
		if err.Error() != "cannot follow yourself" {
			t.Errorf("Ожидалось сообщение 'cannot follow yourself', получено: %v", err)
		}
		if repo.createCalled {
			t.Error("Create() не должен вызывать репозиторий при подписке на себя")
		}
	})

	t.Run("Ошибка при проверке существования подписки", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Create(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке проверки существования")
		}
	})

	t.Run("Ошибка при уже существующей подписке", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsResult: true,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Create(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при уже существующей подписке")
		}
		if err.Error() != "already following" {
			t.Errorf("Ожидалось сообщение 'already following', получено: %v", err)
		}
	})

	t.Run("Ошибка при создании в репозитории", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsResult: false,
			createErr:    errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Create(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_Delete(t *testing.T) {
	t.Run("Успешное удаление подписки", func(t *testing.T) {
		repo := &TestFollowRepository{}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Delete(context.Background(), followerID, followingID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
		if !repo.deleteCalled {
			t.Error("Delete() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при удалении подписки", func(t *testing.T) {
		repo := &TestFollowRepository{
			deleteErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		err := service.Delete(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_Exists(t *testing.T) {
	t.Run("Успешная проверка существования подписки", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsResult: true,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		exists, err := service.Exists(context.Background(), followerID, followingID)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() должен вернуть true")
		}
		if !repo.existsCalled {
			t.Error("Exists() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при проверке существования подписки", func(t *testing.T) {
		repo := &TestFollowRepository{
			existsErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		followerID := uuid.New()
		followingID := uuid.New()

		_, err := service.Exists(context.Background(), followerID, followingID)
		if err == nil {
			t.Error("Exists() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_GetFollowers(t *testing.T) {
	t.Run("Успешное получение подписчиков", func(t *testing.T) {
		expectedUsers := []*user.User{
			createTestUser(),
			createTestUser(),
		}
		repo := &TestFollowRepository{
			getFollowersResult: expectedUsers,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		users, err := service.GetFollowers(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowers() error = %v", err)
		}
		if len(users) != len(expectedUsers) {
			t.Errorf("GetFollowers() вернул %d пользователей, ожидалось %d", len(users), len(expectedUsers))
		}
		if !repo.getFollowersCalled {
			t.Error("GetFollowers() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при получении подписчиков", func(t *testing.T) {
		repo := &TestFollowRepository{
			getFollowersErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		_, err := service.GetFollowers(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowers() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_GetFollowing(t *testing.T) {
	t.Run("Успешное получение подписок", func(t *testing.T) {
		expectedUsers := []*user.User{
			createTestUser(),
		}
		repo := &TestFollowRepository{
			getFollowingResult: expectedUsers,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		users, err := service.GetFollowing(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowing() error = %v", err)
		}
		if len(users) != len(expectedUsers) {
			t.Errorf("GetFollowing() вернул %d пользователей, ожидалось %d", len(users), len(expectedUsers))
		}
		if !repo.getFollowingCalled {
			t.Error("GetFollowing() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при получении подписок", func(t *testing.T) {
		repo := &TestFollowRepository{
			getFollowingErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		_, err := service.GetFollowing(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowing() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_GetFollowersCount(t *testing.T) {
	t.Run("Успешное получение количества подписчиков", func(t *testing.T) {
		expectedCount := 10
		repo := &TestFollowRepository{
			followersCount: expectedCount,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		count, err := service.GetFollowersCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowersCount() error = %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetFollowersCount() = %v, ожидалось %v", count, expectedCount)
		}
		if !repo.getFollowersCountCalled {
			t.Error("GetFollowersCount() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при получении количества подписчиков", func(t *testing.T) {
		repo := &TestFollowRepository{
			followersCountErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		_, err := service.GetFollowersCount(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowersCount() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestFollowService_GetFollowingCount(t *testing.T) {
	t.Run("Успешное получение количества подписок", func(t *testing.T) {
		expectedCount := 5
		repo := &TestFollowRepository{
			followingCount: expectedCount,
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		count, err := service.GetFollowingCount(context.Background(), userID)
		if err != nil {
			t.Errorf("GetFollowingCount() error = %v", err)
		}
		if count != expectedCount {
			t.Errorf("GetFollowingCount() = %v, ожидалось %v", count, expectedCount)
		}
		if !repo.getFollowingCountCalled {
			t.Error("GetFollowingCount() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при получении количества подписок", func(t *testing.T) {
		repo := &TestFollowRepository{
			followingCountErr: errors.New("database error"),
		}
		notif := &TestNotificationService{}
		service := createTestFollowService(repo, notif)

		userID := uuid.New()

		_, err := service.GetFollowingCount(context.Background(), userID)
		if err == nil {
			t.Error("GetFollowingCount() должен вернуть ошибку при ошибке репозитория")
		}
	})
}
