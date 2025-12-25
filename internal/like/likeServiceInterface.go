package like

import (
	"context"

	"github.com/google/uuid"
)

type LikeService interface {
	Create(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error
	Delete(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) error
	Exists(ctx context.Context, userId uuid.UUID, contentId uuid.UUID) (bool, error)
	GetLikesCount(ctx context.Context, contentId uuid.UUID) (int, error)
}
