package like

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

// MockLikeService для тестов
type TestLikeService struct {
	createErr      error
	deleteErr      error
	existsResult   bool
	existsErr      error
	likesCount     int
	likesCountErr  error
	createCalled   bool
	deleteCalled   bool
	existsCalled   bool
	getCountCalled bool
}

func (m *TestLikeService) Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	m.createCalled = true
	return m.createErr
}

func (m *TestLikeService) Delete(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	m.deleteCalled = true
	return m.deleteErr
}

func (m *TestLikeService) Exists(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) (bool, error) {
	m.existsCalled = true
	return m.existsResult, m.existsErr
}

func (m *TestLikeService) GetLikesCount(ctx context.Context, contentId uuid.UUID) (int, error) {
	m.getCountCalled = true
	return m.likesCount, m.likesCountErr
}

func TestLikeHandler_Create(t *testing.T) {
	t.Run("Успешное создание лайка", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/api/like/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
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
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/like/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID поста", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/like/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "invalid content id" {
			t.Errorf("Ожидалось сообщение 'invalid content id', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestLikeService{
			createErr: errors.New("service error"),
		}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, "/api/like/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot like" {
			t.Errorf("Ожидалось сообщение 'cannot like', получено '%s'", body)
		}
	})
}

func TestLikeHandler_Delete(t *testing.T) {
	t.Run("Успешное удаление лайка", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/unlike/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
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
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/unlike/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID поста", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/unlike/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestLikeService{
			deleteErr: errors.New("service error"),
		}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/unlike/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot unlike" {
			t.Errorf("Ожидалось сообщение 'cannot unlike', получено '%s'", body)
		}
	})
}

func TestLikeHandler_Exists(t *testing.T) {
	t.Run("Успешная проверка существования", func(t *testing.T) {
		testService := &TestLikeService{
			existsResult: true,
		}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/liked/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Exists(w, req)

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
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/liked/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Exists(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID поста", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/liked/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Exists(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestLikeService{
			existsErr: errors.New("service error"),
		}
		handler := NewLikeHandler(testService)

		userID := uuid.New()
		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/liked/"+contentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Exists(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot check like" {
			t.Errorf("Ожидалось сообщение 'cannot check like', получено '%s'", body)
		}
	})
}

func TestLikeHandler_GetLikesCount(t *testing.T) {
	t.Run("Успешное получение количества лайков", func(t *testing.T) {
		testService := &TestLikeService{
			likesCount: 5,
		}
		handler := NewLikeHandler(testService)

		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/likes/"+contentID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetLikesCount(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !testService.getCountCalled {
			t.Error("Метод GetLikesCount сервиса не был вызван")
		}

		var result int
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if result != 5 {
			t.Errorf("Ожидалось 5, получено %d", result)
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		req := httptest.NewRequest(http.MethodPost, "/api/likes/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.GetLikesCount(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID поста", func(t *testing.T) {
		testService := &TestLikeService{}
		handler := NewLikeHandler(testService)

		req := httptest.NewRequest(http.MethodGet, "/api/likes/invalid-uuid", nil)
		w := httptest.NewRecorder()
		handler.GetLikesCount(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "invalid user id" {
			t.Errorf("Ожидалось сообщение 'invalid user id', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		testService := &TestLikeService{
			likesCountErr: errors.New("service error"),
		}
		handler := NewLikeHandler(testService)

		contentID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/likes/"+contentID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetLikesCount(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get followers count" {
			t.Errorf("Ожидалось сообщение 'cannot get followers count', получено '%s'", body)
		}
	})
}
