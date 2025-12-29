package user

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Мок для тестов
type TestUserRepository struct {
	usersByID       map[uuid.UUID]*User
	usersByUsername map[string]*User
	createdUsers    []*User
	updatedUsers    []UpdatedUserData
	avatarUpdates   []AvatarUpdateData
	searchResults   []*User
	errors          map[string]error
	callLog         []string
}

type UpdatedUserData struct {
	ID          uuid.UUID
	Username    string
	FullName    string
	Description string
}

type AvatarUpdateData struct {
	ID        uuid.UUID
	AvatarURL string
}

func NewTestUserRepository() *TestUserRepository {
	return &TestUserRepository{
		usersByID:       make(map[uuid.UUID]*User),
		usersByUsername: make(map[string]*User),
		errors:          make(map[string]error),
		searchResults:   make([]*User, 0),
	}
}

func (r *TestUserRepository) Create(ctx context.Context, u *User) error {
	r.callLog = append(r.callLog, "Create")
	if err, ok := r.errors["Create"]; ok {
		return err
	}
	if _, exists := r.usersByUsername[u.Username]; exists {
		return errors.New("username already exists")
	}
	r.usersByID[u.ID] = u
	r.usersByUsername[u.Username] = u
	r.createdUsers = append(r.createdUsers, u)
	return nil
}

func (r *TestUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	r.callLog = append(r.callLog, "GetByID")
	if err, ok := r.errors["GetByID"]; ok {
		return nil, err
	}
	user, exists := r.usersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *TestUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	r.callLog = append(r.callLog, "GetByUsername")
	if err, ok := r.errors["GetByUsername"]; ok {
		return nil, err
	}
	user, exists := r.usersByUsername[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *TestUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.callLog = append(r.callLog, "ExistsByUsername")
	if err, ok := r.errors["ExistsByUsername"]; ok {
		return false, err
	}
	_, exists := r.usersByUsername[username]
	return exists, nil
}

func (r *TestUserRepository) Update(ctx context.Context, id uuid.UUID, username, fullname, description string) error {
	r.callLog = append(r.callLog, "Update")
	if err, ok := r.errors["Update"]; ok {
		return err
	}
	user, exists := r.usersByID[id]
	if !exists {
		return errors.New("user not found")
	}
	if username != "" && username != user.Username {
		if _, exists := r.usersByUsername[username]; exists {
			return errors.New("username already taken")
		}
		delete(r.usersByUsername, user.Username)
		user.Username = username
		r.usersByUsername[username] = user
	}
	if fullname != "" {
		fullNameCopy := fullname
		user.FullName = &fullNameCopy
	}
	if description != "" {
		descCopy := description
		user.Description = &descCopy
	}
	r.updatedUsers = append(r.updatedUsers, UpdatedUserData{
		ID:          id,
		Username:    username,
		FullName:    fullname,
		Description: description,
	})
	return nil
}

func (r *TestUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	r.callLog = append(r.callLog, "UpdatePassword")
	if err, ok := r.errors["UpdatePassword"]; ok {
		return err
	}
	user, exists := r.usersByID[userID]
	if !exists {
		return errors.New("user not found")
	}
	user.Password = passwordHash
	return nil
}

func (r *TestUserRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error {
	r.callLog = append(r.callLog, "UpdateAvatar")
	if err, ok := r.errors["UpdateAvatar"]; ok {
		return err
	}
	user, exists := r.usersByID[id]
	if !exists {
		return errors.New("user not found")
	}
	avatarURLCopy := avatarURL
	user.AvatarURL = &avatarURLCopy
	r.avatarUpdates = append(r.avatarUpdates, AvatarUpdateData{
		ID:        id,
		AvatarURL: avatarURL,
	})
	return nil
}

func (r *TestUserRepository) SearchByFullName(ctx context.Context, searchQuery string) ([]*User, error) {
	r.callLog = append(r.callLog, "SearchByFullName")
	if err, ok := r.errors["SearchByFullName"]; ok {
		return nil, err
	}
	return r.searchResults, nil
}

func (r *TestUserRepository) AddUser(user *User) {
	r.usersByID[user.ID] = user
	r.usersByUsername[user.Username] = user
}

func (r *TestUserRepository) SetSearchResults(users []*User) {
	r.searchResults = users
}

func (r *TestUserRepository) SetError(methodName string, err error) {
	r.errors[methodName] = err
}

func (r *TestUserRepository) Reset() {
	r.usersByID = make(map[uuid.UUID]*User)
	r.usersByUsername = make(map[string]*User)
	r.createdUsers = []*User{}
	r.updatedUsers = []UpdatedUserData{}
	r.avatarUpdates = []AvatarUpdateData{}
	r.searchResults = []*User{}
	r.errors = make(map[string]error)
	r.callLog = []string{}
}

func (r *TestUserRepository) WasCalled(methodName string) bool {
	for _, call := range r.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

type TestAvatarStorage struct {
	saveFunc func(ctx context.Context, userID uuid.UUID, file io.Reader) (string, error)
}

func (m *TestAvatarStorage) Save(ctx context.Context, userID uuid.UUID, file io.Reader) (string, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, userID, file)
	}
	return "/uploads/avatar.jpg", nil
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("ошибка при хешировании пароля: " + err.Error())
	}
	return string(hash)
}

