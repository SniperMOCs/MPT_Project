package like

import (
	"context"
	"errors"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/google/uuid"
)

type likeService struct {
	repo                LikeRepository
	notificationService notification.NotificationService
	contentService      content.ContentService
}

func NewLikeService(repo LikeRepository, ns notification.NotificationService, cs content.ContentService) LikeService {
	return &likeService{
		repo:                repo,
		notificationService: ns,
		contentService:      cs,
	}
}

func (s *likeService) Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, userId, contentId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already following")
	}

	c, err := s.contentService.GetByID(ctx, contentId)
	if err != nil {
		return err
	}

	if err := s.repo.Create(ctx, userId, contentId); err != nil {
		return err
	}
	s.notificationService.Notify(ctx, userId, c.Author.ID, c.ID, notification.Like)
	return nil
}

func (s *likeService) Delete(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	return s.repo.Delete(ctx, userId, contentId)
}

func (s *likeService) Exists(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, userId, contentId)
}

func (s *likeService) GetLikesCount(ctx context.Context, contentId uuid.UUID) (int, error) {
	return s.repo.GetLikesCount(ctx, contentId)
}
