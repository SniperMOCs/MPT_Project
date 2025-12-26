package like

import (
	"encoding/json"
	"net/http"
	"strings"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type LikeHandler struct {
	likeService LikeService
}

func NewLikeHandler(likeService LikeService) *LikeHandler {
	return &LikeHandler{likeService: likeService}
}

// @Summary      Поставить лайк
// @Description  Постановка лайка текущим пользователем на указанный пост
// @Tags         Лайки
// @Accept       json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID поста (UUID)"
// @Success      201  "Лайк поставлен"
// @Failure      400  {string}  string  "Неверный ID поста"
// @Failure      401  {string}  string  "Не авторизован"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /like/{id} [post]
func (h *LikeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)
	// /api/like/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/like/")
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := h.likeService.Create(r.Context(), myID, contentId); err != nil {
		http.Error(w, "cannot like", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary      Убрать лайк
// @Description  Удаление лайка текущего пользователя с указанного поста
// @Tags         Лайки
// @Accept       json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID поста (UUID)"
// @Success      200  "Лайк удален"
// @Failure      400  {string}  string  "Неверный ID поста"
// @Failure      401  {string}  string  "Не авторизован"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /unlike/{id} [delete]
func (h *LikeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	// /api/unlike/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/unlike/")
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := h.likeService.Delete(r.Context(), myID, contentId); err != nil {
		http.Error(w, "cannot unlike", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary      Проверить лайк
// @Description  Проверяет, лайкнул ли текущий пользователь указанный пост
// @Tags         Лайки
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID поста (UUID)"
// @Success      200  {boolean} bool    "Результат проверки"
// @Failure      400  {string}  string  "Неверный ID поста"
// @Failure      401  {string}  string  "Не авторизован"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /liked/{id} [get]
func (h *LikeHandler) Exists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/liked/")
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	isLiked, err := h.likeService.Exists(r.Context(), myID, contentId)
	if err != nil {
		http.Error(w, "cannot check like", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(isLiked)
}

// @Summary      Получить количество лайков
// @Description  Возвращает количество лайков указанного поста
// @Tags         Лайки
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "ID поста (UUID)"
// @Success      200  {integer} int     "Количество лайков"
// @Failure      400  {string}  string  "Неверный ID поста"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /likes/{id} [get]
func (h *LikeHandler) GetLikesCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/likes/")
	contentId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	likeCount, err := h.likeService.GetLikesCount(r.Context(), contentId)
	if err != nil {
		http.Error(w, "cannot get followers count", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(likeCount)
}
