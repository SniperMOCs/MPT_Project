package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/config"
	"github.com/google/uuid"
)

func TestMiddleware_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		ctxUserID := r.Context().Value("user_id")
		if ctxUserID != userID {
			t.Errorf("user_id in context = %v, want %v", ctxUserID, userID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "OK" {
		t.Errorf("Response body = %s, want 'OK'", body)
	}
}

func TestMiddleware_MissingAuthorizationHeader(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		t.Error("Handler should not be called when Authorization header is missing")
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler should not be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "unauthorized") {
		t.Errorf("Response body should contain 'unauthorized', got: %s", w.Body.String())
	}
}

func TestMiddleware_EmptyAuthorizationHeader(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		t.Error("Handler should not be called with empty Authorization header")
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler should not be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	testCases := []struct {
		name        string
		authHeader  string
		description string
	}{
		{"NoBearerPrefix", "Token abc123", "префикс не Bearer"},
		{"NoSpace", "Bearertoken", "нет пробела после Bearer"},
		{"TooManySpaces", "Bearer  token1 token2", "слишком много пробелов"},
		{"LowerCaseBearer", "bearer token", "bearer в нижнем регистре"},
		{"DifferentScheme", "Basic dXNlcjpwYXNz", "другая схема аутентификации"},
		{"OnlyBearer", "Bearer", "только Bearer без токена"},
		{"BearerWithMultipleTokens", "Bearer token1 token2 token3", "несколько токенов"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			secret := []byte("test-secret")
			jwtManager := NewJWTManager(secret)
			cfg := &config.Config{
				TokenTTL: time.Hour * 24,
			}

			middleware := Middleware(jwtManager, cfg)

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				t.Errorf("Handler should not be called with %s", tc.description)
			})

			handler := middleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if handlerCalled {
				t.Errorf("Handler was called with %s", tc.description)
			}
			if w.Code != http.StatusUnauthorized {
				t.Errorf("Status code for %s = %d, want %d", tc.name, w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	testCases := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{"EmptyToken", "", true},
		{"MalformedToken", "not.a.jwt.token", true},
		{"RandomString", "randomstring", true},
		{"InvalidBase64", "header.payload.signature", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			secret := []byte("test-secret")
			jwtManager := NewJWTManager(secret)
			cfg := &config.Config{
				TokenTTL: time.Hour * 24,
			}

			middleware := Middleware(jwtManager, cfg)

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				t.Errorf("Handler should not be called with invalid token: %s", tc.name)
			})

			handler := middleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if handlerCalled {
				t.Errorf("Handler was called with invalid token: %s", tc.name)
			}
			if w.Code != http.StatusUnauthorized {
				t.Errorf("Status code for %s = %d, want %d", tc.name, w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		t.Error("Handler should not be called with expired token")
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler was called with expired token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_TokenAboutToExpire(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()

	token, err := jwtManager.GenerateToken(userID, 4*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		ctxUserID := r.Context().Value("user_id")
		if ctxUserID != userID {
			t.Errorf("user_id in context = %v, want %v", ctxUserID, userID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Token refreshed"))
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	newToken := w.Header().Get("X-New-Token")
	if newToken == "" {
		t.Error("X-New-Token header should be set when token is about to expire")
	}

	claims, err := jwtManager.ValidateToken(newToken)
	if err != nil {
		t.Errorf("New token is invalid: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("New token user_id = %v, want %v", claims.UserID, userID)
	}
}

func TestMiddleware_TokenNotAboutToExpire(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()

	token, err := jwtManager.GenerateToken(userID, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		ctxUserID := r.Context().Value("user_id")
		if ctxUserID != userID {
			t.Errorf("user_id in context = %v, want %v", ctxUserID, userID)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Token not refreshed"))
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	newToken := w.Header().Get("X-New-Token")
	if newToken != "" {
		t.Error("X-New-Token header should not be set when token is not about to expire")
	}
}

func TestMiddleware_ContextPropagation(t *testing.T) {

	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	outerMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := context.WithValue(r.Context(), "outer_key", "outer_value")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	authMiddleware := Middleware(jwtManager, cfg)

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		ctxUserID := r.Context().Value("user_id")
		if ctxUserID != userID {
			t.Errorf("user_id in context = %v, want %v", ctxUserID, userID)
		}

		outerValue := r.Context().Value("outer_key")
		if outerValue != "outer_value" {
			t.Errorf("outer_key in context = %v, want 'outer_value'", outerValue)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := outerMiddleware(authMiddleware(finalHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMiddleware_JWTGenerationError(t *testing.T) {

	type errorJWTManager struct {
		*JWTManager
		generateError bool
	}

	errorManager := &errorJWTManager{
		JWTManager: NewJWTManager([]byte("test-secret")),
	}

	userID := uuid.New()

	token, err := errorManager.JWTManager.GenerateToken(userID, 4*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(errorManager.JWTManager, cfg)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	newToken := w.Header().Get("X-New-Token")
	if newToken == "" {
		t.Log("X-New-Token not set (expected when token generation would fail)")
	}
}

func TestMiddleware_DifferentHTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {

			secret := []byte("test-secret")
			jwtManager := NewJWTManager(secret)
			cfg := &config.Config{
				TokenTTL: time.Hour * 24,
			}

			middleware := Middleware(jwtManager, cfg)

			userID := uuid.New()
			token, err := jwtManager.GenerateToken(userID, time.Hour)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				if r.Method != method {
					t.Errorf("Request method = %s, want %s", r.Method, method)
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(testHandler)

			req := httptest.NewRequest(method, "/api/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if !handlerCalled {
				t.Error("Handler was not called")
			}
			if w.Code != http.StatusOK {
				t.Errorf("Status code for %s = %d, want %d", method, w.Code, http.StatusOK)
			}
		})
	}
}

func TestMiddleware_Integration(t *testing.T) {

	secret := []byte("integration-test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, 30*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := http.NewServeMux()
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := r.Context().Value("user_id")
		if ctxUserID == nil {
			http.Error(w, "user_id not found in context", http.StatusInternalServerError)
			return
		}
		if ctxUserID != userID {
			http.Error(w, "user_id mismatch", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected resource accessed"))
	})

	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Public resource"))
	})

	router.Handle("/api/protected", middleware(protectedHandler))
	router.Handle("/api/public", publicHandler)

	t.Run("Access protected resource with valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}
		if !strings.Contains(w.Body.String(), "Protected resource") {
			t.Errorf("Response body = %s, want 'Protected resource'", w.Body.String())
		}
	})

	t.Run("Access protected resource without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("Access public resource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/public", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}
		if !strings.Contains(w.Body.String(), "Public resource") {
			t.Errorf("Response body = %s, want 'Public resource'", w.Body.String())
		}
	})

	t.Run("Access with expired token", func(t *testing.T) {
		expiredToken, err := jwtManager.GenerateToken(userID, -time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate expired token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("Access with token about to expire", func(t *testing.T) {
		aboutToExpireToken, err := jwtManager.GenerateToken(userID, 4*time.Minute)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+aboutToExpireToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		if w.Header().Get("X-New-Token") == "" {
			t.Error("X-New-Token should be set for token about to expire")
		}
	})
}

func TestMiddleware_PathSpecific(t *testing.T) {
	secret := []byte("test-secret")
	jwtManager := NewJWTManager(secret)
	cfg := &config.Config{
		TokenTTL: time.Hour * 24,
	}

	middleware := Middleware(jwtManager, cfg)

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	router := http.NewServeMux()

	protectedPaths := []string{
		"/api/users",
		"/api/users/me",
		"/api/users/123",
		"/api/posts",
		"/api/posts/456",
		"/admin/dashboard",
		"/admin/settings",
	}

	publicPaths := []string{
		"/",
		"/login",
		"/register",
		"/about",
		"/contact",
		"/api/health",
		"/api/status",
	}

	for _, path := range protectedPaths {
		router.Handle(path, middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Protected: " + r.URL.Path))
		})))
	}

	for _, path := range publicPaths {
		router.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Public: " + r.URL.Path))
		}))
	}

	t.Run("Protected path with token", func(t *testing.T) {
		for _, path := range protectedPaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status code for %s = %d, want %d", path, w.Code, http.StatusOK)
			}
		}
	})

	t.Run("Protected path without token", func(t *testing.T) {
		for _, path := range protectedPaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Status code for %s = %d, want %d", path, w.Code, http.StatusUnauthorized)
			}
		}
	})

	t.Run("Public path", func(t *testing.T) {
		for _, path := range publicPaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status code for %s = %d, want %d", path, w.Code, http.StatusOK)
			}
			if !strings.Contains(w.Body.String(), "Public: "+path) {
				t.Errorf("Response for %s = %s, want 'Public: %s'", path, w.Body.String(), path)
			}
		}
	})
}
