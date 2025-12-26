package content

import (
	"time"

	"github.com/google/uuid"
)

// Content представляет модель поста
// @Description Описание поста
type Content struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`        // Уникальный идентификатор поста
	AuthorId    uuid.UUID `json:"author_id" example:"550e8400-e29b-41d4-a716-446655440000"` // ID автора поста
	Description string    `json:"description" example:"Beautiful sunset!"`                  // Описание поста
	MediaURLs   []string  `json:"media_urls" example:"[/uploads/content/post1/image1.jpg]"` // Массив URL медиа-файлов
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`                // Дата создания поста
	IsVisible   bool      `json:"is_visible" example:"true"`                                // Видимость поста
}
