package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type AvatarStorage interface {
	Save(ctx context.Context, userID uuid.UUID, r io.Reader) (string, error)
}
