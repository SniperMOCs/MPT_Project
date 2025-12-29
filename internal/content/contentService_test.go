package content

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Мок для ContentRepository
type TestContentRepository struct {
	contentsByID     map[uuid.UUID]*ContentResponse
	contentsByAuthor map[uuid.UUID][]ContentResponse
	followingContent map[uuid.UUID][]ContentResponse
	allContent       []ContentResponse
	createdContent   []*Content
	updatedContent   []UpdatedContentData
	deletedContent   []uuid.UUID
	errors           map[string]error
	callLog          []string
}

type UpdatedContentData struct {
	ID          uuid.UUID
	Description *string
	IsVisible   *bool
	MediaURLs   *[]string
}

func NewTestContentRepository() *TestContentRepository {
	return &TestContentRepository{
		contentsByID:     make(map[uuid.UUID]*ContentResponse),
		contentsByAuthor: make(map[uuid.UUID][]ContentResponse),
		followingContent: make(map[uuid.UUID][]ContentResponse),
		errors:           make(map[string]error),
	}
}

func (r *TestContentRepository) Create(ctx context.Context, content *Content) (uuid.UUID, error) {
	r.callLog = append(r.callLog, "Create")
	if err, ok := r.errors["Create"]; ok {
		return uuid.Nil, err
	}

	contentResponse := &ContentResponse{
		ID:          content.ID,
		Description: content.Description,
		MediaURLs:   content.MediaURLs,
		CreatedAt:   content.CreatedAt,
		IsVisible:   content.IsVisible,
		Author: ContentAuthor{
			ID: content.AuthorId,
		},
	}

	r.contentsByID[content.ID] = contentResponse
	r.createdContent = append(r.createdContent, content)

	if _, exists := r.contentsByAuthor[content.AuthorId]; !exists {
		r.contentsByAuthor[content.AuthorId] = []ContentResponse{}
	}
	r.contentsByAuthor[content.AuthorId] = append(r.contentsByAuthor[content.AuthorId], *contentResponse)

	return content.ID, nil
}

func (r *TestContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error) {
	r.callLog = append(r.callLog, "GetByID")
	if err, ok := r.errors["GetByID"]; ok {
		return nil, err
	}
	content, exists := r.contentsByID[id]
	if !exists {
		return nil, errors.New("content not found")
	}
	return content, nil
}

func (r *TestContentRepository) GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error) {
	r.callLog = append(r.callLog, "GetByAuthor")
	if err, ok := r.errors["GetByAuthor"]; ok {
		return nil, err
	}
	content, exists := r.contentsByAuthor[authorId]
	if !exists {
		return []ContentResponse{}, nil
	}
	return content, nil
}

func (r *TestContentRepository) GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error) {
	r.callLog = append(r.callLog, "GetByFollowing")
	if err, ok := r.errors["GetByFollowing"]; ok {
		return nil, err
	}
	content, exists := r.followingContent[userId]
	if !exists {
		return []ContentResponse{}, nil
	}
	return content, nil
}

func (r *TestContentRepository) GetAll(ctx context.Context) ([]ContentResponse, error) {
	r.callLog = append(r.callLog, "GetAll")
	if err, ok := r.errors["GetAll"]; ok {
		return nil, err
	}
	return r.allContent, nil
}

func (r *TestContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.callLog = append(r.callLog, "Delete")
	if err, ok := r.errors["Delete"]; ok {
		return err
	}

	content, exists := r.contentsByID[id]
	if !exists {
		return errors.New("content not found")
	}

	content.IsVisible = false
	r.deletedContent = append(r.deletedContent, id)

	return nil
}

func (r *TestContentRepository) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID) error {
	r.callLog = append(r.callLog, "Update")
	if err, ok := r.errors["Update"]; ok {
		return err
	}

	content, exists := r.contentsByID[id]
	if !exists {
		return errors.New("content not found")
	}

	if description != nil {
		content.Description = *description
	}
	if isVisible != nil {
		content.IsVisible = *isVisible
	}
	if mediaURLs != nil {
		content.MediaURLs = *mediaURLs
	}

	r.updatedContent = append(r.updatedContent, UpdatedContentData{
		ID:          id,
		Description: description,
		IsVisible:   isVisible,
		MediaURLs:   mediaURLs,
	})

	return nil
}

