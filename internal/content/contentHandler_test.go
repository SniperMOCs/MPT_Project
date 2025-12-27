package content

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Мок для ContentService
type TestContentService struct {
	createdContent   []CreateContentData
	uploadedMedia    []UploadMediaData
	updatedContent   []UpdateContentData
	deletedContent   []DeleteContentData
	contentByID      map[uuid.UUID]*ContentResponse
	contentByAuthor  map[uuid.UUID][]ContentResponse
	followingContent map[uuid.UUID][]ContentResponse
	allContent       []ContentResponse
	errors           map[string]error
	callLog          []string
}

type CreateContentData struct {
	Description string
	AuthorId    uuid.UUID
}

type UploadMediaData struct {
	ContentId uuid.UUID
	UserId    uuid.UUID
	Files     []*multipart.FileHeader
}

type UpdateContentData struct {
	Description *string
	IsVisible   *bool
	MediaURLs   *[]string
	ContentId   uuid.UUID
	UserId      uuid.UUID
}

type DeleteContentData struct {
	ContentId uuid.UUID
	UserId    uuid.UUID
}

func NewTestContentService() *TestContentService {
	return &TestContentService{
		contentByID:      make(map[uuid.UUID]*ContentResponse),
		contentByAuthor:  make(map[uuid.UUID][]ContentResponse),
		followingContent: make(map[uuid.UUID][]ContentResponse),
		errors:           make(map[string]error),
	}
}

func (s *TestContentService) Create(ctx context.Context, description string, authorId uuid.UUID) (uuid.UUID, error) {
	s.callLog = append(s.callLog, "Create")
	if err, ok := s.errors["Create"]; ok {
		return uuid.Nil, err
	}

	s.createdContent = append(s.createdContent, CreateContentData{
		Description: description,
		AuthorId:    authorId,
	})

	// Создаем тестовый контент
	contentID := uuid.New()
	content := &ContentResponse{
		ID:          contentID,
		Description: description,
		Author: ContentAuthor{
			ID: authorId,
		},
	}
	s.contentByID[contentID] = content

	return contentID, nil
}

func (s *TestContentService) GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByID")
	if err, ok := s.errors["GetByID"]; ok {
		return nil, err
	}
	content, exists := s.contentByID[id]
	if !exists {
		return nil, errors.New("content not found")
	}
	return content, nil
}

func (s *TestContentService) GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByAuthor")
	if err, ok := s.errors["GetByAuthor"]; ok {
		return nil, err
	}
	content, exists := s.contentByAuthor[authorId]
	if !exists {
		return []ContentResponse{}, nil
	}
	return content, nil
}

func (s *TestContentService) GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByFollowing")
	if err, ok := s.errors["GetByFollowing"]; ok {
		return nil, err
	}
	content, exists := s.followingContent[userId]
	if !exists {
		return []ContentResponse{}, nil
	}
	return content, nil
}

func (s *TestContentService) GetAll(ctx context.Context) ([]ContentResponse, error) {
	s.callLog = append(s.callLog, "GetAll")
	if err, ok := s.errors["GetAll"]; ok {
		return nil, err
	}
	return s.allContent, nil
}

func (s *TestContentService) Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	s.callLog = append(s.callLog, "Delete")
	if err, ok := s.errors["Delete"]; ok {
		return err
	}

	content, exists := s.contentByID[id]
	if !exists {
		return errors.New("content not found")
	}
	if content.Author.ID != userId {
		return errors.New("forbidden")
	}

	s.deletedContent = append(s.deletedContent, DeleteContentData{
		ContentId: id,
		UserId:    userId,
	})

	return nil
}

func (s *TestContentService) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, userId uuid.UUID) error {
	s.callLog = append(s.callLog, "Update")
	if err, ok := s.errors["Update"]; ok {
		return err
	}

	content, exists := s.contentByID[id]
	if !exists {
		return errors.New("content not found")
	}
	if content.Author.ID != userId {
		return errors.New("forbidden")
	}

	s.updatedContent = append(s.updatedContent, UpdateContentData{
		Description: description,
		IsVisible:   isVisible,
		MediaURLs:   mediaURLs,
		ContentId:   id,
		UserId:      userId,
	})

	return nil
}

