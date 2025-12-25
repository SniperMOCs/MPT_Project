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
