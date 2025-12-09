package content

import (
	"time"

	"github.com/google/uuid"
)

type ContentType string

const (
	Post  ContentType = "POST"
	Story ContentType = "STORY"
	Clip  ContentType = "CLIP"
)

type Content struct {
	ID          uuid.UUID
	AuthorId    uuid.UUID
	ContentType ContentType
	Description string
	MediaURLs   []string
	CreatedAt   time.Time
	IsVisible   bool
}
