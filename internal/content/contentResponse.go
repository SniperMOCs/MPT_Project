package content

import (
	"time"

	"github.com/google/uuid"
)

// ContentAuthor представляет данные для отображения автора
// @Description Данные для отображения пользователя как автора поста, комментария или уведомления
type ContentAuthor struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // ID автора
	Username  string    `json:"username" example:"john_doe"`                       // Username автора
	AvatarURL *string   `json:"avatar_url" example:"/uploads/avatars/john.jpg"`    // URL аватара автора
}

// ContentResponse представляет данные для отображения поста
// @Description Данные для отображения поста и его автора
type ContentResponse struct {
	ID          uuid.UUID     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`        // Уникальный идентификатор поста
	Description string        `json:"description" example:"Beautiful sunset!"`                  // Описание поста
	MediaURLs   []string      `json:"media_urls" example:"[/uploads/content/post1/image1.jpg]"` // Массив URL медиа-файлов
	CreatedAt   time.Time     `json:"created_at" example:"2024-01-01T00:00:00Z"`                // Дата создания поста
	IsVisible   bool          `json:"is_visible" example:"true"`                                // Видимость поста
	Author      ContentAuthor `json:"author"`                                                   // Автор поста
}
