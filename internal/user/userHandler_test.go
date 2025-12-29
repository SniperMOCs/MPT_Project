package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
)

// Тестовый сервис для обработчиков
type TestService struct {
	createFunc           func(ctx context.Context, u *User) error
	getByIDFunc          func(ctx context.Context, id uuid.UUID) (*User, error)
	getByUsernameFunc    func(ctx context.Context, username string) (*User, error)
	updateFunc           func(ctx context.Context, id uuid.UUID, username, fullname, description string) error
	updatePasswordFunc   func(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	updateAvatarFunc     func(ctx context.Context, id uuid.UUID, file io.Reader) error
	searchByFullNameFunc func(ctx context.Context, searchQuery string) ([]*User, error)
}

func (s *TestService) Create(ctx context.Context, u *User) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, u)
	}
	return nil
}

func (s *TestService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return &User{ID: id, Username: "testuser"}, nil
}

func (s *TestService) GetByUsername(ctx context.Context, username string) (*User, error) {
	if s.getByUsernameFunc != nil {
		return s.getByUsernameFunc(ctx, username)
	}
	return &User{ID: uuid.New(), Username: username}, nil
}

func (s *TestService) Update(ctx context.Context, id uuid.UUID, username, fullname, description string) error {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, id, username, fullname, description)
	}
	return nil
}

func (s *TestService) UpdatePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	if s.updatePasswordFunc != nil {
		return s.updatePasswordFunc(ctx, id, oldPassword, newPassword)
	}
	return nil
}

func (s *TestService) UpdateAvatar(ctx context.Context, id uuid.UUID, file io.Reader) error {
	if s.updateAvatarFunc != nil {
		return s.updateAvatarFunc(ctx, id, file)
	}
	return nil
}

func (s *TestService) SearchByFullName(ctx context.Context, searchQuery string) ([]*User, error) {
	if s.searchByFullNameFunc != nil {
		return s.searchByFullNameFunc(ctx, searchQuery)
	}
	return []*User{}, nil
}

func (s *TestService) CheckPassword(u *User, password string) bool {
	return true
}

