package follow

import (
	"time"

	"github.com/google/uuid"
)

// Follow представляет модель подписки
// @Description Информация о подписке
type Follow struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`           // Уникальный идентификатор подписки
	FollowerId  uuid.UUID `json:"follower_id" example:"550e8400-e29b-41d4-a716-446655440000"`  // ID подписчика
	FollowingId uuid.UUID `json:"following_id" example:"550e8400-e29b-41d4-a716-446655440000"` // ID пользователя, на которого подписались
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`                   // Дата создания подписки
}