func (r *TestContentRepository) AddContent(content *ContentResponse) {
	r.contentsByID[content.ID] = content

	if _, exists := r.contentsByAuthor[content.Author.ID]; !exists {
		r.contentsByAuthor[content.Author.ID] = []ContentResponse{}
	}
	r.contentsByAuthor[content.Author.ID] = append(r.contentsByAuthor[content.Author.ID], *content)

	r.allContent = append(r.allContent, *content)
}

func (r *TestContentRepository) SetFollowingContent(userId uuid.UUID, content []ContentResponse) {
	r.followingContent[userId] = content
}

func (r *TestContentRepository) SetError(methodName string, err error) {
	r.errors[methodName] = err
}

func (r *TestContentRepository) Reset() {
	r.contentsByID = make(map[uuid.UUID]*ContentResponse)
	r.contentsByAuthor = make(map[uuid.UUID][]ContentResponse)
	r.followingContent = make(map[uuid.UUID][]ContentResponse)
	r.allContent = []ContentResponse{}
	r.createdContent = []*Content{}
	r.updatedContent = []UpdatedContentData{}
	r.deletedContent = []uuid.UUID{}
	r.errors = make(map[string]error)
	r.callLog = []string{}
}

func (r *TestContentRepository) WasCalled(methodName string) bool {
	for _, call := range r.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

// Мок для MediaStorage
type TestMediaStorage struct {
	savedFiles  map[uuid.UUID][]*multipart.FileHeader
	deletedURLs []string
	savedURLs   map[uuid.UUID][]string
	errors      map[string]error
	callLog     []string
}

func NewTestMediaStorage() *TestMediaStorage {
	return &TestMediaStorage{
		savedFiles: make(map[uuid.UUID][]*multipart.FileHeader),
		savedURLs:  make(map[uuid.UUID][]string),
		errors:     make(map[string]error),
	}
}

func (s *TestMediaStorage) Save(ctx context.Context, contentId uuid.UUID, files []*multipart.FileHeader) ([]string, error) {
	s.callLog = append(s.callLog, "Save")
	if err, ok := s.errors["Save"]; ok {
		return nil, err
	}

	s.savedFiles[contentId] = files

	urls := make([]string, len(files))
	for i := range files {
		urls[i] = "/uploads/content/" + contentId.String() + "/file" + string(rune(i)) + ".jpg"
	}

	s.savedURLs[contentId] = urls
	return urls, nil
}

func (s *TestMediaStorage) Delete(ctx context.Context, urls []string) error {
	s.callLog = append(s.callLog, "Delete")
	if err, ok := s.errors["Delete"]; ok {
		return err
	}

	s.deletedURLs = append(s.deletedURLs, urls...)
	return nil
}

func (s *TestMediaStorage) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestMediaStorage) Reset() {
	s.savedFiles = make(map[uuid.UUID][]*multipart.FileHeader)
	s.deletedURLs = []string{}
	s.savedURLs = make(map[uuid.UUID][]string)
	s.errors = make(map[string]error)
	s.callLog = []string{}
}

func (s *TestMediaStorage) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func createTestContentResponse(authorId uuid.UUID) *ContentResponse {
	return &ContentResponse{
		ID:          uuid.New(),
		Description: "Test content description",
		MediaURLs:   []string{"/uploads/image1.jpg"},
		CreatedAt:   time.Now(),
		IsVisible:   true,
		Author: ContentAuthor{
			ID:        authorId,
			Username:  "testuser",
			AvatarURL: stringPtr("/uploads/avatar.jpg"),
		},
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestContentService_Create(t *testing.T) {
	t.Run("Успешное создание контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := &contentService{
			repo:         repo,
			mediaStorage: NewTestMediaStorage(),
		}

		ctx := context.Background()
		description := "Test post description"
		authorId := uuid.New()

		contentID, err := service.Create(ctx, description, authorId)

		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if !repo.WasCalled("Create") {
			t.Error("Create() не был вызван в репозитории")
		}

		if contentID == uuid.Nil {
			t.Error("Create() должен вернуть не-nil UUID")
		}

		if len(repo.createdContent) != 1 {
			t.Errorf("Создано %d контента, ожидалось 1", len(repo.createdContent))
		}

		createdContent := repo.createdContent[0]
		if createdContent.Description != description {
			t.Errorf("Description = %s, ожидался %s", createdContent.Description, description)
		}
		if createdContent.AuthorId != authorId {
			t.Errorf("AuthorId = %v, ожидался %v", createdContent.AuthorId, authorId)
		}
		if !createdContent.IsVisible {
			t.Error("IsVisible должен быть true")
		}
		if len(createdContent.MediaURLs) != 0 {
			t.Error("MediaURLs должен быть пустым при создании")
		}
	})

	t.Run("Ошибка при создании контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := &contentService{
			repo:         repo,
			mediaStorage: NewTestMediaStorage(),
		}

		ctx := context.Background()
		repo.SetError("Create", errors.New("repository error"))

		contentID, err := service.Create(ctx, "Test description", uuid.New())

		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке в репозитории")
		}
		if contentID != uuid.Nil {
			t.Error("Create() должен вернуть uuid.Nil при ошибке")
		}
	})
}

