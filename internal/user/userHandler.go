package user

import (
	"encoding/json"
	"net/http"
	"strings"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// @Summary      Регистрация нового пользователя
// @Description  Создание нового аккаунта пользователя
// @Tags         Пользователи
// @Accept       json
// @Param        dto  body      user.RegisterDTO  true  "Данные для регистрации"
// @Success      201   "Пользователь успешно создан"
// @Failure      400   {string}  string     "Неверный формат запроса или username уже занят"
// @Router       /users/register [post]
func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	u := &User{
		Username: dto.Username,
		Password: dto.Password,
	}

	if err := h.userService.Create(r.Context(), u); err != nil {
		http.Error(w, "cannot create user", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary      Получить текущего пользователя
// @Description  Возвращает информацию о текущем аутентифицированном пользователе
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  user.User  "Информация о пользователе"
// @Failure      401  {string}  string     "Не авторизован"
// @Failure      404  {string}  string     "Пользователь не найден"
// @Router       /users/me [get]
func (h *UserHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// @Summary      Обновить профиль
// @Description  Обновление информации профиля текущего пользователя
// @Tags         Пользователи
// @Accept       json
// @Security     BearerAuth
// @Param        dto  body      user.UserDTO  true  "Данные для обновления"
// @Success      200  "Профиль обновлен"
// @Failure      400  {string}  string        "Неверный формат запроса"
// @Failure      401  {string}  string        "Не авторизован"
// @Router       /users/update [patch]
func (h *UserHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	var dto UserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.Update(r.Context(), userID, dto.Username, dto.FullName, dto.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Изменить пароль
// @Description  Изменение пароля текущего пользователя
// @Tags         Пользователи
// @Accept       json
// @Security     BearerAuth
// @Param        dto  body      user.UpdatePasswordDTO  true  "Старый и новый пароль"
// @Success      204  "Пароль успешно изменен"
// @Failure      400  {string}  string                  "Неверный формат запроса или неверный старый пароль"
// @Failure      401  {string}  string                  "Не авторизован"
// @Router       /users/setPassword [patch]
func (h *UserHandler) HandleSetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	var dto UpdatePasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdatePassword(r.Context(), userID, dto.OldPassword, dto.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Получить пользователя по ID
// @Description  Возвращает информацию о пользователе по его ID
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        id   path      string      true  "ID пользователя (UUID)"
// @Success      200  {object}  user.User   "Информация о пользователе"
// @Failure      400  {string}  string      "Неверный ID пользователя"
// @Failure      404  {string}  string      "Пользователь не найден"
// @Router       /users/id/{id} [get]
func (h *UserHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/id/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	u, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// @Summary      Получить пользователя по username
// @Description  Возвращает информацию о пользователе по его username
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        username  path      string      true  "Username пользователя"
// @Success      200       {object}  user.User   "Информация о пользователе"
// @Failure      400       {string}  string      "Username не указан"
// @Failure      404       {string}  string      "Пользователь не найден"
// @Router       /users/username/{username} [get]
func (h *UserHandler) HandleGetByUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimPrefix(r.URL.Path, "/api/users/username/")

	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	u, err := h.userService.GetByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// @Summary      Загрузить аватар
// @Description  Загрузка аватара для текущего пользователя
// @Tags         Пользователи
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        file  formData  file    true  "Файл изображения"
// @Success      200   "Аватар загружен"
// @Failure      400   {string}  string  "Неверный файл"
// @Failure      401   {string}  string  "Не авторизован"
// @Failure      500   {string}  string  "Ошибка сервера"
// @Router       /users/setAvatar [post]
func (h *UserHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = h.userService.UpdateAvatar(r.Context(), userID, file); err != nil {
		http.Error(w, "cannot update avatar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Поиск пользователей по fullname и username
// @Description  Поиск пользователей по полному имени или username. Возвращает список пользователей, у которых в fullname или username содержится хотя бы одно из указанных слов
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        q     query     string      true  "Строка поиска (слова разделяются пробелами)"
// @Success      200   {array}   user.User   "Список найденных пользователей"
// @Failure      400   {string}  string       "Неверный формат запроса"
// @Failure      500   {string}  string       "Ошибка сервера"
// @Router       /users/search [get]
func (h *UserHandler) HandleSearchByFullName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		http.Error(w, "search query is required", http.StatusBadRequest)
		return
	}

	users, err := h.userService.SearchByFullName(r.Context(), searchQuery)
	if err != nil {
		http.Error(w, "cannot search users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
