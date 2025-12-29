package comment

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

// Мок для CommentRepository
type TestCommentRepository struct {
	comments          map[uuid.UUID]*Comment
	commentsByContent map[uuid.UUID][]CommentResponse
	createdComments   []*Comment
	deletedComments   []uuid.UUID
	errors            map[string]error
	callLog           []string
}

// Мок для NotificationService
type TestNotificationService struct {
	notifications []NotificationData
	errors        map[string]error
}

// Мок для ContentService
type TestContentService struct {
	contents map[uuid.UUID]*content.ContentResponse
	errors   map[string]error
}

type NotificationData struct {
	FromUserID uuid.UUID
	ToUserID   uuid.UUID
	ContentID  uuid.UUID
	Type       notification.NotificationType
}

func NewTestCommentRepository() *TestCommentRepository {
	return &TestCommentRepository{
		comments:          make(map[uuid.UUID]*Comment),
		commentsByContent: make(map[uuid.UUID][]CommentResponse),
		errors:            make(map[string]error),
	}
}

func (r *TestCommentRepository) Create(ctx context.Context, c *Comment) error {
	r.callLog = append(r.callLog, "Create")
	if err, ok := r.errors["Create"]; ok {
		return err
	}
	r.comments[c.ID] = c
	r.createdComments = append(r.createdComments, c)
	return nil
}

func (r *TestCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.callLog = append(r.callLog, "Delete")
	if err, ok := r.errors["Delete"]; ok {
		return err
	}
	if _, exists := r.comments[id]; !exists {
		return errors.New("comment not found")
	}
	delete(r.comments, id)
	r.deletedComments = append(r.deletedComments, id)
	return nil
}

func (r *TestCommentRepository) GetByContentId(ctx context.Context, contentID uuid.UUID) ([]CommentResponse, error) {
	r.callLog = append(r.callLog, "GetByContentId")
	if err, ok := r.errors["GetByContentId"]; ok {
		return nil, err
	}
	comments, exists := r.commentsByContent[contentID]
	if !exists {
		return []CommentResponse{}, nil
	}
	return comments, nil
}

func (r *TestCommentRepository) GetById(ctx context.Context, id uuid.UUID) (*Comment, error) {
	r.callLog = append(r.callLog, "GetById")
	if err, ok := r.errors["GetById"]; ok {
		return nil, err
	}
	comment, exists := r.comments[id]
	if !exists {
		return nil, errors.New("comment not found")
	}
	return comment, nil
}

func (r *TestCommentRepository) AddComment(comment *Comment) {
	r.comments[comment.ID] = comment
}

func (r *TestCommentRepository) SetCommentsByContent(contentID uuid.UUID, comments []CommentResponse) {
	r.commentsByContent[contentID] = comments
}

func (r *TestCommentRepository) SetError(methodName string, err error) {
	r.errors[methodName] = err
}

func (r *TestCommentRepository) Reset() {
	r.comments = make(map[uuid.UUID]*Comment)
	r.commentsByContent = make(map[uuid.UUID][]CommentResponse)
	r.createdComments = []*Comment{}
	r.deletedComments = []uuid.UUID{}
	r.errors = make(map[string]error)
	r.callLog = []string{}
}

