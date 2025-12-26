package user

// UserDTO представляет данные для обновления профиля пользователя
// @Description Данные для обновления профиля
type UserDTO struct {
	Username    string `json:"username,omitempty" example:"john_doe"`     // Имя пользователя (опционально)
	FullName    string `json:"full_name,omitempty" example:"John Doe"`    // Полное имя (опционально)
	Description string `json:"description,omitempty" example:"Developer"` // Описание профиля (опционально)
}
