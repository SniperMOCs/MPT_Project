package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

// Мок для CommentService
type TestCommentService struct {
	createErr            error
	deleteErr            error
	getByContentIdErr    error
	createCalled         bool
	deleteCalled         bool
	getByContentIdCalled bool
	createArgs           CreateArgs
	deleteArgs           DeleteArgs
	comments             []CommentResponse
}

type CreateArgs struct {
	authorID  uuid.UUID
	parentID  *uuid.UUID
	contentID uuid.UUID
	text      string
}

type DeleteArgs struct {
	commentID uuid.UUID
	authorID  uuid.UUID
}

func (m *TestCommentService) Create(ctx context.Context, authorId uuid.UUID, parentId *uuid.UUID, contentId uuid.UUID, text string) error {
	m.createCalled = true
	m.createArgs = CreateArgs{
		authorID:  authorId,
		parentID:  parentId,
		contentID: contentId,
		text:      text,
	}
	return m.createErr
}

func (m *TestCommentService) Delete(ctx context.Context, id uuid.UUID, authorId uuid.UUID) error {
	m.deleteCalled = true
	m.deleteArgs = DeleteArgs{
		commentID: id,
		authorID:  authorId,
	}
	return m.deleteErr
}

func (m *TestCommentService) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]CommentResponse, error) {
	m.getByContentIdCalled = true
	return m.comments, m.getByContentIdErr
}

func TestCommentHandler_Create(t *testing.T) {
	t.Run("Успешное создание комментария", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		contentID := uuid.New()
		dto := CommentDTO{
			ContentId: contentID,
			Text:      "Great post!",
		}

		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/comment", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, w.Code)
		}
		if !service.createCalled {
			t.Error("Метод Create сервиса не был вызван")
		}
		if service.createArgs.authorID != myID {
			t.Error("Неверный ID автора передан в сервис")
		}
		if service.createArgs.contentID != contentID {
			t.Error("Неверный ID поста передан в сервис")
		}
		if service.createArgs.text != "Great post!" {
			t.Error("Неверный текст передан в сервис")
		}
	})

	t.Run("Успешное создание ответа на комментарий", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		parentID := uuid.New()
		contentID := uuid.New()
		dto := CommentDTO{
			ParentId:  &parentID,
			ContentId: contentID,
			Text:      "Thanks for the comment!",
		}

		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/comment", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, w.Code)
		}
		if !service.createCalled {
			t.Error("Метод Create сервиса не был вызван")
		}
		if *service.createArgs.parentID != parentID {
			t.Error("Неверный ID родительского комментария передан в сервис")
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/comment", nil)
		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный формат JSON", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/comment", strings.NewReader("invalid json"))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		service := &TestCommentService{
			createErr: errors.New("service error"),
		}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		dto := CommentDTO{
			ContentId: uuid.New(),
			Text:      "Test comment",
		}

		body, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/comment", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "service error" {
			t.Errorf("Ожидалось сообщение 'service error', получено '%s'", body)
		}
	})
}

func TestCommentHandler_Delete(t *testing.T) {
	t.Run("Успешное удаление комментария", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		commentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/comment/delete/"+commentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !service.deleteCalled {
			t.Error("Метод Delete сервиса не был вызван")
		}
		if service.deleteArgs.commentID != commentID {
			t.Error("Неверный ID комментария передан в сервис")
		}
		if service.deleteArgs.authorID != myID {
			t.Error("Неверный ID автора передан в сервис")
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		req := httptest.NewRequest(http.MethodPost, "/api/comment/delete/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID комментария", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/comment/delete/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "invalid comment id" {
			t.Errorf("Ожидалось сообщение 'invalid comment id', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		service := &TestCommentService{
			deleteErr: errors.New("cannot delete"),
		}
		handler := NewCommentHandler(service)

		myID := uuid.New()
		commentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/api/comment/delete/"+commentID.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, myID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot delete comment" {
			t.Errorf("Ожидалось сообщение 'cannot delete comment', получено '%s'", body)
		}
	})
}

func TestCommentHandler_GetByContentId(t *testing.T) {
	t.Run("Успешное получение комментариев к посту", func(t *testing.T) {
		contentID := uuid.New()
		expectedComments := []CommentResponse{
			{
				ID:        uuid.New(),
				ParentId:  nil,
				ContentId: contentID,
				Text:      "First comment",
				CreatedAt: time.Now(),
			},
		}

		service := &TestCommentService{
			comments: expectedComments,
		}
		handler := NewCommentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/comment/content/"+contentID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetByContentId(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
		}
		if !service.getByContentIdCalled {
			t.Error("Метод GetByContentId сервиса не был вызван")
		}

		var result []CommentResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Errorf("Ошибка декодирования JSON: %v", err)
		}
		if len(result) != len(expectedComments) {
			t.Errorf("Ожидалось %d комментариев, получено %d", len(expectedComments), len(result))
		}
	})

	t.Run("Неверный метод", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		req := httptest.NewRequest(http.MethodPost, "/api/comment/content/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handler.GetByContentId(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Неверный ID поста", func(t *testing.T) {
		service := &TestCommentService{}
		handler := NewCommentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/comment/content/invalid-uuid", nil)
		w := httptest.NewRecorder()
		handler.GetByContentId(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "invalid content id" {
			t.Errorf("Ожидалось сообщение 'invalid content id', получено '%s'", body)
		}
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		service := &TestCommentService{
			getByContentIdErr: errors.New("database error"),
		}
		handler := NewCommentHandler(service)

		contentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/comment/content/"+contentID.String(), nil)
		w := httptest.NewRecorder()
		handler.GetByContentId(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Ожидался статус %d, получен %d", http.StatusInternalServerError, w.Code)
		}
		if body := strings.TrimSpace(w.Body.String()); body != "cannot get comments" {
			t.Errorf("Ожидалось сообщение 'cannot get comments', получено '%s'", body)
		}
	})
}
