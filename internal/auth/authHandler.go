package auth

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// @Summary      Вход в систему
// @Description  Аутентификация пользователя и получение JWT токена
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        request  body      loginRequest  true  "Данные для входа"
// @Success      200      {object}  loginResponse  "Токен доступа"
// @Failure      400      {string}  string        "Неверный формат запроса"
// @Failure      401      {string}  string        "Неверные учетные данные"
// @Router       /users/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password requiered", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{Token: token})
}
