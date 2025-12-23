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
	ID          uuid.UUID   `json:"id"`
	AuthorId    uuid.UUID   `json:"author_id"`
	ContentType ContentType `json:"content_type"`
	Description string      `json:"description"`
	MediaURLs   []string    `json:"media_urls"`
	CreatedAt   time.Time   `json:"created_at"`
	IsVisible   bool        `json:"is_visible"`
}