func TestContentService_GetByID(t *testing.T) {
	t.Run("Успешное получение контента по ID", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		content := createTestContentResponse(uuid.New())
		repo.AddContent(content)

		foundContent, err := service.GetByID(context.Background(), content.ID)

		if err != nil {
			t.Errorf("GetByID() error = %v", err)
		}
		if foundContent.ID != content.ID {
			t.Errorf("GetByID() вернул контент с ID %v, ожидался %v", foundContent.ID, content.ID)
		}
	})

	t.Run("Ошибка если контент не найден", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		_, err := service.GetByID(context.Background(), uuid.New())

		if err == nil {
			t.Error("GetByID() должен вернуть ошибку если контент не найден")
		}
	})
}

func TestContentService_GetByAuthor(t *testing.T) {
	t.Run("Успешное получение контента по автору", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		content1 := createTestContentResponse(authorId)
		content2 := createTestContentResponse(authorId)

		repo.AddContent(content1)
		repo.AddContent(content2)

		foundContent, err := service.GetByAuthor(context.Background(), authorId)

		if err != nil {
			t.Errorf("GetByAuthor() error = %v", err)
		}
		if len(foundContent) != 2 {
			t.Errorf("GetByAuthor() вернул %d контентов, ожидалось 2", len(foundContent))
		}
	})

	t.Run("Пустой результат если у автора нет контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		foundContent, err := service.GetByAuthor(context.Background(), uuid.New())

		if err != nil {
			t.Errorf("GetByAuthor() error = %v", err)
		}
		if len(foundContent) != 0 {
			t.Errorf("GetByAuthor() вернул %d контентов, ожидалось 0", len(foundContent))
		}
	})
}

func TestContentService_GetByFollowing(t *testing.T) {
	t.Run("Успешное получение контента подписок", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		userId := uuid.New()
		followingContent := []ContentResponse{
			*createTestContentResponse(uuid.New()),
			*createTestContentResponse(uuid.New()),
		}

		repo.SetFollowingContent(userId, followingContent)

		foundContent, err := service.GetByFollowing(context.Background(), userId)

		if err != nil {
			t.Errorf("GetByFollowing() error = %v", err)
		}
		if len(foundContent) != 2 {
			t.Errorf("GetByFollowing() вернул %d контентов, ожидалось 2", len(foundContent))
		}
	})

	t.Run("Ошибка при получении контента подписок", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		repo.SetError("GetByFollowing", errors.New("database error"))

		_, err := service.GetByFollowing(context.Background(), uuid.New())

		if err == nil {
			t.Error("GetByFollowing() должен вернуть ошибку")
		}
	})
}

func TestContentService_GetAll(t *testing.T) {
	t.Run("Успешное получение всего контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		content1 := createTestContentResponse(uuid.New())
		content2 := createTestContentResponse(uuid.New())

		repo.AddContent(content1)
		repo.AddContent(content2)

		allContent, err := service.GetAll(context.Background())

		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}
		if len(allContent) != 2 {
			t.Errorf("GetAll() вернул %d контентов, ожидалось 2", len(allContent))
		}
	})

	t.Run("Ошибка при получении всего контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		repo.SetError("GetAll", errors.New("database error"))

		_, err := service.GetAll(context.Background())

		if err == nil {
			t.Error("GetAll() должен вернуть ошибку")
		}
	})
}

