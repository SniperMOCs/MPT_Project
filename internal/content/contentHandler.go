package content

import (
	"encoding/json"
	"net/http"
	"strings"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type ContentHandler struct {
	contentService ContentService
}

func NewContentHandler(contentService ContentService) *ContentHandler {
	return &ContentHandler{contentService: contentService}
}

// @Summary      Создать пост
// @Description  Создание нового поста текущим пользователем
// @Tags         Контент
// @Accept       json
// @Security     BearerAuth
// @Param        dto  body      content.ContentDTO  true  "Данные поста"
// @Success      201  "Пост создан"
// @Failure      400  {string}  string              "Неверный формат запроса"
// @Failure      401  {string}  string              "Не авторизован"
// @Failure      500  {string}  string              "Ошибка сервера"
// @Router       /content/create [post]
func (h *ContentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	var dto ContentDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.contentService.Create(r.Context(), dto.Description, userID); err != nil {
		http.Error(w, "cannot create content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary      Получить пост по ID
// @Description  Возвращает информацию о посте по его ID
// @Tags         Контент
// @Accept       json
// @Produce      json
// @Param        id   path      string         true  "ID поста (UUID)"
// @Success      200  {object}  content.Content  "Информация о посте"
// @Failure      400  {string}  string         "Неверный ID поста"
// @Failure      404  {string}  string         "Пост не найден"
// @Router       /content/{id} [get]
func (h *ContentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/content/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	c, err := h.contentService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "content not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// @Summary      Получить посты пользователя
// @Description  Возвращает список всех постов указанного пользователя
// @Tags         Контент
// @Accept       json
// @Produce      json
// @Param        id   path      string           true  "ID пользователя (UUID)"
// @Success      200  {array}   content.Content  "Список постов"
// @Failure      400  {string}  string           "Неверный ID пользователя"
// @Failure      500  {string}  string           "Ошибка сервера"
// @Router       /content/user/{id} [get]
func (h *ContentHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/content/user/")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	list, err := h.contentService.GetByAuthor(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}

// @Summary      Удалить пост
// @Description  Удаление поста текущим пользователем (только свои посты)
// @Tags         Контент
// @Accept       json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID поста (UUID)"
// @Success      200  "Пост удален"
// @Failure      400  {string}  string  "Неверный ID поста или нет прав на удаление"
// @Failure      401  {string}  string  "Не авторизован"
// @Router       /content/delete/{id} [delete]
func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/content/delete/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := h.contentService.Delete(r.Context(), id, userID); err != nil {
		http.Error(w, "cannot delete content", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Обновить пост
// @Description  Обновление информации о посте (только свои посты)
// @Tags         Контент
// @Accept       json
// @Security     BearerAuth
// @Param        id   path      string              true  "ID поста (UUID)"
// @Param        dto  body      content.UpdatePostDTO  true  "Данные для обновления"
// @Success      200  "Пост обновлен"
// @Failure      400  {string}  string                "Неверный формат запроса"
// @Failure      401  {string}  string                "Не авторизован"
// @Failure      500  {string}  string                "Ошибка сервера"
// @Router       /content/update/{id} [patch]
func (h *ContentHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/content/update/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	var dto UpdatePostDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.contentService.Update(r.Context(), dto.Description, dto.IsVisible, &dto.MediaURLs, id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Загрузить медиа к посту
// @Description  Загрузка медиа-файлов (изображения/видео) к существующему посту
// @Tags         Контент
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        id     path      string  true  "ID поста (UUID)"
// @Param        files  formData  file    true  "Медиа-файлы"
// @Success      201    "Медиа загружено"
// @Failure      400    {string}  string  "Неверный формат запроса или файлы не предоставлены"
// @Failure      401    {string}  string  "Не авторизован"
// @Failure      500    {string}  string  "Ошибка сервера"
// @Router       /content/media/{id} [post]
func (h *ContentHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/content/media/")
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}

	if err := h.contentService.UploadMedia(r.Context(), contentId, userID, files); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
