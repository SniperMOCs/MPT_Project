package like

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/google/uuid"
)

type TestLikeRepository struct {
	likes        map[uuid.UUID]map[uuid.UUID]bool
	likesCount   map[uuid.UUID]int
	callLog      []string
	errors       map[string]error
	createdLikes []LikeData
	deletedLikes []LikeData
}

type LikeData struct {
	UserID    uuid.UUID
	ContentID uuid.UUID
}

func NewTestLikeRepository() *TestLikeRepository {
	return &TestLikeRepository{
		likes:      make(map[uuid.UUID]map[uuid.UUID]bool),
		likesCount: make(map[uuid.UUID]int),
		errors:     make(map[string]error),
		callLog:    []string{},
	}
}

func (r *TestLikeRepository) Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	r.callLog = append(r.callLog, "Create")
	if err, ok := r.errors["Create"]; ok {
		return err
	}

	if _, exists := r.likes[contentId]; !exists {
		r.likes[contentId] = make(map[uuid.UUID]bool)
	}

	if r.likes[contentId][userId] {
		return errors.New("like already exists")
	}

	r.likes[contentId][userId] = true
	r.likesCount[contentId]++
	r.createdLikes = append(r.createdLikes, LikeData{UserID: userId, ContentID: contentId})
	return nil
}

func (r *TestLikeRepository) Delete(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	r.callLog = append(r.callLog, "Delete")
	if err, ok := r.errors["Delete"]; ok {
		return err
	}

	if contentLikes, exists := r.likes[contentId]; exists {
		if contentLikes[userId] {
			delete(contentLikes, userId)
			r.likesCount[contentId]--
			r.deletedLikes = append(r.deletedLikes, LikeData{UserID: userId, ContentID: contentId})
			return nil
		}
	}
	return nil
}

func (r *TestLikeRepository) Exists(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) (bool, error) {
	r.callLog = append(r.callLog, "Exists")
	if err, ok := r.errors["Exists"]; ok {
		return false, err
	}

	if contentLikes, exists := r.likes[contentId]; exists {
		return contentLikes[userId], nil
	}
	return false, nil
}

func (r *TestLikeRepository) GetLikesCount(ctx context.Context, contentId uuid.UUID) (int, error) {
	r.callLog = append(r.callLog, "GetLikesCount")
	if err, ok := r.errors["GetLikesCount"]; ok {
		return 0, err
	}

	return r.likesCount[contentId], nil
}

func (r *TestLikeRepository) AddLike(userId uuid.UUID, contentId uuid.UUID) {
	if _, exists := r.likes[contentId]; !exists {
		r.likes[contentId] = make(map[uuid.UUID]bool)
	}
	r.likes[contentId][userId] = true
	r.likesCount[contentId]++
}

func (r *TestLikeRepository) SetError(methodName string, err error) {
	r.errors[methodName] = err
}

func (r *TestLikeRepository) SetLikesCount(contentId uuid.UUID, count int) {
	r.likesCount[contentId] = count
}

func (r *TestLikeRepository) Reset() {
	r.likes = make(map[uuid.UUID]map[uuid.UUID]bool)
	r.likesCount = make(map[uuid.UUID]int)
	r.callLog = []string{}
	r.errors = make(map[string]error)
	r.createdLikes = []LikeData{}
	r.deletedLikes = []LikeData{}
}

