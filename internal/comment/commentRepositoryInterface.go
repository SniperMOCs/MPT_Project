package comment

import (
	"context"

	"github.com/google/uuid"
)

type CommentRepository interface {
	Create(ctx context.Context, c *Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByContentId(ctx context.Context, contentID uuid.UUID) ([]CommentResponse, error)
	GetById(ctx context.Context, id uuid.UUID) (*Comment, error)
}
