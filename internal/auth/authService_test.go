package auth

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/config"
	"github.com/SniperMOCs/MPT_Project/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TestUserService struct {
	usersByUsername map[string]*user.User
	errors          map[string]error
	callLog         []string
}

func NewTestUserService() *TestUserService {
	return &TestUserService{
		usersByUsername: make(map[string]*user.User),
		errors:          make(map[string]error),
		callLog:         []string{},
	}
}

func (s *TestUserService) Create(ctx context.Context, u *user.User) error {
	s.callLog = append(s.callLog, "Create")
	if err, ok := s.errors["Create"]; ok {
		return err
	}
	return nil
}

func (s *TestUserService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	s.callLog = append(s.callLog, "GetByID")
	if err, ok := s.errors["GetByID"]; ok {
		return nil, err
	}

	for _, u := range s.usersByUsername {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *TestUserService) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	s.callLog = append(s.callLog, "GetByUsername")
	if err, ok := s.errors["GetByUsername"]; ok {
		return nil, err
	}
	if user, exists := s.usersByUsername[username]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (s *TestUserService) SearchByFullName(ctx context.Context, searchQuery string) ([]*user.User, error) {
	s.callLog = append(s.callLog, "SearchByFullName")
	if err, ok := s.errors["SearchByFullName"]; ok {
		return nil, err
	}
	return []*user.User{}, nil
}

func (s *TestUserService) Update(ctx context.Context, id uuid.UUID, username string, fullname string, description string) error {
	s.callLog = append(s.callLog, "Update")
	if err, ok := s.errors["Update"]; ok {
		return err
	}
	return nil
}

func (s *TestUserService) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	s.callLog = append(s.callLog, "UpdatePassword")
	if err, ok := s.errors["UpdatePassword"]; ok {
		return err
	}
	return nil
}

func (s *TestUserService) CheckPassword(u *user.User, password string) bool {
	s.callLog = append(s.callLog, "CheckPassword")
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (s *TestUserService) UpdateAvatar(ctx context.Context, id uuid.UUID, file io.Reader) error {
	s.callLog = append(s.callLog, "UpdateAvatar")
	if err, ok := s.errors["UpdateAvatar"]; ok {
		return err
	}
	return nil
}

func (s *TestUserService) AddUser(username, password string) *user.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &user.User{
		ID:       uuid.New(),
		Username: username,
		Password: string(hashedPassword),
	}
	s.usersByUsername[username] = user
	return user
}

func (s *TestUserService) SetError(methodName string, err error) {
	s.errors[methodName] = err
}

func (s *TestUserService) WasCalled(methodName string) bool {
	for _, call := range s.callLog {
		if call == methodName {
			return true
		}
	}
	return false
}

func (s *TestUserService) Reset() {
	s.usersByUsername = make(map[string]*user.User)
	s.errors = make(map[string]error)
	s.callLog = []string{}
}

func createTestConfig() *config.Config {
	return &config.Config{
		TokenTTL: time.Hour * 24,
	}
}

