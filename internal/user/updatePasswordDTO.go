package user

// UpdatePasswordDTO представляет данные для изменения пароля
// @Description Данные для изменения пароля пользователя
type UpdatePasswordDTO struct {
	OldPassword string `json:"old_password" example:"oldpass123"` // Текущий пароль
	NewPassword string `json:"new_password" example:"newpass123"` // Новый пароль
}
