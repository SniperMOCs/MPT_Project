package comment

import (
	"time"

	"github.com/google/uuid"
)

// Comment представляет модель комментария
// @Description Информация о комментарии
type Comment struct {
	ID        uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`                  // Уникальный идентификатор комментария
	AuthorId  uuid.UUID  `json:"author_id" example:"550e8400-e29b-41d4-a716-446655440000"`           // ID автора комментария
	ParentId  *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // ID родительского комментария (для ответов, опционально)
	ContentId uuid.UUID  `json:"content_id" example:"550e8400-e29b-41d4-a716-446655440000"`          // ID поста
	Text      string     `json:"text" example:"Great post!"`                                         // Текст комментария
	CreatedAt time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`                          // Дата создания комментария
}