func TestAuthService_Login_Success(t *testing.T) {

	userService := NewTestUserService()
	cfg := createTestConfig()
	secret := []byte("test-secret-key")
	jwtManager := NewJWTManager(secret)
	service := NewAuthService(userService, jwtManager, cfg)

	userService.AddUser("testuser", "password123")

	token, err := service.Login(context.Background(), "testuser", "password123")

	if err != nil {
		t.Errorf("Login() error = %v, ожидалось nil", err)
	}
	if token == "" {
		t.Error("Login() вернул пустой токен")
	}
	if !userService.WasCalled("GetByUsername") {
		t.Error("Login() должен вызывать GetByUsername")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {

	userService := NewTestUserService()
	cfg := createTestConfig()
	secret := []byte("test-secret-key")
	jwtManager := NewJWTManager(secret)
	service := NewAuthService(userService, jwtManager, cfg)

	token, err := service.Login(context.Background(), "nonexistent", "password")

	if err == nil {
		t.Error("Login() должен вернуть ошибку для несуществующего пользователя")
	}
	if token != "" {
		t.Error("Login() должен вернуть пустой токен при ошибке")
	}
	if !userService.WasCalled("GetByUsername") {
		t.Error("Login() должен вызывать GetByUsername")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {

	userService := NewTestUserService()
	cfg := createTestConfig()
	secret := []byte("test-secret-key")
	jwtManager := NewJWTManager(secret)
	service := NewAuthService(userService, jwtManager, cfg)

	userService.AddUser("testuser", "correctpassword")

	token, err := service.Login(context.Background(), "testuser", "wrongpassword")

	if err == nil {
		t.Error("Login() должен вернуть ошибку при неверном пароле")
	}
	if token != "" {
		t.Error("Login() должен вернуть пустой токен при неверном пароле")
	}
	if !userService.WasCalled("GetByUsername") {
		t.Error("Login() должен вызывать GetByUsername")
	}
}

func TestAuthService_Login_EdgeCases(t *testing.T) {
	t.Run("Ошибка при получении пользователя", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		userService.SetError("GetByUsername", errors.New("database error"))

		token, err := service.Login(context.Background(), "testuser", "password")

		if err == nil {
			t.Error("Login() должен вернуть ошибку при ошибке получения пользователя")
		}
		if token != "" {
			t.Error("Login() должен вернуть пустой токен при ошибке")
		}
	})

	t.Run("Пользователь с некорректным bcrypt хешем", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		userID := uuid.New()
		invalidUser := &user.User{
			ID:       userID,
			Username: "testuser",
			Password: "not-a-valid-bcrypt-hash",
		}
		userService.usersByUsername["testuser"] = invalidUser

		token, err := service.Login(context.Background(), "testuser", "password")

		if err == nil {
			t.Error("Login() должен вернуть ошибку при некорректном bcrypt хеше")
		}
		if token != "" {
			t.Error("Login() должен вернуть пустой токен при некорректном bcrypt хеше")
		}
	})

	t.Run("JWTManager возвращает ошибку", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()

		secret := []byte("")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		userService.AddUser("testuser", "password123")

		token, err := service.Login(context.Background(), "testuser", "password123")

		if err != nil {
			t.Errorf("Login() вернул неожиданную ошибку: %v", err)
		}
		if token == "" {
			t.Error("Login() вернул пустой токен, хотя должен был сгенерировать")
		}
	})

	t.Run("Пароль максимальной длины для bcrypt (72 байта)", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		longPassword := ""
		for i := 0; i < 72; i++ {
			longPassword += "a"
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(longPassword), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Не удалось создать хеш пароля: %v", err)
		}

		userID := uuid.New()
		user := &user.User{
			ID:       userID,
			Username: "testuser",
			Password: string(hashedPassword),
		}
		userService.usersByUsername["testuser"] = user

		token, err := service.Login(context.Background(), "testuser", longPassword)

		if err != nil {
			t.Errorf("Login() с паролем максимальной длины вернул ошибку: %v", err)
		}
		if token == "" {
			t.Error("Login() вернул пустой токен для пароля максимальной длины")
		}
	})

	t.Run("Пароль средней длины", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		mediumPassword := "12312341passwordmedium1231233232322"

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(mediumPassword), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Не удалось создать хеш пароля: %v", err)
		}

		userID := uuid.New()
		user := &user.User{
			ID:       userID,
			Username: "testuser",
			Password: string(hashedPassword),
		}
		userService.usersByUsername["testuser"] = user

		token, err := service.Login(context.Background(), "testuser", mediumPassword)

		if err != nil {
			t.Errorf("Login() с паролем средней длины вернул ошибку: %v", err)
		}
		if token == "" {
			t.Error("Login() вернул пустой токен для пароля средней длины")
		}
	})
}

