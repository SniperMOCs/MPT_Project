package user

// RegisterDTO представляет данные для регистрации пользователя
// @Description Данные для регистрации
type RegisterDTO struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"new_password"`
}
