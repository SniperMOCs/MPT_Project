package comment

import (
	"context"

	"github.com/google/uuid"
)

type CommentService interface {
	Create(ctx context.Context, authorId uuid.UUID, parentId *uuid.UUID, contentId uuid.UUID, text string) error
	Delete(ctx context.Context, id uuid.UUID, authorId uuid.UUID) error
	GetByContentId(ctx context.Context, contentID uuid.UUID) ([]Comment, error)
}