func stringPtr(s string) *string {
	return &s
}

func createTestUser() *User {
	fullName := "Test User"
	description := "Test Description"
	avatarURL := "/uploads/avatar.jpg"

	return &User{
		ID:          uuid.New(),
		Username:    "testuser",
		FullName:    &fullName,
		Password:    hashPassword("password123"),
		AvatarURL:   &avatarURL,
		Description: &description,
		CreatedAt:   time.Now(),
	}
}

func createTestUserService(repo *TestUserRepository) *userService {
	return &userService{
		repo:          repo,
		avatarStorage: &TestAvatarStorage{},
	}
}

func TestUserService_Create_EdgeCases(t *testing.T) {
	t.Run("Ошибка при пустом пароле", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		user := &User{
			Username: "testuser",
			Password: " ",
		}

		err := service.Create(context.Background(), user)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при пустом пароле")
		}
	})

	t.Run("Ошибка при ошибке проверки существования username", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("ExistsByUsername", errors.New("database error"))
		service := createTestUserService(repo)

		user := &User{
			Username: "testuser",
			Password: "password123",
		}

		err := service.Create(context.Background(), user)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке проверки username")
		}
	})

	t.Run("Ошибка при ошибке создания в репозитории", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("Create", errors.New("database error"))
		service := createTestUserService(repo)

		user := &User{
			Username: "testuser",
			Password: "password123",
		}

		err := service.Create(context.Background(), user)
		if err == nil {
			t.Error("Create() должен вернуть ошибку при ошибке создания")
		}
	})
}

func TestUserService_Update_EdgeCases(t *testing.T) {
	t.Run("Обновление без изменений", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		err := service.Update(context.Background(), testUser.ID, "", "", "")
		if err != nil {
			t.Errorf("Update() с пустыми полями не должен возвращать ошибку: %v", err)
		}
	})

	t.Run("Ошибка при ошибке проверки существования username", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)
		repo.SetError("ExistsByUsername", errors.New("database error"))

		err := service.Update(context.Background(), testUser.ID, "newusername", "", "")
		if err == nil {
			t.Error("Update() должен вернуть ошибку при ошибке проверки username")
		}
	})

	t.Run("Ошибка при ошибке обновления в репозитории", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)
		repo.SetError("Update", errors.New("database error"))

		err := service.Update(context.Background(), testUser.ID, "newusername", "", "")
		if err == nil {
			t.Error("Update() должен вернуть ошибку при ошибке обновления")
		}
	})

	t.Run("Обновление только описания", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		newDescription := "Updated Description"
		err := service.Update(context.Background(), testUser.ID, "", "", newDescription)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		updatedUser, _ := repo.GetByID(context.Background(), testUser.ID)
		if *updatedUser.Description != newDescription {
			t.Errorf("Update() не обновил Description: got %v, want %v", *updatedUser.Description, newDescription)
		}
	})
}

func TestUserService_UpdatePassword_EdgeCases(t *testing.T) {
	t.Run("Ошибка при ошибке получения пользователя", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		repo.SetError("GetByID", errors.New("database error"))

		err := service.UpdatePassword(context.Background(), uuid.New(), "oldpass", "newpass")
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при ошибке получения пользователя")
		}
	})

	t.Run("Ошибка при ошибке хеширования пароля", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		testUser.Password = hashPassword("oldpassword")
		repo.AddUser(testUser)

		longPassword := strings.Repeat("a", 100)
		err := service.UpdatePassword(context.Background(), testUser.ID, "oldpassword", longPassword)
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при ошибке хеширования")
		}
	})

	t.Run("Ошибка при ошибке обновления пароля в репозитории", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		testUser := createTestUser()
		testUser.Password = hashPassword("oldpassword")
		repo.AddUser(testUser)
		repo.SetError("UpdatePassword", errors.New("database error"))

		err := service.UpdatePassword(context.Background(), testUser.ID, "oldpassword", "newpassword")
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при ошибке обновления в репозитории")
		}
	})
}

