package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTManager_GenerateToken_Success(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)
	userID := uuid.New()

	token, err := manager.GenerateToken(userID, time.Hour)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_ValidateToken_Success(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)
	userID := uuid.New()

	token, err := manager.GenerateToken(userID, time.Hour)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	assert.True(t, claims.IssuedAt.Time.Before(time.Now()))
}

func TestJWTManager_ValidateToken_Expired(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)
	userID := uuid.New()

	token, err := manager.GenerateToken(userID, -time.Hour)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {

	manager1 := NewJWTManager([]byte("secret-1"))
	manager2 := NewJWTManager([]byte("secret-2"))
	userID := uuid.New()

	token, err := manager1.GenerateToken(userID, time.Hour)
	assert.NoError(t, err)

	claims, err := manager2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestJWTManager_ValidateToken_InvalidFormat(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)

	testCases := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{"empty token", "", true},
		{"malformed token", "not.a.jwt.token", true},
		{"wrong parts count", "part1.part2", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			claims, err := manager.ValidateToken(tc.token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestJWTManager_GenerateToken_WithDifferentTTL(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)
	userID := uuid.New()

	ttls := []time.Duration{
		time.Minute * 5,
		time.Hour,
		time.Hour * 24,
		time.Hour * 24 * 7,
	}

	for _, ttl := range ttls {
		t.Run(ttl.String(), func(t *testing.T) {

			token, err := manager.GenerateToken(userID, ttl)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := manager.ValidateToken(token)
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.WithinDuration(t, time.Now().Add(ttl), claims.ExpiresAt.Time, time.Second*2)
		})
	}
}

func TestJWTManager_TokenStructure(t *testing.T) {

	secret := []byte("test-secret-key-123")
	manager := NewJWTManager(secret)
	userID := uuid.New()

	token, err := manager.GenerateToken(userID, time.Hour)

	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Токен должен иметь 3 части, разделенные точками. Получено %d частей", len(parts))
	}

	if parts[0] == "" {
		t.Error("Header should not be empty")
	}
	if parts[1] == "" {
		t.Error("Payload should not be empty")
	}
	if parts[2] == "" {
		t.Error("Signature should not be empty")
	}

	if !strings.HasPrefix(parts[0], "eyJ") {
		t.Error("Header должен быть в формате base64 JSON")
	}

	dotCount := strings.Count(token, ".")
	if dotCount != 2 {
		t.Errorf("Токен должен содержать 2 точки, содержит %d", dotCount)
	}

	if len(token) < 50 {
		t.Errorf("Токен слишком короткий: %d символов", len(token))
	}
}
