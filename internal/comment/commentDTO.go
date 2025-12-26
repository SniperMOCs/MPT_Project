package comment

import "github.com/google/uuid"

// CommentDTO представляет данные для создания комментария
// @Description Данные для создания нового комментария
type CommentDTO struct {
	ParentId  *uuid.UUID `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"` // ID родительского комментария (для ответов, опционально)
	ContentId uuid.UUID  `json:"content_id" example:"550e8400-e29b-41d4-a716-446655440000"`          // ID поста
	Text      string     `json:"text" example:"Great post!"`                                         // Текст комментария
}
