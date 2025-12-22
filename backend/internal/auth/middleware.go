package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/SniperMOCs/MPT_Project/internal/config"
)

func Middleware(jwt *JWTManager, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if time.Until(claims.ExpiresAt.Time) < 5*time.Minute {
				newToken, _ := jwt.GenerateToken(claims.UserID, cfg.TokenTTL)
				w.Header().Set("X-New-Token", newToken)
			}

			ctx := context.WithValue(
				r.Context(),
				"user_id",
				claims.UserID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