func (r *TestCommentRepository) WasCalled(methodName string) bool {
	for _, call := range r.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func NewTestNotificationService() *TestNotificationService {
	return &TestNotificationService{
		notifications: []NotificationData{},
		errors:        make(map[string]error),
	}
}

func (n *TestNotificationService) Notify(ctx context.Context, fromUserID, toUserID, contentID uuid.UUID, typ notification.NotificationType) (*notification.NotificationResponse, error) {
	if err, ok := n.errors["Notify"]; ok {
		return nil, err
	}
	n.notifications = append(n.notifications, NotificationData{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		ContentID:  contentID,
		Type:       typ,
	})

	avatarURL := "/uploads/avatar.jpg"
	author := content.ContentAuthor{
		ID:        fromUserID,
		Username:  "testuser",
		AvatarURL: &avatarURL,
	}

	return &notification.NotificationResponse{
		ID:        uuid.New(),
		Recipient: toUserID,
		Type:      typ,
		TargetId:  contentID,
		CreatedAt: time.Now(),
		Read:      false,
		Actor:     author,
	}, nil
}

func (n *TestNotificationService) GetByRecipient(ctx context.Context, userID uuid.UUID) ([]notification.NotificationResponse, error) {
	return nil, nil
}

func (n *TestNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (n *TestNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (n *TestNotificationService) SetError(methodName string, err error) {
	n.errors[methodName] = err
}

func NewTestContentService() *TestContentService {
	return &TestContentService{
		contents: make(map[uuid.UUID]*content.ContentResponse),
		errors:   make(map[string]error),
	}
}

// Исправляем сигнатуру GetByID - теперь возвращает *content.ContentResponse
func (c *TestContentService) GetByID(ctx context.Context, id uuid.UUID) (*content.ContentResponse, error) {
	if err, ok := c.errors["GetByID"]; ok {
		return nil, err
	}
	contentObj, exists := c.contents[id]
	if !exists {
		return nil, errors.New("content not found")
	}
	return contentObj, nil
}

// Метод GetByAuthor
func (c *TestContentService) GetByAuthor(ctx context.Context, authorID uuid.UUID) ([]content.ContentResponse, error) {
	if err, ok := c.errors["GetByAuthor"]; ok {
		return nil, err
	}
	var result []content.ContentResponse
	for _, contentObj := range c.contents {
		if contentObj.Author.ID == authorID {
			result = append(result, *contentObj)
		}
	}
	return result, nil
}

// Метод GetByFollowing
func (c *TestContentService) GetByFollowing(ctx context.Context, userID uuid.UUID) ([]content.ContentResponse, error) {
	if err, ok := c.errors["GetByFollowing"]; ok {
		return nil, err
	}
	return []content.ContentResponse{}, nil
}

// Метод GetAll
func (c *TestContentService) GetAll(ctx context.Context) ([]content.ContentResponse, error) {
	if err, ok := c.errors["GetAll"]; ok {
		return nil, err
	}
	var result []content.ContentResponse
	for _, contentObj := range c.contents {
		result = append(result, *contentObj)
	}
	return result, nil
}

// Добавляем недостающие методы из интерфейса ContentService

func (c *TestContentService) Create(ctx context.Context, description string, authorId uuid.UUID) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (c *TestContentService) Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	return nil
}

func (c *TestContentService) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, userId uuid.UUID) error {
	return nil
}

func (c *TestContentService) UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error {
	return nil
}

// Вспомогательный метод для тестов
func (c *TestContentService) AddTestContentResponse(id uuid.UUID, authorID uuid.UUID, authorUsername string) {
	avatarURL := "/uploads/avatar.jpg"
	contentObj := &content.ContentResponse{
		ID:          id,
		Description: "Test content description",
		MediaURLs:   []string{"/uploads/media1.jpg"},
		CreatedAt:   time.Now(),
		IsVisible:   true,
		Author: content.ContentAuthor{
			ID:        authorID,
			Username:  authorUsername,
			AvatarURL: &avatarURL,
		},
	}
	c.contents[id] = contentObj
}

func (c *TestContentService) SetError(methodName string, err error) {
	c.errors[methodName] = err
}

func createTestCommentService(repo *TestCommentRepository, ns *TestNotificationService, cs *TestContentService) *commentService {
	return &commentService{
		repo: repo,
		ns:   ns,
		cs:   cs,
	}
}

func createTestCommentResponse() CommentResponse {
	parentID := uuid.New()
	avatarURL := "/uploads/avatar.jpg"
	return CommentResponse{
		ID:        uuid.New(),
		ParentId:  &parentID,
		ContentId: uuid.New(),
		Text:      "Test comment response",
		CreatedAt: time.Now(),
		Author: content.ContentAuthor{
			ID:        uuid.New(),
			Username:  "testuser",
			AvatarURL: &avatarURL,
		},
	}
}

func TestCommentService_Create(t *testing.T) {
	t.Run("Успешное создание комментария", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		authorID := uuid.New()
		contentID := uuid.New()
		contentAuthorID := uuid.New()

		// Создаем ContentResponse вместо Content
		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")

		err := service.Create(context.Background(), authorID, nil, contentID, "Test comment")
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		if !repo.WasCalled("Create") {
			t.Error("Create() должен вызывать репозиторий")
		}
	})

	t.Run("Создание комментария с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		repo.SetError("Create", errors.New("database error"))

		contentID := uuid.New()
		contentAuthorID := uuid.New()
		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")

		err := service.Create(context.Background(), uuid.New(), nil, contentID, "Test comment")
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке репозитория")
		}
	})

	t.Run("Создание комментария с ошибкой получения контента", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		cs.SetError("GetByID", errors.New("content not found"))

		err := service.Create(context.Background(), uuid.New(), nil, uuid.New(), "Test comment")
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке получения контента")
		}
	})
}

func TestCommentService_Delete(t *testing.T) {
	t.Run("Успешное удаление комментария", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		commentID := uuid.New()
		authorID := uuid.New()

		comment := &Comment{
			ID:        commentID,
			AuthorId:  authorID,
			ParentId:  nil,
			ContentId: uuid.New(),
			Text:      "Test comment",
			CreatedAt: time.Now(),
		}
		repo.AddComment(comment)

		err := service.Delete(context.Background(), commentID, authorID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		if !repo.WasCalled("Delete") {
			t.Error("Delete() должен вызывать репозиторий")
		}
	})

	t.Run("Удаление чужого комментария", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		commentID := uuid.New()
		authorID := uuid.New()
		differentAuthorID := uuid.New()

		comment := &Comment{
			ID:        commentID,
			AuthorId:  authorID,
			ParentId:  nil,
			ContentId: uuid.New(),
			Text:      "Test comment",
			CreatedAt: time.Now(),
		}
		repo.AddComment(comment)

		err := service.Delete(context.Background(), commentID, differentAuthorID)
		if err == nil {
			t.Error("Delete() должен вернуть ошибку при попытке удалить чужой комментарий")
		}
		if err.Error() != "forbidden" {
			t.Errorf("Delete() должен вернуть ошибку 'forbidden', получил: %v", err)
		}
	})

	t.Run("Удаление несуществующего комментария", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		err := service.Delete(context.Background(), uuid.New(), uuid.New())
		if err == nil {
			t.Error("Delete() должен вернуть ошибку при удалении несуществующего комментария")
		}
		if err.Error() != "comment not found" {
			t.Errorf("Delete() должен вернуть ошибку 'comment not found', получил: %v", err)
		}
	})
}