func TestUserService_UpdateAvatar(t *testing.T) {
	t.Run("Успешное обновление аватара", func(t *testing.T) {
		repo := NewTestUserRepository()
		storage := &TestAvatarStorage{
			saveFunc: func(ctx context.Context, userID uuid.UUID, file io.Reader) (string, error) {
				return "/uploads/avatar_new.jpg", nil
			},
		}
		service := &userService{
			repo:          repo,
			avatarStorage: storage,
		}

		testUser := createTestUser()
		repo.AddUser(testUser)

		file := strings.NewReader("fake image data")
		err := service.UpdateAvatar(context.Background(), testUser.ID, file)
		if err != nil {
			t.Errorf("UpdateAvatar() error = %v", err)
		}

		if !repo.WasCalled("UpdateAvatar") {
			t.Error("UpdateAvatar() должен вызывать репозиторий")
		}
	})

	t.Run("Ошибка при сохранении файла в хранилище", func(t *testing.T) {
		repo := NewTestUserRepository()
		storage := &TestAvatarStorage{
			saveFunc: func(ctx context.Context, userID uuid.UUID, file io.Reader) (string, error) {
				return "", errors.New("storage error")
			},
		}
		service := &userService{
			repo:          repo,
			avatarStorage: storage,
		}

		testUser := createTestUser()
		repo.AddUser(testUser)

		file := strings.NewReader("fake image data")
		err := service.UpdateAvatar(context.Background(), testUser.ID, file)
		if err == nil {
			t.Error("UpdateAvatar() должен вернуть ошибку при ошибке хранилища")
		}
	})

	t.Run("Ошибка при обновлении в репозитории", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("UpdateAvatar", errors.New("database error"))
		storage := &TestAvatarStorage{
			saveFunc: func(ctx context.Context, userID uuid.UUID, file io.Reader) (string, error) {
				return "/uploads/avatar_new.jpg", nil
			},
		}
		service := &userService{
			repo:          repo,
			avatarStorage: storage,
		}

		testUser := createTestUser()
		repo.AddUser(testUser)

		file := strings.NewReader("fake image data")
		err := service.UpdateAvatar(context.Background(), testUser.ID, file)
		if err == nil {
			t.Error("UpdateAvatar() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestUserService_SearchByFullName_EdgeCases(t *testing.T) {
	t.Run("Поиск с пустым запросом", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := createTestUserService(repo)

		users := []*User{
			{
				ID:       uuid.New(),
				Username: "user1",
				FullName: stringPtr("John Doe"),
			},
			{
				ID:       uuid.New(),
				Username: "user2",
				FullName: stringPtr("Jane Smith"),
			},
		}
		repo.SetSearchResults(users)

		results, err := service.SearchByFullName(context.Background(), "")
		if err != nil {
			t.Errorf("SearchByFullName() с пустым запросом не должен возвращать ошибку: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("SearchByFullName() вернул %d результатов, ожидалось 2", len(results))
		}
	})

	t.Run("Ошибка при поиске в репозитории", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("SearchByFullName", errors.New("database error"))
		service := createTestUserService(repo)

		_, err := service.SearchByFullName(context.Background(), "search")
		if err == nil {
			t.Error("SearchByFullName() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestUserService_GetByUsername_EdgeCases(t *testing.T) {
	t.Run("Ошибка при ошибке репозитория", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("GetByUsername", errors.New("database error"))
		service := createTestUserService(repo)

		_, err := service.GetByUsername(context.Background(), "testuser")
		if err == nil {
			t.Error("GetByUsername() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestUserService_GetByID_EdgeCases(t *testing.T) {
	t.Run("Ошибка при ошибке репозитория", func(t *testing.T) {
		repo := NewTestUserRepository()
		repo.SetError("GetByID", errors.New("database error"))
		service := createTestUserService(repo)

		_, err := service.GetByID(context.Background(), uuid.New())
		if err == nil {
			t.Error("GetByID() должен вернуть ошибку при ошибке репозитория")
		}
	})
}

func TestUserService_Update_UsernameExists(t *testing.T) {
	repo := NewTestUserRepository()
	service := createTestUserService(repo)

	user1 := &User{
		ID:       uuid.New(),
		Username: "user1",
		Password: hashPassword("password1"),
	}
	user2 := &User{
		ID:       uuid.New(),
		Username: "user2",
		Password: hashPassword("password2"),
	}
	repo.AddUser(user1)
	repo.AddUser(user2)

	t.Run("Обновление username на существующий", func(t *testing.T) {
		err := service.Update(context.Background(), user1.ID, "user2", "", "")
		if err == nil {
			t.Error("Update() должен вернуть ошибку при занятом username")
		}
		if err.Error() != "username taken" {
			t.Errorf("Update() ошибка = %v, ожидалось 'username taken'", err)
		}
	})
}
