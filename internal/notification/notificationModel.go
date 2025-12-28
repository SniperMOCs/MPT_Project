package notification

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType представляет тип уведомления
type NotificationType string

const (
	Like     NotificationType = "LIKE"     // Уведомление о лайке
	Comment  NotificationType = "COMMENT"  // Уведомление о комментарии
	Follow   NotificationType = "FOLLOW"   // Уведомление о подписке
	Response NotificationType = "RESPONSE" // Уведомление об ответе на комментарий
)

// Notification представляет модель уведомления
// @Description Информация об уведомлении
type Notification struct {
	ID        uuid.UUID        `json:"ID" example:"550e8400-e29b-41d4-a716-446655440000"`        // Уникальный идентификатор уведомления
	Recipient uuid.UUID        `json:"Recipient" example:"550e8400-e29b-41d4-a716-446655440000"` // ID получателя уведомления
	Actor     uuid.UUID        `json:"Actor" example:"550e8400-e29b-41d4-a716-446655440000"`     // ID пользователя, который выполнил действие
	Type      NotificationType `json:"Type" example:"LIKE" enums:"LIKE,COMMENT,FOLLOW"`          // Тип уведомления
	TargetId  uuid.UUID        `json:"TargetId" example:"550e8400-e29b-41d4-a716-446655440000"`  // ID целевого объекта (пост, комментарий и т.д.)
	CreatedAt time.Time        `json:"CreatedAt" example:"2024-01-01T00:00:00Z"`                 // Дата создания уведомления
	Read      bool             `json:"Read" example:"false"`                                     // Прочитано ли уведомление
}
