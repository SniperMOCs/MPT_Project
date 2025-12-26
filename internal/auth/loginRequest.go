package auth

// loginRequest представляет данные для входа пользователя
// @Description Данные для аутентификации пользователя
type loginRequest struct {
	Username string `json:"username" example:"john_doe"`    // Имя пользователя
	Password string `json:"password" example:"password123"` // Пароль пользователя
}
