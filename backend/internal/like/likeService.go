package like

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type likeService struct {
	repo LikeRepository
}

func NewLikeService(repo LikeRepository) LikeService {
	return &likeService{repo: repo}
}

func (s *likeService) Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, userId, contentId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already following")
	}

	return s.repo.Create(ctx, userId, contentId)
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
