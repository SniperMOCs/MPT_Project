package user

import (
	"context"
	"errors"
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

func TestUserService_Create(t *testing.T) {
	t.Run("Успешное создание пользователя", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		user := &User{
			Username: "newuser",
			Password: "password123",
		}

		err := service.Create(context.Background(), user)

		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if user.ID == uuid.Nil {
			t.Error("Create() не установил ID")
		}
		if user.Password == "password123" {
			t.Error("Create() не захешировал пароль")
		}
		if user.CreatedAt.IsZero() {
			t.Error("Create() не установил CreatedAt")
		}
	})

	t.Run("Ошибка при пустом username", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		user := &User{
			Username: " ",
			Password: "password123",
		}

		err := service.Create(context.Background(), user)

		if err == nil {
			t.Error("Create() должен вернуть ошибку при пустом username")
		}
	})

	t.Run("Ошибка при занятом username", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		// Cоздание пользователя
		existingUser := &User{
			ID:       uuid.New(),
			Username: "user11211511",
			Password: hashPassword("password123"),
		}
		repo.AddUser(existingUser)

		// Создать пользователя с таким же username
		newUser := &User{
			Username: "user11211511",
			Password: "password456",
		}

		err := service.Create(context.Background(), newUser)

		if err == nil {
			t.Error("Create() должен вернуть ошибку при занятом username")
		}
	})
}

func TestUserService_GetByID(t *testing.T) {
	t.Run("Успешное получение пользователя по ID", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		foundUser, err := service.GetByID(context.Background(), testUser.ID)

		if err != nil {
			t.Errorf("GetByID() error = %v", err)
		}
		if foundUser.ID != testUser.ID {
			t.Errorf("GetByID() вернул пользователя с ID %v, ожидался %v", foundUser.ID, testUser.ID)
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		_, err := service.GetByID(context.Background(), uuid.New())

		if err == nil {
			t.Error("GetByID() должен вернуть ошибку если пользователь не найден")
		}
	})
}

func TestUserService_GetByUsername(t *testing.T) {
	t.Run("Успешное получение пользователя по username", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		foundUser, err := service.GetByUsername(context.Background(), testUser.Username)

		if err != nil {
			t.Errorf("GetByUsername() error = %v", err)
		}
		if foundUser.Username != testUser.Username {
			t.Errorf("GetByUsername() вернул пользователя с username %s, ожидался %s",
				foundUser.Username, testUser.Username)
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		_, err := service.GetByUsername(context.Background(), "nonexistent")

		if err == nil {
			t.Error("GetByUsername() должен вернуть ошибку если пользователь не найден")
		}
	})
}

func TestUserService_Update(t *testing.T) {
	t.Run("Успешное обновление всех полей", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		err := service.Update(context.Background(), testUser.ID,
			"newusername", "New Full Name", "New Description")

		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		// Проверяем что обновилось
		updatedUser, _ := repo.GetByID(context.Background(), testUser.ID)
		if updatedUser.Username != "newusername" {
			t.Error("Update() не обновил username")
		}
		if *updatedUser.FullName != "New Full Name" {
			t.Error("Update() не обновил FullName")
		}
		if *updatedUser.Description != "New Description" {
			t.Error("Update() не обновил Description")
		}
	})

	t.Run("Ошибка при занятом username", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		// Создаем двух пользователей
		user1 := &User{
			ID:       uuid.New(),
			Username: "user1",
			Password: hashPassword("pass1"),
		}
		user2 := &User{
			ID:       uuid.New(),
			Username: "user2",
			Password: hashPassword("pass2"),
		}
		repo.AddUser(user1)
		repo.AddUser(user2)

		// Пытаемся сменить username user2 на user1
		err := service.Update(context.Background(), user2.ID, "user1", "", "")

		if err == nil {
			t.Error("Update() должен вернуть ошибку при занятом username")
		}
	})

	t.Run("Обновление только одного поля", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		originalUsername := testUser.Username
		repo.AddUser(testUser)

		// Обновляем только fullname
		err := service.Update(context.Background(), testUser.ID,
			"", "Only FullName Updated", "")

		if err != nil {
			t.Errorf("Update() error = %v", err)
		}

		updatedUser, _ := repo.GetByID(context.Background(), testUser.ID)
		if updatedUser.Username != originalUsername {
			t.Error("Update() не должен менять username если он пустой")
		}
		if *updatedUser.FullName != "Only FullName Updated" {
			t.Error("Update() должен обновить FullName")
		}
	})
}

