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
	"testing"

	contextkeys "github.com/SniperMOCs/MPT_Project/internal/contextKeys"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TestUserHandler структура для тестов
type TestUserHandler struct {
	handler *UserHandler
	service *TestUserService
}

// TestUserService для тестов хэндлера
type TestUserService struct {
	usersByID       map[uuid.UUID]*User
	usersByUsername map[string]*User
	errors          map[string]error
	callLog         []string
	avatarUpdates   []AvatarUpdateData
	searchResults   []*User
}

func NewTestUserService() *TestUserService {
	return &TestUserService{
		usersByID:       make(map[uuid.UUID]*User),
		usersByUsername: make(map[string]*User),
		errors:          make(map[string]error),
		searchResults:   make([]*User, 0),
	}
}

func (s *TestUserService) Create(ctx context.Context, u *User) error {
	s.callLog = append(s.callLog, "Create")
	if err, ok := s.errors["Create"]; ok {
		return err
	}
	if _, exists := s.usersByUsername[u.Username]; exists {
		return errors.New("username taken")
	}
	s.usersByID[u.ID] = u
	s.usersByUsername[u.Username] = u
	return nil
}

func (s *TestUserService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	s.callLog = append(s.callLog, "GetByID")
	if err, ok := s.errors["GetByID"]; ok {
		return nil, err
	}
	user, exists := s.usersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *TestUserService) GetByUsername(ctx context.Context, username string) (*User, error) {
	s.callLog = append(s.callLog, "GetByUsername")
	if err, ok := s.errors["GetByUsername"]; ok {
		return nil, err
	}
	user, exists := s.usersByUsername[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *TestUserService) SearchByFullName(ctx context.Context, searchQuery string) ([]*User, error) {
	s.callLog = append(s.callLog, "SearchByFullName")
	if err, ok := s.errors["SearchByFullName"]; ok {
		return nil, err
	}
	return s.searchResults, nil
}

func (s *TestUserService) Update(ctx context.Context, id uuid.UUID, username string, fullname string, description string) error {
	s.callLog = append(s.callLog, "Update")
	if err, ok := s.errors["Update"]; ok {
		return err
	}
	user, exists := s.usersByID[id]
	if !exists {
		return errors.New("user not found")
	}
	if username != "" && username != user.Username {
		if _, exists := s.usersByUsername[username]; exists {
			return errors.New("username taken")
		}
		delete(s.usersByUsername, user.Username)
		user.Username = username
		s.usersByUsername[username] = user
	}
	if fullname != "" {
		fullNameCopy := fullname
		user.FullName = &fullNameCopy
	}
	if description != "" {
		descCopy := description
		user.Description = &descCopy
	}
	return nil
}

func (s *TestUserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	s.callLog = append(s.callLog, "UpdatePassword")
	if err, ok := s.errors["UpdatePassword"]; ok {
		return err
	}
	user, exists := s.usersByID[userID]
	if !exists {
		return errors.New("user not found")
	}
	if !s.CheckPassword(user, oldPassword) {
		return errors.New("incorrect password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return nil
}

func (s *TestUserService) CheckPassword(u *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (s *TestUserService) UpdateAvatar(ctx context.Context, id uuid.UUID, file io.Reader) error {
	s.callLog = append(s.callLog, "UpdateAvatar")
	if err, ok := s.errors["UpdateAvatar"]; ok {
		return err
	}
	user, exists := s.usersByID[id]
	if !exists {
		return errors.New("user not found")
	}
	avatarURL := "/uploads/avatar.jpg"
	user.AvatarURL = &avatarURL
	s.avatarUpdates = append(s.avatarUpdates, AvatarUpdateData{
		ID:        id,
		AvatarURL: avatarURL,
	})
	return nil
}

func (s *TestUserService) AddUser(user *User) {
	s.usersByID[user.ID] = user
	s.usersByUsername[user.Username] = user
}

func (s *TestUserService) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestUserService) SetSearchResults(users []*User) {
	s.searchResults = users
}

func (s *TestUserService) Reset() {
	s.usersByID = make(map[uuid.UUID]*User)
	s.usersByUsername = make(map[string]*User)
	s.errors = make(map[string]error)
	s.callLog = []string{}
	s.avatarUpdates = []AvatarUpdateData{}
	s.searchResults = []*User{}
}

func (s *TestUserService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func createTestUserHandler() *TestUserHandler {
	service := NewTestUserService()
	handler := NewUserHandler(service)
	return &TestUserHandler{
		handler: handler,
		service: service,
	}
}

func createTestContext(userID uuid.UUID) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, contextkeys.UserIDKey, userID)
}

func TestUserHandler_HandleRegister(t *testing.T) {
	t.Run("Успешная регистрация пользователя", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		dto := RegisterDTO{
			Username: "newuser",
			Password: "password123",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.handler.HandleRegister(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("HandleRegister() status = %v, want %v", w.Code, http.StatusCreated)
		}
		if !th.service.WasCalled("Create") {
			t.Error("HandleRegister() должен вызывать Create")
		}
	})

	t.Run("Ошибка при неправильном методе", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodGet, "/users/register", nil)
		w := httptest.NewRecorder()

		th.handler.HandleRegister(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleRegister() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка при невалидном JSON", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.handler.HandleRegister(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleRegister() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "invalid request\n" {
			t.Errorf("HandleRegister() body = %v, want %v", w.Body.String(), "invalid request\n")
		}
	})

	t.Run("Ошибка при создании пользователя", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		// Создаем пользователя с таким же username
		existingUser := createTestUser()
		existingUser.Username = "existinguser"
		th.service.AddUser(existingUser)

		dto := RegisterDTO{
			Username: "existinguser",
			Password: "password123",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.handler.HandleRegister(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleRegister() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "cannot create user\n" {
			t.Errorf("HandleRegister() body = %v, want %v", w.Body.String(), "cannot create user\n")
		}
	})
}

func TestUserHandler_HandleMe(t *testing.T) {
	t.Run("Успешное получение текущего пользователя", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleMe(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleMe() status = %v, want %v", w.Code, http.StatusOK)
		}

		var responseUser User
		if err := json.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
			t.Errorf("HandleMe() неверный JSON: %v", err)
		}

		if responseUser.Username != testUser.Username {
			t.Errorf("HandleMe() username = %v, want %v", responseUser.Username, testUser.Username)
		}
		if responseUser.Password != "" {
			t.Error("HandleMe() не должен возвращать пароль")
		}
	})

	t.Run("Ошибка при неправильном методе", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleMe(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleMe() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := createTestContext(nonExistentID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleMe(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleMe() status = %v, want %v", w.Code, http.StatusNotFound)
		}
		if w.Body.String() != "user not found\n" {
			t.Errorf("HandleMe() body = %v, want %v", w.Body.String(), "user not found\n")
		}
	})
}

