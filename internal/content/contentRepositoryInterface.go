package content

import (
	"context"

	"github.com/google/uuid"
)

type ContentRepository interface {
	Create(ctx context.Context, content *Content) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error)
	GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error)
	GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error)
	GetAll(ctx context.Context) ([]ContentResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID) error
}
