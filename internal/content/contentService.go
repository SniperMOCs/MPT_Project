package content

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/storage"
	"github.com/google/uuid"
)

type contentService struct {
	repo         ContentRepository
	mediaStorage storage.MediaStorage
}

func NewContentService(repo ContentRepository) ContentService {
	return &contentService{
		repo:         repo,
		mediaStorage: storage.NewFileMediaStorage("./uploads", "/uploads"),
	}
}

func (s *contentService) Create(ctx context.Context, description string, authorId uuid.UUID) error {
	content := &Content{
		ID:          uuid.New(),
		AuthorId:    authorId,
		Description: description,
		MediaURLs:   []string{},
		CreatedAt:   time.Now(),
		IsVisible:   true,
	}
	return s.repo.Create(ctx, content)
}

func (s *contentService) GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *contentService) GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error) {
	return s.repo.GetByAuthor(ctx, authorId)
}

func (s *contentService) GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error) {
	return s.repo.GetByFollowing(ctx, userId)
}

func (s *contentService) GetAll(ctx context.Context) ([]ContentResponse, error) {
	return s.repo.GetAll(ctx)
}

func (s *contentService) Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("content not found")
	}
	if c.Author.ID != userId {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *contentService) Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, userId uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("content not found")
	}
	if c.Author.ID != userId {
		return errors.New("forbidden")
	}

	var toDelete []string
	if mediaURLs != nil {
		if len(*mediaURLs) < 1 {
			return errors.New("content must have at least one media")
		}
		newSet := make(map[string]struct{}, len(*mediaURLs))
		for _, m := range *mediaURLs {
			newSet[m] = struct{}{}
		}

		for _, old := range c.MediaURLs {
			if _, ok := newSet[old]; !ok {
				toDelete = append(toDelete, old)
			}
		}
	}
	if err := s.repo.Update(ctx, description, isVisible, mediaURLs, id); err != nil {
		return err
	}

	if len(toDelete) > 0 {
		_ = s.mediaStorage.Delete(ctx, toDelete)
	}

	return nil
}

func (s *contentService) UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error {
	c, err := s.repo.GetByID(ctx, contentId)
	if err != nil {
		return errors.New("content not found")
	}
	if c.Author.ID != userId {
		return errors.New("forbidden")
	}

	urls, err := s.mediaStorage.Save(ctx, contentId, files)
	if err != nil {
		return err
	}
	c.MediaURLs = append(c.MediaURLs, urls...)

	return s.repo.Update(ctx, &c.Description, &c.IsVisible, &c.MediaURLs, contentId)
}
