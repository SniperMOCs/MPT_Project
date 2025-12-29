package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAuthService для тестирования handler
type TestAuthService struct {
	mock.Mock
}

func (m *TestAuthService) Login(ctx context.Context, username string, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {

	testService := new(TestAuthService)
	handler := NewAuthHandler(testService)

	testService.On("Login", mock.Anything, "testuser", "password123").Return("test.jwt.token", nil)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test.jwt.token", response["access_token"])
	testService.AssertExpectations(t)
}

func TestAuthHandler_Login_WrongMethod(t *testing.T) {

	testService := new(TestAuthService)
	handler := NewAuthHandler(testService)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {

	testService := new(TestAuthService)
	handler := NewAuthHandler(testService)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestAuthHandler_Login_EmptyCredentials(t *testing.T) {

	testService := new(TestAuthService)
	handler := NewAuthHandler(testService)

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"empty username", "", "password"},
		{"empty password", "user", ""},
		{"both empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]string{
				"username": tc.username,
				"password": tc.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "username and password requiered")
		})
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {

	testService := new(TestAuthService)
	handler := NewAuthHandler(testService)

	testService.On("Login", mock.Anything, "invaliduser", "wrongpass").
		Return("", errors.New("invalid credentials"))

	reqBody := map[string]string{
		"username": "invaliduser",
		"password": "wrongpass",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
	testService.AssertExpectations(t)
}
