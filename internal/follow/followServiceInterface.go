package follow

import (
	"context"

	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
)

type FollowService interface {
	Create(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error
	Delete(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) error
	Exists(ctx context.Context, followerId uuid.UUID, followingId uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, userId uuid.UUID) ([]*user.User, error)
	GetFollowing(ctx context.Context, userId uuid.UUID) ([]*user.User, error)
	GetFollowersCount(ctx context.Context, userId uuid.UUID) (int, error)
	GetFollowingCount(ctx context.Context, userId uuid.UUID) (int, error)
}
