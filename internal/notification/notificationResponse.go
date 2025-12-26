package notification

import (
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/content"
	"github.com/google/uuid"
)

// NotificationResponse представляет данные для отображения уведомлений
// @Description Данные для отображения уведомления и его автора
type NotificationResponse struct {
	ID        uuid.UUID             `json:"ID" example:"550e8400-e29b-41d4-a716-446655440000"`        // Уникальный идентификатор уведомления
	Recipient uuid.UUID             `json:"Recipient" example:"550e8400-e29b-41d4-a716-446655440000"` // ID получателя уведомления
	Type      NotificationType      `json:"Type" example:"LIKE" enums:"LIKE,COMMENT,FOLLOW"`          // Тип уведомления
	TargetId  uuid.UUID             `json:"TargetId" example:"550e8400-e29b-41d4-a716-446655440000"`  // ID целевого объекта (пост, комментарий и т.д.)
	CreatedAt time.Time             `json:"CreatedAt" example:"2024-01-01T00:00:00Z"`                 // Дата создания уведомления
	Read      bool                  `json:"Read" example:"false"`                                     // Прочитано ли уведомление
	Actor     content.ContentAuthor `json:"author"`                                                   // Пользователь, действия которого вызвали уведомление
}