func (s *TestContentService) UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error {
	s.callLog = append(s.callLog, "UploadMedia")
	if err, ok := s.errors["UploadMedia"]; ok {
		return err
	}

	s.uploadedMedia = append(s.uploadedMedia, UploadMediaData{
		ContentId: contentId,
		UserId:    userId,
		Files:     files,
	})

	return nil
}

func (s *TestContentService) AddContent(content *ContentResponse) {
	s.contentByID[content.ID] = content
}

func (s *TestContentService) SetContentByAuthor(authorId uuid.UUID, content []ContentResponse) {
	s.contentByAuthor[authorId] = content
}

func (s *TestContentService) SetFollowingContent(userId uuid.UUID, content []ContentResponse) {
	s.followingContent[userId] = content
}

func (s *TestContentService) SetAllContent(content []ContentResponse) {
	s.allContent = content
}

func (s *TestContentService) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestContentService) Reset() {
	s.createdContent = []CreateContentData{}
	s.uploadedMedia = []UploadMediaData{}
	s.updatedContent = []UpdateContentData{}
	s.deletedContent = []DeleteContentData{}
	s.contentByID = make(map[uuid.UUID]*ContentResponse)
	s.contentByAuthor = make(map[uuid.UUID][]ContentResponse)
	s.followingContent = make(map[uuid.UUID][]ContentResponse)
	s.allContent = []ContentResponse{}
	s.errors = make(map[string]error)
	s.callLog = []string{}
}