func TestUserHandler_HandleUpdate(t *testing.T) {
	t.Run("Успешное обновление профиля", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		dto := UserDTO{
			Username:    "newusername",
			FullName:    "New Full Name",
			Description: "New Description",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUpdate(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleUpdate() status = %v, want %v", w.Code, http.StatusOK)
		}
		if !th.service.WasCalled("Update") {
			t.Error("HandleUpdate() должен вызывать Update")
		}
	})

	t.Run("Ошибка при неправильном методе", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/users/update", nil)
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUpdate(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleUpdate() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка при невалидном JSON", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		req := httptest.NewRequest(http.MethodPatch, "/users/update", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleUpdate() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("Ошибка при обновлении", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		// Создаем другого пользователя с username который хотим занять
		otherUser := createTestUser()
		otherUser.ID = uuid.New()
		otherUser.Username = "takenusername"
		th.service.AddUser(otherUser)

		dto := UserDTO{
			Username: "takenusername",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUpdate(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleUpdate() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestUserHandler_HandleSetPassword(t *testing.T) {
	t.Run("Успешная смена пароля", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		dto := UpdatePasswordDTO{
			OldPassword: "password123",
			NewPassword: "newpassword456",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/setPassword", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleSetPassword(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("HandleSetPassword() status = %v, want %v", w.Code, http.StatusNoContent)
		}
		if !th.service.WasCalled("UpdatePassword") {
			t.Error("HandleSetPassword() должен вызывать UpdatePassword")
		}
	})

	t.Run("Ошибка при неправильном методе", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/users/setPassword", nil)
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleSetPassword(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleSetPassword() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка при неверном старом пароле", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		dto := UpdatePasswordDTO{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword456",
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPatch, "/users/setPassword", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleSetPassword(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleSetPassword() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
	})
}

func TestUserHandler_HandleGetByID(t *testing.T) {
	t.Run("Успешное получение пользователя по ID", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		req := httptest.NewRequest(http.MethodGet, "/api/users/id/"+testUserID.String(), nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleGetByID() status = %v, want %v", w.Code, http.StatusOK)
		}

		var responseUser User
		if err := json.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
			t.Errorf("HandleGetByID() неверный JSON: %v", err)
		}

		if responseUser.ID != testUserID {
			t.Errorf("HandleGetByID() ID = %v, want %v", responseUser.ID, testUserID)
		}
		if responseUser.Password != "" {
			t.Error("HandleGetByID() не должен возвращать пароль")
		}
	})

	t.Run("Ошибка при неправильном методе", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/users/id/"+testUserID.String(), nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByID(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("HandleGetByID() status = %v, want %v", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Ошибка при невалидном UUID", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodGet, "/api/users/id/invalid-uuid", nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByID(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleGetByID() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "invalid user id\n" {
			t.Errorf("HandleGetByID() body = %v, want %v", w.Body.String(), "invalid user id\n")
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/users/id/"+nonExistentID.String(), nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByID(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleGetByID() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_HandleGetByUsername(t *testing.T) {
	t.Run("Успешное получение пользователя по username", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUsername := testUser.Username
		th.service.AddUser(testUser)

		req := httptest.NewRequest(http.MethodGet, "/api/users/username/"+testUsername, nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByUsername(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleGetByUsername() status = %v, want %v", w.Code, http.StatusOK)
		}

		var responseUser User
		if err := json.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
			t.Errorf("HandleGetByUsername() неверный JSON: %v", err)
		}

		if responseUser.Username != testUsername {
			t.Errorf("HandleGetByUsername() username = %v, want %v", responseUser.Username, testUsername)
		}
	})

	t.Run("Ошибка при пустом username", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodGet, "/api/users/username/", nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByUsername(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleGetByUsername() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "username required\n" {
			t.Errorf("HandleGetByUsername() body = %v, want %v", w.Body.String(), "username required\n")
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodGet, "/api/users/username/nonexistent", nil)
		w := httptest.NewRecorder()

		th.handler.HandleGetByUsername(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("HandleGetByUsername() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_HandleUploadAvatar(t *testing.T) {
	t.Run("Успешная загрузка аватара", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUser := createTestUser()
		testUserID := testUser.ID
		th.service.AddUser(testUser)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "avatar.jpg")
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users/setAvatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUploadAvatar(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleUploadAvatar() status = %v, want %v", w.Code, http.StatusOK)
		}
		if !th.service.WasCalled("UpdateAvatar") {
			t.Error("HandleUploadAvatar() должен вызывать UpdateAvatar")
		}
	})

	t.Run("Ошибка при отсутствии файла", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users/setAvatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUploadAvatar(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleUploadAvatar() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "invalid file\n" {
			t.Errorf("HandleUploadAvatar() body = %v, want %v", w.Body.String(), "invalid file\n")
		}
	})

	t.Run("Ошибка при обновлении аватара", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		testUserID := uuid.New()
		th.service.SetError("UpdateAvatar", errors.New("storage error"))

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "avatar.jpg")
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/users/setAvatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx := createTestContext(testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		th.handler.HandleUploadAvatar(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("HandleUploadAvatar() status = %v, want %v", w.Code, http.StatusInternalServerError)
		}
		if w.Body.String() != "cannot update avatar\n" {
			t.Errorf("HandleUploadAvatar() body = %v, want %v", w.Body.String(), "cannot update avatar\n")
		}
	})
}

func TestUserHandler_HandleSearchByFullName(t *testing.T) {
	t.Run("Успешный поиск пользователей", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		searchResults := []*User{
			{
				ID:       uuid.New(),
				Username: "john_doe",
				FullName: stringPtr("John Doe"),
			},
			{
				ID:       uuid.New(),
				Username: "jane_smith",
				FullName: stringPtr("Jane Smith"),
			},
		}
		th.service.SetSearchResults(searchResults)

		req := httptest.NewRequest(http.MethodGet, "/users/search?q=john", nil)
		w := httptest.NewRecorder()

		th.handler.HandleSearchByFullName(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HandleSearchByFullName() status = %v, want %v", w.Code, http.StatusOK)
		}

		var responseUsers []*User
		if err := json.Unmarshal(w.Body.Bytes(), &responseUsers); err != nil {
			t.Errorf("HandleSearchByFullName() неверный JSON: %v", err)
		}

		if len(responseUsers) != 2 {
			t.Errorf("HandleSearchByFullName() вернул %d результатов, ожидалось 2", len(responseUsers))
		}
	})

	t.Run("Ошибка при отсутствии поискового запроса", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		req := httptest.NewRequest(http.MethodGet, "/users/search", nil)
		w := httptest.NewRecorder()

		th.handler.HandleSearchByFullName(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("HandleSearchByFullName() status = %v, want %v", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "search query is required\n" {
			t.Errorf("HandleSearchByFullName() body = %v, want %v", w.Body.String(), "search query is required\n")
		}
	})

	t.Run("Ошибка при поиске", func(t *testing.T) {
		th := createTestUserHandler()
		defer th.service.Reset()

		th.service.SetError("SearchByFullName", errors.New("search error"))

		req := httptest.NewRequest(http.MethodGet, "/users/search?q=test", nil)
		w := httptest.NewRecorder()

		th.handler.HandleSearchByFullName(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("HandleSearchByFullName() status = %v, want %v", w.Code, http.StatusInternalServerError)
		}
		if w.Body.String() != "cannot search users\n" {
			t.Errorf("HandleSearchByFullName() body = %v, want %v", w.Body.String(), "cannot search users\n")
		}
	})
}