func TestContentService_Delete(t *testing.T) {
	t.Run("Успешное удаление контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		err := service.Delete(context.Background(), content.ID, authorId)

		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		if !repo.WasCalled("Delete") {
			t.Error("Delete() не был вызван в репозитории")
		}

		if len(repo.deletedContent) != 1 || repo.deletedContent[0] != content.ID {
			t.Error("Контент не был помечен как удаленный")
		}
	})

	t.Run("Ошибка при удалении чужого контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		otherUserId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		err := service.Delete(context.Background(), content.ID, otherUserId)

		if err == nil {
			t.Error("Delete() должен вернуть ошибку при удалении чужого контента")
		}
		if err != nil && err.Error() != "forbidden" {
			t.Errorf("Delete() вернул ошибку %v, ожидалась 'forbidden'", err)
		}
	})

	t.Run("Ошибка если контент не найден", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		err := service.Delete(context.Background(), uuid.New(), uuid.New())

		if err == nil {
			t.Error("Delete() должен вернуть ошибку если контент не найден")
		}
		if err != nil && err.Error() != "content not found" {
			t.Errorf("Delete() вернул ошибку %v, ожидалась 'content not found'", err)
		}
	})
}

func TestContentService_Update(t *testing.T) {
	t.Run("Успешное обновление контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		newDescription := "Updated description"
		newIsVisible := false
		newMediaURLs := []string{"/uploads/new-image.jpg"}

		err := service.Update(context.Background(), &newDescription, &newIsVisible, &newMediaURLs, content.ID, authorId)

		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if !repo.WasCalled("Update") {
			t.Error("Update() не был вызван в репозитории")
		}

		if !mediaStorage.WasCalled("Delete") {
			t.Error("Старые медиа файлы должны быть удалены")
		}

		if len(mediaStorage.deletedURLs) != 1 || mediaStorage.deletedURLs[0] != "/uploads/image1.jpg" {
			t.Error("Старые медиа URL не были удалены")
		}
	})

	t.Run("Успешное обновление только описания", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		newDescription := "Only description updated"

		err := service.Update(context.Background(), &newDescription, nil, nil, content.ID, authorId)

		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		if !repo.WasCalled("Update") {
			t.Error("Update() не был вызван в репозитории")
		}
	})

	t.Run("Ошибка при обновлении чужого контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		otherUserId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		newDescription := "Updated description"

		err := service.Update(context.Background(), &newDescription, nil, nil, content.ID, otherUserId)

		if err == nil {
			t.Error("Update() должен вернуть ошибку при обновлении чужого контента")
		}
		if err != nil && err.Error() != "forbidden" {
			t.Errorf("Update() вернул ошибку %v, ожидалась 'forbidden'", err)
		}
	})

	t.Run("Ошибка при пустых mediaURLs", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		repo.AddContent(content)

		emptyMediaURLs := []string{}

		err := service.Update(context.Background(), nil, nil, &emptyMediaURLs, content.ID, authorId)

		if err == nil {
			t.Error("Update() должен вернуть ошибку при пустых mediaURLs")
		}
		if err != nil && err.Error() != "content must have at least one media" {
			t.Errorf("Update() вернул ошибку %v, ожидалась 'content must have at least one media'", err)
		}
	})
}

