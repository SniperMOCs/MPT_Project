package comment

import (
	"encoding/json"
	"net/http"
	"strings"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type CommentHandler struct {
	service CommentService
}

func NewCommentHandler(service CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

// @Summary      Создать комментарий
// @Description  Создание нового комментария к посту или ответ на существующий комментарий
// @Tags         Комментарии
// @Accept       json
// @Security     BearerAuth
// @Param        dto  body      comment.CommentDTO  true  "Данные комментария"
// @Success      201  "Комментарий создан"
// @Failure      400  {string}  string              "Неверный формат запроса"
// @Failure      401  {string}  string              "Не авторизован"
// @Router       /comment [post]
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)
	var dto CommentDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), myID, dto.ParentId, dto.ContentId, dto.Text); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary      Удалить комментарий
// @Description  Удаление комментария (только свои комментарии)
// @Tags         Комментарии
// @Accept       json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID комментария (UUID)"
// @Success      200  "Комментарий удален"
// @Failure      400  {string}  string  "Неверный ID комментария или нет прав на удаление"
// @Failure      401  {string}  string  "Не авторизован"
// @Router       /comment/delete/{id} [delete]
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/comment/delete/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		http.Error(w, "cannot delete comment", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Получить комментарии к посту
// @Description  Возвращает список всех комментариев к указанному посту
// @Tags         Комментарии
// @Accept       json
// @Produce      json
// @Param        id   path      string            true  "ID поста (UUID)"
// @Success      200  {array}   comment.Comment  "Список комментариев"
// @Failure      400  {string}  string           "Неверный ID поста"
// @Failure      500  {string}  string           "Ошибка сервера"
// @Router       /comment/content/{id} [get]
func (h *CommentHandler) GetByContentId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/comment/content/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	list, err := h.service.GetByContentId(r.Context(), id)
	if err != nil {
		http.Error(w, "cannot get comments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}