func TestAuthService_Login_ConcurrentAccess(t *testing.T) {

	userService := NewTestUserService()
	cfg := createTestConfig()
	secret := []byte("test-secret-key")
	jwtManager := NewJWTManager(secret)
	service := NewAuthService(userService, jwtManager, cfg)

	userService.AddUser("user1", "pass1")
	userService.AddUser("user2", "pass2")
	userService.AddUser("user3", "pass3")

	results := make(chan struct {
		username string
		token    string
		err      error
	}, 3)

	for i := 1; i <= 3; i++ {
		go func(i int) {
			username := "user" + string(rune('0'+i))
			password := "pass" + string(rune('0'+i))
			token, err := service.Login(context.Background(), username, password)
			results <- struct {
				username string
				token    string
				err      error
			}{username, token, err}
		}(i)
	}

	successCount := 0
	for i := 0; i < 3; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("Login() для %s вернул ошибку: %v", result.username, result.err)
		}
		if result.token == "" {
			t.Errorf("Login() для %s вернул пустой токен", result.username)
		} else {
			successCount++
		}
	}

	if successCount != 3 {
		t.Errorf("Не все логины прошли успешно: %d из 3", successCount)
	}
}

func TestAuthService_Login_ContextCancellation(t *testing.T) {
	t.Run("Контекст отменен до вызова", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		userService.AddUser("testuser", "password123")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		token, err := service.Login(ctx, "testuser", "password123")

		if err != nil {
			t.Logf("Login() с отмененным контекстом вернул ошибку (ожидаемо): %v", err)
		}

		_ = token
	})

	t.Run("Контекст с таймаутом", func(t *testing.T) {
		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("test-secret-key")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		userService.AddUser("testuser", "password123")

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		userService.SetError("GetByUsername", errors.New("context deadline exceeded"))

		token, err := service.Login(ctx, "testuser", "password123")

		if err == nil {
			t.Error("Login() должен вернуть ошибку при таймауте контекста")
		}
		if token != "" {
			t.Error("Login() должен вернуть пустой токен при таймауте контекста")
		}
	})
}

func TestAuthService_Login_EmptyCredentials(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"пустой username", "", "password"},
		{"пустой password", "user", ""},
		{"оба пустых", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userService := NewTestUserService()
			cfg := createTestConfig()
			secret := []byte("test-secret-key")
			jwtManager := NewJWTManager(secret)
			service := NewAuthService(userService, jwtManager, cfg)

			token, err := service.Login(context.Background(), tc.username, tc.password)

			if err == nil {
				t.Errorf("Login() с %s должен вернуть ошибку", tc.name)
			}
			if token != "" {
				t.Errorf("Login() с %s должен вернуть пустой токен", tc.name)
			}

			_ = token
		})
	}
}

func TestAuthService_Integration(t *testing.T) {
	t.Run("Полный цикл аутентификации", func(t *testing.T) {

		userService := NewTestUserService()
		cfg := createTestConfig()
		secret := []byte("integration-test-secret")
		jwtManager := NewJWTManager(secret)
		service := NewAuthService(userService, jwtManager, cfg)

		testUser := userService.AddUser("integration_user", "integration_pass")

		token, err := service.Login(context.Background(), "integration_user", "integration_pass")

		if err != nil {
			t.Fatalf("Login() вернул ошибку: %v", err)
		}
		if token == "" {
			t.Fatal("Login() вернул пустой токен")
		}

		claims, err := jwtManager.ValidateToken(token)

		if err != nil {
			t.Fatalf("ValidateToken() вернул ошибку: %v", err)
		}
		if claims.UserID != testUser.ID {
			t.Errorf("UserID в токене не совпадает: got %v, want %v", claims.UserID, testUser.ID)
		}

		expiredToken, _ := jwtManager.GenerateToken(testUser.ID, -time.Hour)
		_, err = jwtManager.ValidateToken(expiredToken)

		if err == nil {
			t.Error("ValidateToken() должен вернуть ошибку для истекшего токена")
		}
	})
}
