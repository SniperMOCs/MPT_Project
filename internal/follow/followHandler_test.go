package follow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"time"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
)

// MockFollowService для тестов
type TestFollowService struct {
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

func (m *TestFollowService) Create(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	m.createCalled = true
	return m.createErr
}

func (m *TestFollowService) Delete(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	m.deleteCalled = true
	return m.deleteErr
}

func (m *TestFollowService) Exists(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) (bool, error) {
	m.existsCalled = true
	return m.existsResult, m.existsErr
}

func (m *TestFollowService) GetFollowers(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	m.getFollowersCalled = true
	return m.getFollowersResult, m.getFollowersErr
}

func (m *TestFollowService) GetFollowing(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	m.getFollowingCalled = true
	return m.getFollowingResult, m.getFollowingErr
}

func (m *TestFollowService) GetFollowersCount(ctx context.Context, userId uuid.UUID) (int, error) {
	m.getFollowersCountCalled = true
	return m.followersCount, m.followersCountErr
}

func (m *TestFollowService) GetFollowingCount(ctx context.Context, userId uuid.UUID) (int, error) {
	m.getFollowingCountCalled = true
	return m.followingCount, m.followingCountErr
}

func createTestUserDTO() *user.User {
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

func TestFollowHandler_Create(t *testing.T) {
	t.Run("Успешное создание подписки", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		followingID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/api/follow/"+followingID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, w.Code)
		}
		if !testService.createCalled {
			t.Error("Метод Create сервиса не был вызван")
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/follow/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/follow/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "invalid user id" {
			t.Errorf("Ожидалось сообщение 'invalid user id', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestFollowService{
			createErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		followingID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/api/follow/"+followingID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot follow" {
			t.Errorf("Ожидалось сообщение 'cannot follow', получено '%s'", body)
		}
	})
}

func TestFollowHandler_Delete(t *testing.T) {
	t.Run("Успешное удаление подписки", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		followingID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/unfollow/"+followingID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.deleteCalled {
			t.Error("Метод Delete сервиса не был вызван")
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/unfollow/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/unfollow/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestFollowService{
			deleteErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		followingID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/unfollow/"+followingID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot unfollow" {
			t.Errorf("Ожидалось сообщение 'cannot unfollow', получено '%s'", body)
		}
	})
}

func TestFollowHandler_GetFollowers(t *testing.T) {
	t.Run("Успешное получение подписчиков", func(t *testing.T) {
		expectedUsers := []*user.User{
			createTestUserDTO(),
			createTestUserDTO(),
		}
		testService := &TestFollowService{
			getFollowersResult: expectedUsers,
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/followers/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.getFollowersCalled {
			t.Error("Метод GetFollowers сервиса не был вызван")
		}

		var result []*user.User
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if len(result) != len(expectedUsers) {
			t.Errorf("Ожидалось %d пользователей, получено %d", len(expectedUsers), len(result))
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/followers/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowers(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/followers/invalid-uuid", nil)
		w := httptest.NewRecorder()
		handler.GetFollowers(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestFollowService{
			getFollowersErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/followers/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowers(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get followers" {
			t.Errorf("Ожидалось сообщение 'cannot get followers', получено '%s'", body)
		}
	})
}

func TestFollowHandler_GetFollowing(t *testing.T) {
	t.Run("Успешное получение подписок", func(t *testing.T) {
		expectedUsers := []*user.User{
			createTestUserDTO(),
		}
		testService := &TestFollowService{
			getFollowingResult: expectedUsers,
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/following/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowing(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.getFollowingCalled {
			t.Error("Метод GetFollowing сервиса не был вызван")
		}

		var result []*user.User
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if len(result) != len(expectedUsers) {
			t.Errorf("Ожидалось %d пользователей, получено %d", len(expectedUsers), len(result))
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/following/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowing(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/following/invalid-uuid", nil)
		w := httptest.NewRecorder()
		handler.GetFollowing(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestFollowService{
			getFollowingErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/following/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetFollowing(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get following" {
			t.Errorf("Ожидалось сообщение 'cannot get following', получено '%s'", body)
		}
	})
}

func TestFollowHandler_FollowInfo(t *testing.T) {
	t.Run("Успешное получение информации о подписках", func(t *testing.T) {
		testService := &TestFollowService{
			followersCount: 100,
			followingCount: 50,
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/follow/info/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.FollowInfo(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.getFollowersCountCalled {
			t.Error("Метод GetFollowersCount сервиса не был вызван")
		}
		if !testService.getFollowingCountCalled {
			t.Error("Метод GetFollowingCount сервиса не был вызван")
		}

		var result FollowInfoDTO
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if result.FollowersCount != 100 {
			t.Errorf("Ожидалось 100 подписчиков, получено %d", result.FollowersCount)
		}
		if result.FollowingCount != 50 {
			t.Errorf("Ожидалось 50 подписок, получено %d", result.FollowingCount)
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/follow/info/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.FollowInfo(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/follow/info/invalid-uuid", nil)
		w := httptest.NewRecorder()
		handler.FollowInfo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса при получении количества подписчиков", func(t *testing.T) {
		testService := &TestFollowService{
			followersCountErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/follow/info/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.FollowInfo(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get followers count" {
			t.Errorf("Ожидалось сообщение 'cannot get followers count', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса при получении количества подписок", func(t *testing.T) {
		testService := &TestFollowService{
			followersCount:    100,
			followingCountErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/follow/info/"+userID.String(), nil)
		w := httptest.NewRecorder()
		handler.FollowInfo(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get following count" {
			t.Errorf("Ожидалось сообщение 'cannot get following count', получено '%s'", body)
		}
	})
}

func TestFollowHandler_IsFollower(t *testing.T) {
	t.Run("Успешная проверка подписки", func(t *testing.T) {
		testService := &TestFollowService{
			existsResult: true,
		}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/follower/"+userID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.IsFollower(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.existsCalled {
			t.Error("Метод Exists сервиса не был вызван")
		}

		var result bool
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if !result {
			t.Error("Ожидалось true, получено false")
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/follower/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.IsFollower(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID пользователя", func(t *testing.T) {
		testService := &TestFollowService{}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/follower/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.IsFollower(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestFollowService{
			existsErr: errors.New("service error"),
		}
		handler := NewFollowHandler(testService)

		myID := uuid.New()
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/follower/"+userID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.IsFollower(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot check follow" {
			t.Errorf("Ожидалось сообщение 'cannot check follow', получено '%s'", body)
		}
	})
}
