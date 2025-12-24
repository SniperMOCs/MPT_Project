package content

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
)

type ContentService interface {
	Create(ctx context.Context, contentType ContentType, description string, authorId uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Content, error)
	GetByType(ctx context.Context, ct ContentType) ([]Content, error)
	GetByAuthorAndType(ctx context.Context, authorId uuid.UUID, ct ContentType) ([]Content, error)
	Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error
	Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, userId uuid.UUID) error
	UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error
}