func TestUserService_UpdatePassword(t *testing.T) {
	t.Run("Успешная смена пароля", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		oldPassword := "oldpassword123"
		testUser.Password = hashPassword(oldPassword)
		repo.AddUser(testUser)

		newPassword := "newpasswor123132313123312312313d456"
		err := service.UpdatePassword(context.Background(), testUser.ID,
			oldPassword, newPassword)

		if err != nil {
			t.Errorf("UpdatePassword() error = %v", err)
		}

		// Проверяем что пароль изменился
		updatedUser, _ := repo.GetByID(context.Background(), testUser.ID)
		if service.CheckPassword(updatedUser, oldPassword) {
			t.Error("Старый пароль больше не должен работать")
		}
		if !service.CheckPassword(updatedUser, newPassword) {
			t.Error("Новый пароль должен работать")
		}
	})

	t.Run("Ошибка при неправильном старом пароле", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		testUser.Password = hashPassword("correctpassword")
		repo.AddUser(testUser)

		err := service.UpdatePassword(context.Background(), testUser.ID,
			"wrongpassword", "newpassword")

		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при неправильном старом пароле")
		}
	})

	t.Run("Ошибка при пустых паролях", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		testUser := createTestUser()
		repo.AddUser(testUser)

		// Пустой старый пароль
		err := service.UpdatePassword(context.Background(), testUser.ID,
			"", "newpassword")
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при пустом старом пароле")
		}

		// Пустой новый пароль
		err = service.UpdatePassword(context.Background(), testUser.ID,
			"oldpassword", "")
		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку при пустом новом пароле")
		}
	})

	t.Run("Ошибка если пользователь не найден", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		err := service.UpdatePassword(context.Background(), uuid.New(),
			"oldpass", "newpass")

		if err == nil {
			t.Error("UpdatePassword() должен вернуть ошибку если пользователь не найден")
		}
	})
}

func TestUserService_CheckPassword(t *testing.T) {
	repo := NewTestUserRepository()
	service := NewUserService(repo)

	t.Run("Правильный пароль", func(t *testing.T) {
		user := &User{
			Password: hashPassword("correctpassword"),
		}

		result := service.CheckPassword(user, "correctpassword")
		if !result {
			t.Error("CheckPassword() должен вернуть true для правильного пароля")
		}
	})

	t.Run("Неправильный пароль", func(t *testing.T) {
		user := &User{
			Password: hashPassword("correctpassword"),
		}

		result := service.CheckPassword(user, "wrongpassword")
		if result {
			t.Error("CheckPassword() должен вернуть false для неправильного пароля")
		}
	})

	t.Run("Пустой пароль", func(t *testing.T) {
		user := &User{
			Password: hashPassword("correctpassword"),
		}

		result := service.CheckPassword(user, "")
		if result {
			t.Error("CheckPassword() должен вернуть false для пустого пароля")
		}
	})
}

func TestUserService_SearchByFullName(t *testing.T) {
	t.Run("Успешный поиск пользователей", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		// Создаем тестовых пользователей для поиска
		users := []*User{
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

		// Устанавливаем результаты поиска в моке
		repo.SetSearchResults(users)

		results, err := service.SearchByFullName(context.Background(), "john")

		if err != nil {
			t.Errorf("SearchByFullName() error = %v", err)
		}
		if len(results) != 2 {
			t.Errorf("SearchByFullName() вернул %d результатов, ожидалось 2", len(results))
		}
	})

	t.Run("Ошибка поиска", func(t *testing.T) {
		repo := NewTestUserRepository()
		service := NewUserService(repo)

		repo.SetError("SearchByFullName", errors.New("search error"))

		_, err := service.SearchByFullName(context.Background(), "query")

		if err == nil {
			t.Error("SearchByFullName() должен вернуть ошибку")
		}
	})
}
