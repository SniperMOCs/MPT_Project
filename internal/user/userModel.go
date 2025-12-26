package user

import (
	"time"

	"github.com/google/uuid"
)

// User представляет модель пользователя
// @Description Информация о пользователе
type User struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`          // Уникальный идентификатор пользователя
	Username    string    `json:"username" example:"john_doe"`                                // Имя пользователя
	FullName    *string   `json:"full_name,omitempty" example:"John Doe"`                     // Полное имя пользователя (опционально)
	Password    string    `json:"-"`                                                          // Пароль (не возвращается в ответах)
	AvatarURL   *string   `json:"avatar_url,omitempty" example:"/uploads/avatars/avatar.jpg"` // URL аватара (опционально)
	Description *string   `json:"description,omitempty" example:"Software developer"`         // Описание профиля (опционально)
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`                  // Дата создания аккаунта
}
