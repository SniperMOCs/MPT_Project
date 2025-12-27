package content

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
)

type ContentService interface {
	Create(ctx context.Context, description string, authorId uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ContentResponse, error)
	GetByAuthor(ctx context.Context, authorId uuid.UUID) ([]ContentResponse, error)
	GetByFollowing(ctx context.Context, userId uuid.UUID) ([]ContentResponse, error)
	GetAll(ctx context.Context) ([]ContentResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error
	Update(ctx context.Context, description *string, isVisible *bool, mediaURLs *[]string, id uuid.UUID, userId uuid.UUID) error
	UploadMedia(ctx context.Context, contentId uuid.UUID, userId uuid.UUID, files []*multipart.FileHeader) error
}