func (s *TestContentService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func TestContentHandler_Create(t *testing.T) {
	t.Run("Успешное создание контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		dto := ContentDTO{
			Description: "Test post description",
		}

		reqBody, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/content/create", bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Create() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}

		if !service.WasCalled("Create") {
			t.Error("Create() не был вызван в сервисе")
		}

		if len(service.createdContent) != 1 {
			t.Errorf("Создано %d контента, ожидалось 1", len(service.createdContent))
		}

		createdData := service.createdContent[0]
		if createdData.Description != dto.Description {
			t.Errorf("Description = %s, ожидался %s", createdData.Description, dto.Description)
		}
		if createdData.AuthorId != userId {
			t.Errorf("AuthorId = %v, ожидался %v", createdData.AuthorId, userId)
		}

		// Проверяем что вернулся UUID в ответе
		var responseID uuid.UUID
		err := json.NewDecoder(w.Body).Decode(&responseID)
		if err != nil {
			t.Errorf("Не удалось декодировать UUID из ответа: %v", err)
		}
		if responseID == uuid.Nil {
			t.Error("В ответе должен быть не-nil UUID")
		}
	})

	t.Run("Ошибка при неверном методе", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/content/create", nil)
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Create() вернул статус %d, ожидался %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка при неверном теле запроса", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/content/create", strings.NewReader("invalid json"))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Create() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
		if !strings.Contains(w.Body.String(), "invalid body") {
			t.Error("Create() должен вернуть сообщение об ошибке")
		}
	})

	t.Run("Ошибка сервера при создании", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		dto := ContentDTO{
			Description: "Test post",
		}

		reqBody, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPost, "/api/content/create", bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		service.SetError("Create", errors.New("service error"))

		handler.Create(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Create() вернул статус %d, ожидался %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestContentHandler_GetByID(t *testing.T) {
	t.Run("Успешное получение контента по ID", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		contentId := uuid.New()
		expectedContent := &ContentResponse{
			ID:          contentId,
			Description: "Test content",
			Author: ContentAuthor{
				ID:       uuid.New(),
				Username: "testuser",
			},
		}
		service.AddContent(expectedContent)

		req := httptest.NewRequest(http.MethodGet, "/api/content/"+contentId.String(), nil)
		w := httptest.NewRecorder()

		handler.GetByID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByID() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("GetByID() должен вернуть Content-Type: application/json")
		}

		var response ContentResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("GetByID() вернул невалидный JSON: %v", err)
		}
		if response.ID != contentId {
			t.Errorf("GetByID() вернул контент с ID %v, ожидался %v", response.ID, contentId)
		}
	})

	t.Run("Ошибка при неверном UUID", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/content/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.GetByID(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetByID() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("Ошибка если контент не найден", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		contentId := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/content/"+contentId.String(), nil)
		w := httptest.NewRecorder()

		handler.GetByID(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetByID() вернул статус %d, ожидался %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestContentHandler_GetByAuthor(t *testing.T) {
	t.Run("Успешное получение контента по автору", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		authorId := uuid.New()
		contentList := []ContentResponse{
			{
				ID:          uuid.New(),
				Description: "Post 1",
				Author: ContentAuthor{
					ID: authorId,
				},
			},
			{
				ID:          uuid.New(),
				Description: "Post 2",
				Author: ContentAuthor{
					ID: authorId,
				},
			},
		}
		service.SetContentByAuthor(authorId, contentList)

		req := httptest.NewRequest(http.MethodGet, "/api/content/user/"+authorId.String(), nil)
		w := httptest.NewRecorder()

		handler.GetByAuthor(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetByAuthor() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}

		var response []ContentResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("GetByAuthor() вернул невалидный JSON: %v", err)
		}
		if len(response) != 2 {
			t.Errorf("GetByAuthor() вернул %d контентов, ожидалось 2", len(response))
		}
	})

	t.Run("Ошибка при неверном UUID автора", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/content/user/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.GetByAuthor(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("GetByAuthor() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("Ошибка сервера", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		authorId := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/content/user/"+authorId.String(), nil)
		w := httptest.NewRecorder()

		service.SetError("GetByAuthor", errors.New("database error"))

		handler.GetByAuthor(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("GetByAuthor() вернул статус %d, ожидался %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestContentHandler_Delete(t *testing.T) {
	t.Run("Успешное удаление контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentId := uuid.New()

		content := &ContentResponse{
			ID: contentId,
			Author: ContentAuthor{
				ID: userId,
			},
		}
		service.AddContent(content)

		req := httptest.NewRequest(http.MethodDelete, "/api/content/delete/"+contentId.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Delete() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}

		if !service.WasCalled("Delete") {
			t.Error("Delete() не был вызван в сервисе")
		}

		if len(service.deletedContent) != 1 {
			t.Errorf("Удалено %d контента, ожидалось 1", len(service.deletedContent))
		}

		deletedData := service.deletedContent[0]
		if deletedData.ContentId != contentId {
			t.Errorf("ContentId = %v, ожидался %v", deletedData.ContentId, contentId)
		}
		if deletedData.UserId != userId {
			t.Errorf("UserId = %v, ожидался %v", deletedData.UserId, userId)
		}
	})

	t.Run("Ошибка при неверном UUID контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/content/delete/invalid-uuid", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Delete() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("Ошибка при удалении чужого контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		authorId := uuid.New()
		otherUserId := uuid.New()
		contentId := uuid.New()

		content := &ContentResponse{
			ID: contentId,
			Author: ContentAuthor{
				ID: authorId,
			},
		}
		service.AddContent(content)

		req := httptest.NewRequest(http.MethodDelete, "/api/content/delete/"+contentId.String(), nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, otherUserId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Delete(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Delete() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestContentHandler_Update(t *testing.T) {
	t.Run("Успешное обновление контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentId := uuid.New()

		content := &ContentResponse{
			ID: contentId,
			Author: ContentAuthor{
				ID: userId,
			},
		}
		service.AddContent(content)

		dto := UpdatePostDTO{
			Description: stringPtr("Updated description"),
			IsVisible:   boolPtr(false),
			MediaURLs:   []string{"/uploads/image1.jpg"},
		}

		reqBody, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPatch, "/api/content/update/"+contentId.String(), bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Update() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}

		if !service.WasCalled("Update") {
			t.Error("Update() не был вызван в сервисе")
		}

		if len(service.updatedContent) != 1 {
			t.Errorf("Обновлено %d контента, ожидалось 1", len(service.updatedContent))
		}

		updatedData := service.updatedContent[0]
		if *updatedData.Description != *dto.Description {
			t.Errorf("Description = %s, ожидался %s", *updatedData.Description, *dto.Description)
		}
		if *updatedData.IsVisible != *dto.IsVisible {
			t.Errorf("IsVisible = %v, ожидался %v", *updatedData.IsVisible, *dto.IsVisible)
		}
		if len(*updatedData.MediaURLs) != len(dto.MediaURLs) {
			t.Errorf("MediaURLs содержит %d элементов, ожидалось %d", len(*updatedData.MediaURLs), len(dto.MediaURLs))
		}
	})

	t.Run("Ошибка при неверном теле запроса", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentId := uuid.New()

		req := httptest.NewRequest(http.MethodPatch, "/api/content/update/"+contentId.String(), strings.NewReader("invalid json"))
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Update() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestContentHandler_UploadMedia(t *testing.T) {
	t.Run("Успешная загрузка медиа", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentId := uuid.New()

		content := &ContentResponse{
			ID: contentId,
			Author: ContentAuthor{
				ID: userId,
			},
		}
		service.AddContent(content)

		// Создаем multipart запрос
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Добавляем файл
		part, _ := writer.CreateFormFile("files", "test.jpg")
		io.WriteString(part, "fake image content")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/content/media/"+contentId.String(), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadMedia(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("UploadMedia() вернул статус %d, ожидался %d", w.Code, http.StatusCreated)
		}

		if !service.WasCalled("UploadMedia") {
			t.Error("UploadMedia() не был вызван в сервисе")
		}

		if len(service.uploadedMedia) != 1 {
			t.Errorf("Загружено %d медиа, ожидалось 1", len(service.uploadedMedia))
		}
	})

	t.Run("Ошибка при отсутствии файлов", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentId := uuid.New()

		// Создаем multipart запрос без файлов
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/content/media/"+contentId.String(), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadMedia(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("UploadMedia() вернул статус %d, ожидался %d", w.Code, http.StatusBadRequest)
		}
		if !strings.Contains(w.Body.String(), "no files provided") {
			t.Error("UploadMedia() должен вернуть сообщение об отсутствии файлов")
		}
	})
}

func TestContentHandler_GetSubscriptions(t *testing.T) {
	t.Run("Успешное получение ленты подписок", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		contentList := []ContentResponse{
			{
				ID:          uuid.New(),
				Description: "Post from followed user",
				Author: ContentAuthor{
					ID:       uuid.New(),
					Username: "followeduser",
				},
			},
		}
		service.SetFollowingContent(userId, contentList)

		req := httptest.NewRequest(http.MethodGet, "/api/content/subscriptions", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetSubscriptions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetSubscriptions() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("GetSubscriptions() должен вернуть Content-Type: application/json")
		}
	})

	t.Run("Ошибка сервера", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		userId := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/content/subscriptions", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userId)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		service.SetError("GetByFollowing", errors.New("database error"))

		handler.GetSubscriptions(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("GetSubscriptions() вернул статус %d, ожидался %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestContentHandler_GetAll(t *testing.T) {
	t.Run("Успешное получение всего контента", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		contentList := []ContentResponse{
			{
				ID:          uuid.New(),
				Description: "Public post 1",
				Author: ContentAuthor{
					ID:       uuid.New(),
					Username: "user1",
				},
			},
			{
				ID:          uuid.New(),
				Description: "Public post 2",
				Author: ContentAuthor{
					ID:       uuid.New(),
					Username: "user2",
				},
			},
		}
		service.SetAllContent(contentList)

		req := httptest.NewRequest(http.MethodGet, "/api/content/all", nil)
		w := httptest.NewRecorder()

		handler.GetAll(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetAll() вернул статус %d, ожидался %d", w.Code, http.StatusOK)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("GetAll() должен вернуть Content-Type: application/json")
		}
	})

	t.Run("Ошибка сервера", func(t *testing.T) {
		service := NewTestContentService()
		handler := NewContentHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/content/all", nil)
		w := httptest.NewRecorder()

		service.SetError("GetAll", errors.New("database error"))

		handler.GetAll(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("GetAll() вернул статус %d, ожидался %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestContentHandler_InvalidHTTPMethods(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	testCases := []struct {
		name    string
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{"Create with GET", http.MethodGet, "/api/content/create", handler.Create},
		{"GetByID with POST", http.MethodPost, "/api/content/123", handler.GetByID},
		{"GetByAuthor with POST", http.MethodPost, "/api/content/user/123", handler.GetByAuthor},
		{"Delete with GET", http.MethodGet, "/api/content/delete/123", handler.Delete},
		{"Update with GET", http.MethodGet, "/api/content/update/123", handler.Update},
		{"UploadMedia with GET", http.MethodGet, "/api/content/media/123", handler.UploadMedia},
		{"GetSubscriptions with POST", http.MethodPost, "/api/content/subscriptions", handler.GetSubscriptions},
		{"GetAll with POST", http.MethodPost, "/api/content/all", handler.GetAll},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			tc.handler(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code,
				"Ожидался статус 405 для метода %s", tc.method)
		})
	}
}

func TestContentHandler_GetByID_ServiceError(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	contentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/content/"+contentID.String(), nil)
	w := httptest.NewRecorder()

	service.SetError("GetByID", errors.New("database error"))

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "content not found")
}

func TestContentHandler_Update_InvalidBody(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	userID := uuid.New()
	contentID := uuid.New()

	req := httptest.NewRequest(http.MethodPatch, "/api/content/update/"+contentID.String(),
		strings.NewReader("{invalid json"))
	ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid body")
}

func TestContentHandler_UploadMedia_InvalidMultipart(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	userID := uuid.New()
	contentID := uuid.New()

	// Не multipart запрос
	req := httptest.NewRequest(http.MethodPost, "/api/content/media/"+contentID.String(),
		strings.NewReader("not multipart"))
	ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadMedia(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid multipart")
}

func TestContentHandler_GetSubscriptions_ServiceError(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/content/subscriptions", nil)
	ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	service.SetError("GetByFollowing", errors.New("database connection failed"))

	handler.GetSubscriptions(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "cannot get subscriptions")
}

func TestContentHandler_GetAll_ServiceError(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/content/all", nil)
	w := httptest.NewRecorder()

	service.SetError("GetAll", errors.New("database error"))

	handler.GetAll(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "cannot get content")
}

func TestContentHandler_UploadMedia_ServiceError(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	userID := uuid.New()
	contentID := uuid.New()

	content := &ContentResponse{
		ID: contentID,
		Author: ContentAuthor{
			ID: userID,
		},
	}
	service.AddContent(content)

	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", "test.jpg")
	io.WriteString(part, "fake image content")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/content/media/"+contentID.String(), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	service.SetError("UploadMedia", errors.New("storage is full"))

	handler.UploadMedia(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "storage is full")
}

func TestContentHandler_UploadMedia_NoFiles(t *testing.T) {
	service := NewTestContentService()
	handler := NewContentHandler(service)

	userID := uuid.New()
	contentID := uuid.New()

	// Создаем multipart запрос БЕЗ файлов
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Не добавляем файлы
	writer.WriteField("other_field", "value")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/content/media/"+contentID.String(), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.UploadMedia(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "no files provided")
}

// Вспомогательная функция для создания указателя на bool
func boolPtr(b bool) *bool {
	return &b
}
