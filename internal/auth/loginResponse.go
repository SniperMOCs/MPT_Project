package auth

// loginResponse представляет ответ с токеном доступа
// @Description Ответ после успешной аутентификации
type loginResponse struct {
	Token string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT токен доступа
}
