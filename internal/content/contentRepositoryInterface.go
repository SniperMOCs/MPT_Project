package content

import (
	"context"

	"github.com/google/uuid"
)

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*Content, error)
	GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]Content, error)
	GetByFollowing(ctx context.Context, userId uuid.UUID) ([]Content, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID) error
}
