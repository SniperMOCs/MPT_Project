package comment

import (
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/google/uuid"
)

// CommentResponse представляет данные для отображения комментариев
// @Description Данные для отображения комментария и его автора
type CommentResponse struct {
	ID        uuid.UUID             `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`                  // Уникальный идентификатор комментария
	ParentId  *uuid.UUID            `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // ID родительского комментария (для ответов, опционально)
	ContentId uuid.UUID             `json:"content_id" example:"550e8400-e29b-41d4-a716-446655440000"`          // ID поста
	Text      string                `json:"text" example:"Great post!"`                                         // Текст комментария
	CreatedAt time.Time             `json:"created_at" example:"2024-01-01T00:00:00Z"`                          // Дата создания комментария
	Author    content.ContentAuthor `json:"author"`                                                             // Автор комментария
}