func (r *TestLikeRepository) WasCalled(methodName string) bool {
	for _, call := range r.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func (r *TestLikeRepository) GetCallCount(methodName string) int {
	count := 0
	for _, call := range r.callLog {
		if call == methodName {
			count++
		}
	}
	return count
}

// Мок для тестов NotificationService
type TestNotificationService struct {
	notifications []NotificationData
	callLog       []string
	errors        map[string]error
}

type NotificationData struct {
	FromUserID uuid.UUID
	ToUserID   uuid.UUID
	ContentID  uuid.UUID
	Type       notification.NotificationType
}

func NewTestNotificationService() *TestNotificationService {
	return &TestNotificationService{
		notifications: []NotificationData{},
		callLog:       []string{},
		errors:        make(map[string]error),
	}
}

func (s *TestNotificationService) Notify(ctx context.Context, fromUserID, toUserID, contentID uuid.UUID, ntype notification.NotificationType) (*notification.NotificationResponse, error) {
	s.callLog = append(s.callLog, "Notify")
	if err, ok := s.errors["Notify"]; ok {
		return nil, err
	}

	if fromUserID == toUserID {
		return nil, nil
	}

	s.notifications = append(s.notifications, NotificationData{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		ContentID:  contentID,
		Type:       ntype,
	})

	avatarURL := "/uploads/avatar.jpg"
	return &notification.NotificationResponse{
		ID:        uuid.New(),
		Recipient: toUserID,
		Type:      ntype,
		TargetId:  contentID,
		CreatedAt: time.Now(),
		Read:      false,
		Actor: content.ContentAuthor{
			ID:        fromUserID,
			Username:  "testuser",
			AvatarURL: &avatarURL,
		},
	}, nil
}

func (s *TestNotificationService) GetByRecipient(ctx context.Context, userID uuid.UUID) ([]notification.NotificationResponse, error) {
	return []notification.NotificationResponse{}, nil
}

func (s *TestNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *TestNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (s *TestNotificationService) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestNotificationService) Reset() {
	s.notifications = []NotificationData{}
	s.callLog = []string{}
	s.errors = make(map[string]error)
}

func (s *TestNotificationService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func (s *TestNotificationService) GetNotificationCount() int {
	return len(s.notifications)
}

// Мок для тестов ContentService
type TestContentService struct {
	contents map[uuid.UUID]*content.ContentResponse
	callLog  []string
	errors   map[string]error
}

func NewTestContentService() *TestContentService {
	return &TestContentService{
		contents: make(map[uuid.UUID]*content.ContentResponse),
		callLog:  []string{},
		errors:   make(map[string]error),
	}
}

func (s *TestContentService) GetByID(ctx context.Context, id uuid.UUID) (*content.ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByID")
	if err, ok := s.errors["GetByID"]; ok {
		return nil, err
	}

	contentResp, exists := s.contents[id]
	if !exists {
		return nil, errors.New("content not found")
	}

	return contentResp, nil
}

func (s *TestContentService) GetByIDWithAuthor(ctx context.Context, id uuid.UUID) (*content.Content, error) {
	s.callLog = append(s.callLog, "GetByID")
	if err, ok := s.errors["GetByID"]; ok {
		return nil, err
	}

	contentResp, exists := s.contents[id]
	if !exists {
		return nil, errors.New("content not found")
	}

	return &content.Content{
		ID:          contentResp.ID,
		Description: contentResp.Description,
		MediaURLs:   contentResp.MediaURLs,
		CreatedAt:   contentResp.CreatedAt,
		IsVisible:   contentResp.IsVisible,
		AuthorId:    contentResp.Author.ID,
	}, nil
}

func (s *TestContentService) Create(ctx context.Context, contentText string, authorID uuid.UUID) (uuid.UUID, error) {
	s.callLog = append(s.callLog, "Create")
	return uuid.New(), nil
}

func (s *TestContentService) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, authorID uuid.UUID) error {
	s.callLog = append(s.callLog, "Update")
	if err, ok := s.errors["Update"]; ok {
		return err
	}

	contentResp, exists := s.contents[id]
	if !exists {
		return errors.New("content not found")
	}

	if authorID != contentResp.Author.ID {
		return errors.New("forbidden")
	}

	if description != nil {
		contentResp.Description = *description
	}
	if isVisible != nil {
		contentResp.IsVisible = *isVisible
	}
	if mediaURLs != nil {
		if len(*mediaURLs) == 0 {
			return errors.New("content must have at least one media")
		}
		contentResp.MediaURLs = *mediaURLs
	}

	return nil
}

func (s *TestContentService) Delete(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error {
	s.callLog = append(s.callLog, "Delete")
	if err, ok := s.errors["Delete"]; ok {
		return err
	}

	contentResp, exists := s.contents[id]
	if !exists {
		return errors.New("content not found")
	}

	if authorID != contentResp.Author.ID {
		return errors.New("forbidden")
	}

	contentResp.IsVisible = false
	return nil
}

func (s *TestContentService) GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]content.ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByAuthor")
	result := []content.ContentResponse{}
	for _, contentResp := range s.contents {
		if contentResp.Author.ID == authorID {
			result = append(result, *contentResp)
		}
	}
	return result, nil
}