func TestHandler_HandleRegister(t *testing.T) {
	t.Run("Успешная регистрация", func(t *testing.T) {
		service := &TestService{
			createFunc: func(ctx context.Context, u *User) error {
				u.ID = uuid.New()
				return nil
			},
		}
		handler := NewUserHandler(service)

		dto := RegisterDTO{
			Username: "testuser",
			Password: "password123",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("HandleRegister() код = %v, ожидался %v", w.Code, http.StatusCreated)
		}
	})

	t.Run("Ошибка при неверном методе запроса", func(t *testing.T) {
		handler := NewUserHandler(&TestService{})

		req := httptest.NewRequest(http.MethodGet, "/users/register", nil)
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleRegister() код = %v, ожидался %v", w.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestHandler_HandleMe(t *testing.T) {
	t.Run("Успешное получение информации о пользователе", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*User, error) {
				return &User{
					ID:       id,
					Username: "testuser",
					Password: "hashedpassword",
				}, nil
			},
		}
		handler := NewUserHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleMe(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleMe() код = %v, ожидался %v", w.Code, http.StatusOK)
		}

		var user User
		if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
			t.Fatalf("HandleMe() неверный JSON ответ: %v", err)
		}

		if user.ID != userID {
			t.Errorf("HandleMe() вернул ID = %v, ожидался %v", user.ID, userID)
		}

		if user.Password != "" {
			t.Error("HandleMe() не должен возвращать пароль")
		}
	})

	t.Run("Ошибка при отсутствии пользователя", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*User, error) {
				return nil, errors.New("user not found")
			},
		}
		handler := NewUserHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleMe(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleMe() код = %v, ожидался %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandler_HandleUpdate(t *testing.T) {
	t.Run("Успешное обновление профиля", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			updateFunc: func(ctx context.Context, id uuid.UUID, username, fullname, description string) error {
				if id != userID {
					return errors.New("wrong user id")
				}
				return nil
			},
		}
		handler := NewUserHandler(service)

		dto := UserDTO{
			Username:    "newusername",
			FullName:    "New Full Name",
			Description: "New Description",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleUpdate() код = %v, ожидался %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Ошибка при неверном JSON", func(t *testing.T) {
		userID := uuid.New()
		handler := NewUserHandler(&TestService{})

		req := httptest.NewRequest(http.MethodPatch, "/users/update", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleUpdate() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleSetPassword(t *testing.T) {
	t.Run("Успешное изменение пароля", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			updatePasswordFunc: func(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
				if id != userID {
					return errors.New("wrong user id")
				}
				return nil
			},
		}
		handler := NewUserHandler(service)

		dto := UpdatePasswordDTO{
			OldPassword: "oldpass123",
			NewPassword: "newpass123",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/setPassword", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleSetPassword(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("HandleSetPassword() код = %v, ожидался %v", w.Code, http.StatusNoContent)
		}
	})

	t.Run("Ошибка при изменении пароля", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			updatePasswordFunc: func(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
				return errors.New("password error")
			},
		}
		handler := NewUserHandler(service)

		dto := UpdatePasswordDTO{
			OldPassword: "oldpass123",
			NewPassword: "newpass123",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/setPassword", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleSetPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleSetPassword() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleGetByID(t *testing.T) {
	t.Run("Успешное получение пользователя по ID", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*User, error) {
				if id != userID {
					return nil, errors.New("wrong id")
				}
				return &User{
					ID:       id,
					Username: "testuser",
					Password: "hashedpassword",
				}, nil
			},
		}
		handler := NewUserHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/users/id/"+userID.String(), nil)
		w := httptest.NewRecorder()

		handler.HandleGetByID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleGetByID() код = %v, ожидался %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Ошибка при неверном UUID", func(t *testing.T) {
		handler := NewUserHandler(&TestService{})

		req := httptest.NewRequest(http.MethodGet, "/api/users/id/invalid-uuid", nil)
		w := httptest.NewRecorder()

		handler.HandleGetByID(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleGetByID() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleGetByUsername(t *testing.T) {
	t.Run("Успешное получение пользователя по username", func(t *testing.T) {
		username := "testuser"
		service := &TestService{
			getByUsernameFunc: func(ctx context.Context, uname string) (*User, error) {
				if uname != username {
					return nil, errors.New("wrong username")
				}
				return &User{
					ID:       uuid.New(),
					Username: uname,
					Password: "hashedpassword",
				}, nil
			},
		}
		handler := NewUserHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/api/users/username/"+username, nil)
		w := httptest.NewRecorder()

		handler.HandleGetByUsername(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleGetByUsername() код = %v, ожидался %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Ошибка при пустом username", func(t *testing.T) {
		handler := NewUserHandler(&TestService{})

		req := httptest.NewRequest(http.MethodGet, "/api/users/username/", nil)
		w := httptest.NewRecorder()

		handler.HandleGetByUsername(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleGetByUsername() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleUploadAvatar(t *testing.T) {
	t.Run("Успешная загрузка аватара", func(t *testing.T) {
		userID := uuid.New()
		service := &TestService{
			updateAvatarFunc: func(ctx context.Context, id uuid.UUID, file io.Reader) error {
				if id != userID {
					return errors.New("wrong user id")
				}
				data, err := io.ReadAll(file)
				if err != nil {
					return err
				}
				if len(data) == 0 {
					return errors.New("empty file")
				}
				return nil
			},
		}
		handler := NewUserHandler(service)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "avatar.jpg")
		part.Write([]byte("fake image data"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users/setAvatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleUploadAvatar(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleUploadAvatar() код = %v, ожидался %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Ошибка при отсутствии файла", func(t *testing.T) {
		userID := uuid.New()
		handler := NewUserHandler(&TestService{})

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users/setAvatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.HandleUploadAvatar(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleUploadAvatar() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleSearchByFullName(t *testing.T) {
	t.Run("Успешный поиск пользователей", func(t *testing.T) {
		service := &TestService{
			searchByFullNameFunc: func(ctx context.Context, searchQuery string) ([]*User, error) {
				return []*User{
					{ID: uuid.New(), Username: "user1"},
					{ID: uuid.New(), Username: "user2"},
				}, nil
			},
		}
		handler := NewUserHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/users/search?q=john", nil)
		w := httptest.NewRecorder()

		handler.HandleSearchByFullName(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleSearchByFullName() код = %v, ожидался %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Ошибка при отсутствии поискового запроса", func(t *testing.T) {
		handler := NewUserHandler(&TestService{})

		req := httptest.NewRequest(http.MethodGet, "/users/search", nil)
		w := httptest.NewRecorder()

		handler.HandleSearchByFullName(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleSearchByFullName() код = %v, ожидался %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_HandleMe_WrongMethod(t *testing.T) {
	handler := NewUserHandler(&TestService{})

	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/users/me", nil)
			w := httptest.NewRecorder()

			handler.HandleMe(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleMe() с методом %s код = %v, ожидался %v", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHandler_HandleUpdate_WrongMethod(t *testing.T) {
	userID := uuid.New()
	handler := NewUserHandler(&TestService{})

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			dto := UserDTO{Username: "newuser"}
			body, _ := json.Marshal(dto)

			req := httptest.NewRequest(method, "/users/update", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleUpdate(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleUpdate() с методом %s код = %v, ожидался %v", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHandler_HandleSetPassword_WrongMethod(t *testing.T) {
	userID := uuid.New()
	handler := NewUserHandler(&TestService{})

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			dto := UpdatePasswordDTO{
				OldPassword: "old",
				NewPassword: "new",
			}
			body, _ := json.Marshal(dto)

			req := httptest.NewRequest(method, "/users/setPassword", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), contextkeys.UserIDKey, userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleSetPassword(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleSetPassword() с методом %s код = %v, ожидался %v", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