func TestCommentService_GetByContentId(t *testing.T) {
	t.Run("Успешное получение комментариев", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		contentID := uuid.New()
		expectedComments := []CommentResponse{
			createTestCommentResponse(),
			createTestCommentResponse(),
		}
		repo.SetCommentsByContent(contentID, expectedComments)

		comments, err := service.GetByContentId(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetByContentId() error = %v", err)
		}

		if len(comments) != len(expectedComments) {
			t.Errorf("GetByContentId() вернул %d комментариев, ожидалось %d", len(comments), len(expectedComments))
		}
	})

	t.Run("Получение комментариев для пустого контента", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		contentID := uuid.New()

		comments, err := service.GetByContentId(context.Background(), contentID)
		if err != nil {
			t.Errorf("GetByContentId() error = %v", err)
		}

		if len(comments) != 0 {
			t.Errorf("GetByContentId() должен вернуть пустой список, получил %d комментариев", len(comments))
		}
	})

	t.Run("Получение комментариев с ошибкой репозитория", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		contentID := uuid.New()
		repo.SetError("GetByContentId", errors.New("database error"))

		_, err := service.GetByContentId(context.Background(), contentID)
		if err == nil {
			t.Error("GetByContentId() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestCommentService_Create_WithParent(t *testing.T) {
	t.Run("Создание ответа на комментарий", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		authorID := uuid.New()
		parentID := uuid.New()
		contentID := uuid.New()
		parentAuthorID := uuid.New()
		contentAuthorID := uuid.New()

		parentComment := &Comment{
			ID:        parentID,
			AuthorId:  parentAuthorID,
			ParentId:  nil,
			ContentId: contentID,
			Text:      "Parent comment",
			CreatedAt: time.Now(),
		}
		repo.AddComment(parentComment)

		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")

		err := service.Create(context.Background(), authorID, &parentID, contentID, "Reply comment")
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})

	t.Run("Создание ответа автору контента (без дублирования уведомления)", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		authorID := uuid.New()
		parentID := uuid.New()
		contentID := uuid.New()
		contentAuthorID := authorID // автор контента и автор родительского комментария - один и тот же

		parentComment := &Comment{
			ID:        parentID,
			AuthorId:  contentAuthorID,
			ParentId:  nil,
			ContentId: contentID,
			Text:      "Parent comment",
			CreatedAt: time.Now(),
		}
		repo.AddComment(parentComment)

		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")

		err := service.Create(context.Background(), uuid.New(), &parentID, contentID, "Reply comment")
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
	})

	t.Run("Создание ответа с ошибкой получения родительского комментария", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		contentID := uuid.New()
		parentID := uuid.New()
		contentAuthorID := uuid.New()

		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")
		repo.SetError("GetById", errors.New("parent comment not found"))

		err := service.Create(context.Background(), uuid.New(), &parentID, contentID, "Test comment")
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке получения родительского комментария")
		}
	})
}

func TestCommentService_EdgeCases(t *testing.T) {
	t.Run("Создание комментария с пустым текстом", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		contentID := uuid.New()
		contentAuthorID := uuid.New()
		cs.AddTestContentResponse(contentID, contentAuthorID, "content_author")

		err := service.Create(context.Background(), uuid.New(), nil, contentID, "")
		if err != nil {
			t.Errorf("Create() с пустым текстом не должен возвращать ошибку: %v", err)
		}
	})

	t.Run("Получение комментариев для несуществующего контента", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		cs.SetError("GetByID", errors.New("content not found"))

		contentID := uuid.New()

		err := service.Create(context.Background(), uuid.New(), nil, contentID, "Test comment")
		if err == nil {
			t.Error("Create() должен вернуть ошибку при несуществующем контенте")
		}
	})

	t.Run("Удаление комментария дважды", func(t *testing.T) {
		repo := NewTestCommentRepository()
		ns := NewTestNotificationService()
		cs := NewTestContentService()
		service := createTestCommentService(repo, ns, cs)

		commentID := uuid.New()
		authorID := uuid.New()

		comment := &Comment{
			ID:        commentID,
			AuthorId:  authorID,
			ParentId:  nil,
			ContentId: uuid.New(),
			Text:      "Test comment",
			CreatedAt: time.Now(),
		}
		repo.AddComment(comment)

		err := service.Delete(context.Background(), commentID, authorID)
		if err != nil {
			t.Errorf("Первое удаление Delete() error = %v", err)
		}

		err = service.Delete(context.Background(), commentID, authorID)
		if err == nil {
			t.Error("Второе удаление должно вернуть ошибку")
		}
	})
}