func (s *TestContentService) GetAll(ctx context.Context) ([]content.ContentResponse, error) {
	s.callLog = append(s.callLog, "GetAll")
	result := []content.ContentResponse{}
	for _, contentResp := range s.contents {
		result = append(result, *contentResp)
	}
	return result, nil
}

func (s *TestContentService) GetByFollowing(ctx context.Context, userID uuid.UUID) ([]content.ContentResponse, error) {
	s.callLog = append(s.callLog, "GetByFollowing")
	return []content.ContentResponse{}, nil
}

func (s *TestContentService) UploadMedia(ctx context.Context, contentID, authorID uuid.UUID, files []*multipart.FileHeader) error {
	s.callLog = append(s.callLog, "UploadMedia")
	return nil
}

func (s *TestContentService) AddSimpleContent(contentID, authorID uuid.UUID) {
	avatarURL := "/uploads/avatar.jpg"
	s.contents[contentID] = &content.ContentResponse{
		ID:          contentID,
		Description: "Test content",
		MediaURLs:   []string{"/uploads/image.jpg"},
		CreatedAt:   time.Now(),
		IsVisible:   true,
		Author: content.ContentAuthor{
			ID:        authorID,
			Username:  "testuser",
			AvatarURL: &avatarURL,
		},
	}
}

func (s *TestContentService) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestContentService) Reset() {
	s.contents = make(map[uuid.UUID]*content.ContentResponse)
	s.callLog = []string{}
	s.errors = make(map[string]error)
}

func (s *TestContentService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func createTestLikeService(repo *TestLikeRepository, ns *TestNotificationService, cs *TestContentService) *likeService {
	return &likeService{
		repo:                repo,
		notificationService: ns,
		contentService:      cs,
	}
}

func TestLikeService_Create(t *testing.T) {
	t.Run("Успешное создание лайка", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()
		authorID := uuid.New()

		cs.AddSimpleContent(contentID, authorID)

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Create() error = %v, ожидалось nil", err)
		}

		if !repo.WasCalled("Create") {
			t.Error("Create() должен вызывать репозиторий")
		}

		if !cs.WasCalled("GetByID") {
			t.Error("Create() должен вызывать contentService.GetByID")
		}

		if !ns.WasCalled("Notify") {
			t.Error("Create() должен вызывать notificationService.Notify")
		}

		exists, _ := repo.Exists(context.Background(), userID, contentID)
		if !exists {
			t.Error("Лайк должен быть создан")
		}
	})

	t.Run("Уведомление самому себе должно работать", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		cs.AddSimpleContent(contentID, userID)

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Create() при лайке самому себе error = %v", err)
		}

		if !ns.WasCalled("Notify") {
			t.Error("Notify должен быть вызван даже при лайке самому себе")
		}
	})

	t.Run("Ошибка при уже существующем лайке", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()
		authorID := uuid.New()

		cs.AddSimpleContent(contentID, authorID)
		repo.AddLike(userID, contentID)

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при уже существующем лайке")
		}

		if err.Error() != "already following" {
			t.Errorf("Create() должен вернуть ошибку 'already following', получил: %v", err)
		}
	})

	t.Run("Ошибка при проверке существования лайка", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()
		authorID := uuid.New()

		cs.AddSimpleContent(contentID, authorID)
		repo.SetError("Exists", errors.New("database error"))

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке проверки существования")
		}
	})

	t.Run("Ошибка при получении контента", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		cs.SetError("GetByID", errors.New("content not found"))

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке получения контента")
		}
	})

	t.Run("Ошибка при создании лайка в репозитории", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()
		authorID := uuid.New()

		cs.AddSimpleContent(contentID, authorID)
		repo.SetError("Create", errors.New("database error"))

		service := createTestLikeService(repo, ns, cs)

		err := service.Create(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке создания в репозитории")
		}
	})
}

