package storage

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
)

type MediaStorage interface {
	Delete(ctx context.Context, urls []string) error
	Save(ctx context.Context, contentId uuid.UUID, files []*multipart.FileHeader) ([]string, error)
}
