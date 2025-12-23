package content

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type contentService struct {
	repo ContentRepository
}

func NewContentService(repo ContentRepository) ContentService {
	return &contentService{repo: repo}
}

func (s *contentService) Create(ctx context.Context, dto ContentDTO, authorId uuid.UUID) error {
	content := &Content{
		ID:          uuid.New(),
		AuthorId:    authorId,
		ContentType: dto.ContentType,
		Description: dto.Description,
		MediaURLs:   []string{},
		CreatedAt:   time.Now(),
		IsVisible:   true,
	}
	return s.repo.Create(ctx, content)
}

func (s *contentService) GetByID(ctx context.Context, id uuid.UUID) (*Content, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *contentService) GetByType(ctx context.Context, ct ContentType) ([]Content, error) {
	return s.repo.GetByType(ctx, ct)
}

func (s *contentService) GetByAuthorAndType(ctx context.Context, authorId uuid.UUID, ct ContentType) ([]Content, error) {
	return s.repo.GetByAuthorAndType(ctx, authorId, ct)
}

func (s *contentService) Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("content not found")
	}
	if c.AuthorId != userId {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *contentService) Update(ctx context.Context, dto UpdatePostDTO, id uuid.UUID, userId uuid.UUID) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("content not found")
	}
	if c.AuthorId != userId {
		return errors.New("forbidden")
	}

	if dto.MediaURLs != nil {
		if len(dto.MediaURLs) < 1 {
			return errors.New("content must have at least one media")
		}
		newSet := make(map[string]struct{}, len(dto.MediaURLs))
		for _, m := range dto.MediaURLs {
			newSet[m] = struct{}{}
		}

		var toDelete []string
		for _, old := range c.MediaURLs {
			if _, ok := newSet[old]; !ok {
				toDelete = append(toDelete, old)
			}
		}

		for _, url := range toDelete {
			os.Remove("." + url)
		}
	}
	return s.repo.Update(ctx, dto, id)
}

func (s *contentService) UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error {
	c, err := s.repo.GetByID(ctx, contentId)
	if err != nil {
		return errors.New("content not found")
	}
	if c.AuthorId != userId {
		return errors.New("forbidden")
	}

	var allowedImageExt = map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	}

	var allowedVideoExt = map[string]bool{
		".mp4": true, ".mov": true, ".webm": true,
	}

	dir := fmt.Sprintf("./uploads/content/%s", contentId)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.New("cannot create directory")
	}

	var urls []string
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			continue
		}
		defer file.Close()
		ext := filepath.Ext(header.Filename)

		extLow := strings.ToLower(ext)
		if !allowedImageExt[extLow] && !allowedVideoExt[extLow] {
			return fmt.Errorf("unsupported file type: %s", ext)
		}

		fileId := uuid.New().String()
		filename := fileId + ext
		path := filepath.Join(dir, filename)
		out, err := os.Create(path)
		if err != nil {
			continue
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			return errors.New("cannot save image")
		}

		urls = append(urls, fmt.Sprintf("/uploads/content/%s/%s", contentId, filename))
	}

	dto := UpdatePostDTO{
		MediaURLs: append(c.MediaURLs, urls...),
	}

	return s.repo.Update(ctx, dto, contentId)
}