func TestLikeService_Delete(t *testing.T) {
	t.Run("Успешное удаление лайка", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		repo.AddLike(userID, contentID)
		service := createTestLikeService(repo, ns, cs)

		err := service.Delete(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Delete() error = %v, ожидалось nil", err)
		}

		if !repo.WasCalled("Delete") {
			t.Error("Delete() должен вызывать репозиторий")
		}

		exists, _ := repo.Exists(context.Background(), userID, contentID)
		if exists {
			t.Error("Лайк должен быть удален")
		}
	})

	t.Run("Ошибка при удалении лайка в репозитории", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		repo.AddLike(userID, contentID)
		repo.SetError("Delete", errors.New("database error"))

		service := createTestLikeService(repo, ns, cs)

		err := service.Delete(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку при ошибке удаления в репозитории")
		}
	})
}

func TestLikeService_Exists(t *testing.T) {
	t.Run("Лайк существует", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		repo.AddLike(userID, contentID)
		service := createTestLikeService(repo, ns, cs)

		exists, err := service.Exists(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Exists() error = %v, ожидалось nil", err)
		}
		if !exists {
			t.Error("Exists() должен вернуть true для существующего лайка")
		}
	})

	t.Run("Ошибка при проверке существования в репозитории", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		repo.SetError("Exists", errors.New("database error"))

		service := createTestLikeService(repo, ns, cs)

		_, err := service.Exists(context.Background(), userID, contentID)
		if err == nil {
			t.Error("Exists() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestLikeService_GetLikesCount(t *testing.T) {
	t.Run("Получение количества лайков", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		contentID := uuid.New()
		expectedCount := 5

		repo.SetLikesCount(contentID, expectedCount)
		service := createTestLikeService(repo, ns, cs)

		count, err := service.GetLikesCount(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetLikesCount() error = %v, ожидалось nil", err)
		}
		if count != expectedCount {
			t.Errorf("GetLikesCount() = %v, ожидалось %v", count, expectedCount)
		}
	})

	t.Run("Ошибка при получении количества лайков в репозитории", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		contentID := uuid.New()

		repo.SetError("GetLikesCount", errors.New("database error"))

		service := createTestLikeService(repo, ns, cs)

		_, err := service.GetLikesCount(context.Background(), contentID)
		if err == nil {
			t.Error("GetLikesCount() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestLikeService_EdgeCases(t *testing.T) {
	t.Run("Проверка существования после удаления", func(t *testing.T) {
		repo := NewTestLikeRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()

		userID := uuid.New()
		contentID := uuid.New()

		repo.AddLike(userID, contentID)
		service := createTestLikeService(repo, ns, cs)

		exists, err := service.Exists(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Exists() до удаления error = %v", err)
		}
		if !exists {
			t.Error("Лайк должен существовать до удаления")
		}

		err = service.Delete(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		exists, err = service.Exists(context.Background(), userID, contentID)
		if err != nil {
			t.Errorf("Exists() после удаления error = %v", err)
		}
		if exists {
			t.Error("Лайк не должен существовать после удаления")
		}
	})
}