func TestContentService_UploadMedia(t *testing.T) {
	t.Run("Успешная загрузка медиа", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		ctx := context.Background()
		authorId := uuid.New()
		contentId := uuid.New()

		existingContent := createTestContentResponse(authorId)
		existingContent.ID = contentId
		existingContent.MediaURLs = []string{"/uploads/old.jpg"}
		repo.AddContent(existingContent)

		files := []*multipart.FileHeader{
			{},
			{},
		}

		err := service.UploadMedia(ctx, contentId, authorId, files)

		if err != nil {
			t.Errorf("UploadMedia() error = %v", err)
		}

		if !mediaStorage.WasCalled("Save") {
			t.Error("Save() не был вызван в mediaStorage")
		}

		if !repo.WasCalled("Update") {
			t.Error("Update() не был вызван в репозитории")
		}

		if len(mediaStorage.savedFiles[contentId]) != 2 {
			t.Errorf("Сохранено %d файлов, ожидалось 2", len(mediaStorage.savedFiles[contentId]))
		}
	})

	t.Run("Ошибка при загрузке медиа для чужого контента", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		ctx := context.Background()
		authorId := uuid.New()
		otherUserId := uuid.New()
		contentId := uuid.New()

		existingContent := createTestContentResponse(authorId)
		existingContent.ID = contentId
		repo.AddContent(existingContent)

		files := []*multipart.FileHeader{{}}

		err := service.UploadMedia(ctx, contentId, otherUserId, files)

		if err == nil {
			t.Error("UploadMedia() должен вернуть ошибку при загрузке медиа для чужого контента")
		}
		if err != nil && err.Error() != "forbidden" {
			t.Errorf("UploadMedia() вернул ошибку %v, ожидалась 'forbidden'", err)
		}
	})

	t.Run("Ошибка если контент не найден", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		ctx := context.Background()
		contentId := uuid.New()
		userId := uuid.New()

		files := []*multipart.FileHeader{{}}

		err := service.UploadMedia(ctx, contentId, userId, files)

		if err == nil {
			t.Error("UploadMedia() должен вернуть ошибку если контент не найден")
		}
		if err != nil && err.Error() != "content not found" {
			t.Errorf("UploadMedia() вернул ошибку %v, ожидалась 'content not found'", err)
		}
	})

	t.Run("Ошибка при сохранении медиа", func(t *testing.T) {
		repo := NewTestContentRepository()
		mediaStorage := NewTestMediaStorage()
		service := &contentService{
			repo:         repo,
			mediaStorage: mediaStorage,
		}

		ctx := context.Background()
		authorId := uuid.New()
		contentId := uuid.New()

		existingContent := createTestContentResponse(authorId)
		existingContent.ID = contentId
		repo.AddContent(existingContent)

		files := []*multipart.FileHeader{{}}
		mediaStorage.SetError("Save", errors.New("storage error"))

		err := service.UploadMedia(ctx, contentId, authorId, files)

		if err == nil {
			t.Error("UploadMedia() должен вернуть ошибку при ошибке сохранения медиа")
		}
	})
}

func TestContentService_Update_RepositoryError(t *testing.T) {
	repo := NewTestContentRepository()
	mediaStorage := NewTestMediaStorage()
	service := &contentService{
		repo:         repo,
		mediaStorage: mediaStorage,
	}

	authorId := uuid.New()
	content := createTestContentResponse(authorId)
	repo.AddContent(content)

	newDescription := "Updated description"

	repo.SetError("Update", errors.New("database error"))

	err := service.Update(context.Background(), &newDescription, nil, nil, content.ID, authorId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestContentService_UploadMedia_RepositoryUpdateError(t *testing.T) {
	repo := NewTestContentRepository()
	mediaStorage := NewTestMediaStorage()
	service := &contentService{
		repo:         repo,
		mediaStorage: mediaStorage,
	}

	ctx := context.Background()
	authorId := uuid.New()
	contentId := uuid.New()

	existingContent := createTestContentResponse(authorId)
	existingContent.ID = contentId
	repo.AddContent(existingContent)

	files := []*multipart.FileHeader{{}}

	mediaStorage.savedURLs[contentId] = []string{"/uploads/new.jpg"}

	repo.SetError("Update", errors.New("update error"))

	err := service.UploadMedia(ctx, contentId, authorId, files)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestContentService_Delete_RepositoryError(t *testing.T) {
	repo := NewTestContentRepository()
	service := NewContentService(repo)

	authorId := uuid.New()
	content := createTestContentResponse(authorId)
	repo.AddContent(content)

	repo.SetError("Delete", errors.New("database error"))

	err := service.Delete(context.Background(), content.ID, authorId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestContentService_Update_EmptyMediaURLs(t *testing.T) {
	t.Run("Ошибка при передаче пустого массива mediaURLs", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		content.MediaURLs = []string{"/uploads/existing.jpg"}
		repo.AddContent(content)

		emptyMediaURLs := []string{}

		err := service.Update(context.Background(), nil, nil, &emptyMediaURLs, content.ID, authorId)

		assert.Error(t, err)
		assert.Equal(t, "content must have at least one media", err.Error())
	})

	t.Run("nil mediaURLs - не проверяем на пустоту", func(t *testing.T) {
		repo := NewTestContentRepository()
		service := NewContentService(repo)

		authorId := uuid.New()
		content := createTestContentResponse(authorId)
		content.MediaURLs = []string{"/uploads/existing.jpg"}
		repo.AddContent(content)

		err := service.Update(context.Background(), nil, nil, nil, content.ID, authorId)

		assert.NoError(t, err)
	})
}
