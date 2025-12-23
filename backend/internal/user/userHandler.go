package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if u.Username == " " || u.Password == " " {
		http.Error(w, "username and password requiered", http.StatusBadRequest)
		return
	}

	if err := h.userService.Create(r.Context(), &u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := uuid.Parse(r.Context().Value("user_id").(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
	}

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := uuid.Parse(r.Context().Value("user_id").(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
	}

	var dto UserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.Update(r.Context(), userID, dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) HandleSetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, err := uuid.Parse(r.Context().Value("user_id").(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
	}

	var dto UpdatePasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if dto.OldPassword == "" || dto.NewPassword == "" {
		http.Error(w, "both passwords requiered", http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdatePassword(r.Context(), userID, dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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

func (h *UserHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userIDstr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
	}

	dir := "./uploads/avatars"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		http.Error(w, "cannot create directory", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path := fmt.Sprintf("%s/%s.png", dir, userIDstr)
	out, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}

	if err = h.userService.UpdateAvatar(r.Context(), userID, path); err != nil {
		http.Error(w, "cannot update avatar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
