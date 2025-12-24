package content

import (
	"context"

	"github.com/google/uuid"
)

type ContentRepository interface {
	Create(ctx context.Context, content *Content) error
	GetByType(ctx context.Context, ct ContentType) ([]Content, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Content, error)
	GetByAuthorAndType(ctx context.Context, authorId uuid.UUID, ct ContentType) ([]Content, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID) error
}
