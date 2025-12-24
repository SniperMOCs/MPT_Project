package follow

import (
	"encoding/json"
	"net/http"
	"strings"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

type FollowHandler struct {
	followService FollowService
}

func NewFollowHandler(followService FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

func (h *FollowHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	// /api/follow/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/follow/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.followService.Create(r.Context(), myID, userId); err != nil {
		http.Error(w, "cannot follow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *FollowHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	// /api/unfollow/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/unfollow/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.followService.Delete(r.Context(), myID, userId); err != nil {
		http.Error(w, "cannot unfollow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// /api/followers/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/followers/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	list, err := h.followService.GetFollowers(r.Context(), userId)
	if err != nil {
		http.Error(w, "cannot get followers", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}
func (h *FollowHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// /api/following/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/following/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	list, err := h.followService.GetFollowing(r.Context(), userId)
	if err != nil {
		http.Error(w, "cannot get following", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}

func (h *FollowHandler) FollowInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/follow/info/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	followerCount, err := h.followService.GetFollowersCount(r.Context(), userId)
	if err != nil {
		http.Error(w, "cannot get followers count", http.StatusInternalServerError)
		return
	}

	followingCount, err := h.followService.GetFollowingCount(r.Context(), userId)
	if err != nil {
		http.Error(w, "cannot get following count", http.StatusInternalServerError)
		return
	}

	var info FollowInfoDTO
	info.FollowersCount = followerCount
	info.FollowingCount = followingCount

	json.NewEncoder(w).Encode(info)
}

func (h *FollowHandler) IsFollower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	myID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/follower/")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	isFollower, err := h.followService.Exists(r.Context(), myID, userId)
	if err != nil {
		http.Error(w, "cannot check follow", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(isFollower)
}
