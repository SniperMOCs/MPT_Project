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

	if dto.ContentType == "" {
		http.Error(w, "content_type requiered", http.StatusBadRequest)
		return
	}

	if err := h.contentService.Create(r.Context(), dto, userID); err != nil {
		http.Error(w, "cannot create content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

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

func (h *ContentHandler) GetByType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := ContentType(r.URL.Query().Get("type"))
	if contentType == "" {
		http.Error(w, "content type requiered", http.StatusNotFound)
		return
	}

	list, err := h.contentService.GetByType(r.Context(), contentType)
	if err != nil {
		http.Error(w, "cannot get content", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}

func (h *ContentHandler) GetByAuthorAndType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// /api/content/user/{id}?type=POST or type=CLIP or type=STORY
	url_parts := strings.Split(r.URL.Path, "/")
	if len(url_parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(url_parts[4])
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	contentType := ContentType(r.URL.Query().Get("type"))
	if contentType == "" {
		http.Error(w, "content type requiered", http.StatusNotFound)
		return
	}

	list, err := h.contentService.GetByAuthorAndType(r.Context(), userID, contentType)
	if err != nil {
		http.Error(w, "cannot get content", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}

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

func (h *ContentHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	if err := h.contentService.Update(r.Context(), dto, id, userID); err != nil {
		http.Error(w, "cannot update content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContentHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(contextkeys.UserIDKey).(uuid.UUID)

	// /api/content/media/{id}
	url_parts := strings.Split(r.URL.Path, "/")
	if len(url_parts) < 5 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	contentId, err := uuid.Parse(url_parts[4])
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
