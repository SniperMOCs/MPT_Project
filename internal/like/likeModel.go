package like

import (
	"time"

	"github.com/google/uuid"
)

// Like представляет модель лайка
// @Description Информация о лайке
type Like struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`         // Уникальный идентификатор лайка
	UserId    uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`    // ID пользователя, который поставил лайк
	ContentId uuid.UUID `json:"content_id" example:"550e8400-e29b-41d4-a716-446655440000"` // ID поста
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`                 // Дата создания лайка
}
