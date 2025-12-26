package follow

import (
	"context"
	"errors"

	"github.com/SniperMOCs/MPT_Project/internal/notification"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
)

type followService struct {
	repo                FollowRepository
	notificationService notification.NotificationService
}

func NewFollowService(repo FollowRepository, notificationService notification.NotificationService) FollowService {
	return &followService{repo: repo, notificationService: notificationService}
}

func (s *followService) Create(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	if followerId == followingId {
		return errors.New("cannot follow yourself")
	}

	exists, err := s.repo.Exists(ctx, followerId, followingId)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already following")
	}

	s.notificationService.Notify(ctx, followerId, followingId, uuid.Nil, notification.Follow)
	return s.repo.Create(ctx, followerId, followingId)
}

func (s *followService) Delete(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error {
	return s.repo.Delete(ctx, followerId, followingId)
}

func (s *followService) Exists(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) (bool, error) {
	return s.repo.Exists(ctx, followerId, followingId)
}

func (s *followService) GetFollowers(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	return s.repo.GetFollowers(ctx, userId)
}

func (s *followService) GetFollowing(ctx context.Context, userId uuid.UUID) ([]*user.User, error) {
	return s.repo.GetFollowing(ctx, userId)
}

func (s *followService) GetFollowersCount(ctx context.Context, userId uuid.UUID) (int, error) {
	return s.repo.GetFollowersCount(ctx, userId)
}

func (s *followService) GetFollowingCount(ctx context.Context, userId uuid.UUID) (int, error) {
	return s.repo.GetFollowingCount(ctx, userId)
}
